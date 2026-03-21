package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// FleetPR represents an open pull request created by gh-fleet sync.
type FleetPR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Repo   string `json:"-"`
	Branch string `json:"headRefName"`
}

// ListFleetPRs returns all open PRs whose head branch starts with "fleet/sync-".
func ListFleetPRs(owner, repo string) ([]FleetPR, error) {
	out, err := exec.Command("gh", "pr", "list",
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--head", "fleet/sync-",
		"--state", "open",
		"--json", "number,title,url,headRefName",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("listing PRs for %s/%s: %w", owner, repo, err)
	}

	var prs []FleetPR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parsing PRs for %s/%s: %w", owner, repo, err)
	}
	for i := range prs {
		prs[i].Repo = repo
	}
	return prs, nil
}

// FetchFleetPRs fetches fleet PRs from multiple repos concurrently.
func FetchFleetPRs(owner string, repos []string) []FleetPR {
	var mu sync.Mutex
	var all []FleetPR
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, repo := range repos {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			prs, err := ListFleetPRs(owner, r)
			if err != nil {
				return // skip repos with errors
			}
			mu.Lock()
			all = append(all, prs...)
			mu.Unlock()
		}(repo)
	}
	wg.Wait()
	return all
}

// MergePR squash-merges a pull request using admin privileges.
func MergePR(owner, repo string, number int, admin bool) error {
	args := []string{"pr", "merge",
		strconv.Itoa(number),
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
		"--squash",
	}
	if admin {
		args = append(args, "--admin")
	}
	out, err := exec.Command("gh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("merging %s/%s#%d: %w\n%s", owner, repo, number, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ClosePR closes a pull request with an optional comment.
func ClosePR(owner, repo string, number int, comment string) error {
	args := []string{"pr", "close",
		strconv.Itoa(number),
		"--repo", fmt.Sprintf("%s/%s", owner, repo),
	}
	if comment != "" {
		args = append(args, "--comment", comment)
	}
	out, err := exec.Command("gh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("closing %s/%s#%d: %w\n%s", owner, repo, number, err, strings.TrimSpace(string(out)))
	}
	return nil
}
