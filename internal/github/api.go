package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Repo holds metadata about a GitHub repository.
type Repo struct {
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	URL             string    `json:"url"`
	Stars           int       `json:"stargazerCount"`
	IsArchived      bool      `json:"isArchived"`
	PrimaryLanguage string    `json:"primaryLanguage"`
	PushedAt        time.Time `json:"pushedAt"`
	GoVersion       string    // populated separately from go.mod
}

type repoJSON struct {
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	URL             string          `json:"url"`
	Stars           int             `json:"stargazerCount"`
	IsArchived      bool            `json:"isArchived"`
	PrimaryLanguage json.RawMessage `json:"primaryLanguage"`
	PushedAt        time.Time       `json:"pushedAt"`
}

// ListGHRepos returns all gh-* repos for the given owner.
func ListGHRepos(owner string) ([]Repo, error) {
	out, err := exec.Command("gh", "repo", "list", owner,
		"--json", "name,description,url,stargazerCount,isArchived,primaryLanguage,pushedAt",
		"--limit", "200",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("listing repos: %w", err)
	}

	var raw []repoJSON
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parsing repos: %w", err)
	}

	var repos []Repo
	for _, r := range raw {
		if !strings.HasPrefix(r.Name, "gh-") {
			continue
		}
		lang := parseLang(r.PrimaryLanguage)
		repos = append(repos, Repo{
			Name:            r.Name,
			Description:     r.Description,
			URL:             r.URL,
			Stars:           r.Stars,
			IsArchived:      r.IsArchived,
			PrimaryLanguage: lang,
			PushedAt:        r.PushedAt,
		})
	}
	return repos, nil
}

func parseLang(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	// Try as object with "name" field
	var obj struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil && obj.Name != "" {
		return obj.Name
	}
	// Try as plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return ""
}

// FetchGoVersion retrieves the Go version from go.mod for a repo.
func FetchGoVersion(owner, repo string) string {
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/contents/go.mod", owner, repo),
		"-q", ".content",
	).Output()
	if err != nil {
		return ""
	}

	decoded, err := decodeBase64(strings.TrimSpace(string(out)))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(decoded, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			return strings.TrimPrefix(line, "go ")
		}
	}
	return ""
}

// FetchGoVersions populates GoVersion for all repos concurrently.
func FetchGoVersions(owner string, repos []Repo) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // limit concurrency

	for i := range repos {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			repos[idx].GoVersion = FetchGoVersion(owner, repos[idx].Name)
		}(i)
	}
	wg.Wait()
}

// FileExists checks if a file exists in a remote repo.
func FileExists(owner, repo, path string) bool {
	err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path),
		"-q", ".name",
	).Run()
	return err == nil
}

// FetchFileContent retrieves base64-decoded file content from a repo.
func FetchFileContent(owner, repo, path string) (string, error) {
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path),
		"-q", ".content",
	).Output()
	if err != nil {
		return "", fmt.Errorf("fetching %s/%s/%s: %w", owner, repo, path, err)
	}
	return decodeBase64(strings.TrimSpace(string(out)))
}

// GetDefaultBranch returns the default branch name for a repo.
func GetDefaultBranch(owner, repo string) (string, error) {
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s", owner, repo),
		"-q", ".default_branch",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CreateBranchAndPR creates a branch, commits a file change, and opens a PR.
func CreateBranchAndPR(owner, repo, branch, targetPath, content, commitMsg, prTitle, prBody string) error {
	// Get default branch SHA
	defaultBranch, err := GetDefaultBranch(owner, repo)
	if err != nil {
		return fmt.Errorf("getting default branch: %w", err)
	}

	refOut, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/git/ref/heads/%s", owner, repo, defaultBranch),
		"-q", ".object.sha",
	).Output()
	if err != nil {
		return fmt.Errorf("getting ref SHA: %w", err)
	}
	sha := strings.TrimSpace(string(refOut))

	// Create branch
	createRef := fmt.Sprintf(`{"ref":"refs/heads/%s","sha":"%s"}`, branch, sha)
	if err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/git/refs", owner, repo),
		"-X", "POST",
		"--input", "-",
	).Run(); err != nil {
		// Use a different approach — pipe the JSON
		cmd := exec.Command("gh", "api",
			fmt.Sprintf("repos/%s/%s/git/refs", owner, repo),
			"-X", "POST",
			"-f", fmt.Sprintf("ref=refs/heads/%s", branch),
			"-f", fmt.Sprintf("sha=%s", sha),
		)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("creating branch: %w (payload: %s)", err, createRef)
		}
	}

	// Create/update file
	encoded := encodeBase64(content)
	// Check if file exists to get its SHA
	existingSHA := ""
	existOut, err := exec.Command("gh", "api",
		fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, targetPath),
		"-q", ".sha",
		"--jq", ".sha",
	).Output()
	if err == nil {
		existingSHA = strings.TrimSpace(string(existOut))
	}

	args := []string{"api",
		fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, targetPath),
		"-X", "PUT",
		"-f", fmt.Sprintf("message=%s", commitMsg),
		"-f", fmt.Sprintf("content=%s", encoded),
		"-f", fmt.Sprintf("branch=%s", branch),
	}
	if existingSHA != "" {
		args = append(args, "-f", fmt.Sprintf("sha=%s", existingSHA))
	}
	if err := exec.Command("gh", args...).Run(); err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	// Create PR
	if err := exec.Command("gh", "pr", "create",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--head", branch,
		"--base", defaultBranch,
		"--title", prTitle,
		"--body", prBody,
	).Run(); err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}

	return nil
}
