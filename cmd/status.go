package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/maxbeizer/gh-fleet/internal/fleet"
	gh "github.com/maxbeizer/gh-fleet/internal/github"
)

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := fleet.LoadConfig(*configDir)
	if err != nil {
		return err
	}

	repos, err := discoverRepos(cfg)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Fetching status for %d repos...\n", len(repos))
	gh.FetchGoVersions(cfg.Owner, repos)

	// Header
	fmt.Printf("%-28s %-10s %-8s %-6s %-12s\n",
		boldStyle.Render("Extension"),
		boldStyle.Render("Language"),
		boldStyle.Render("Go"),
		boldStyle.Render("Stars"),
		boldStyle.Render("Last Push"),
	)
	fmt.Println(strings.Repeat("─", 70))

	for _, r := range repos {
		goVer := r.GoVersion
		if goVer == "" {
			goVer = "—"
		}
		lang := r.PrimaryLanguage
		if lang == "" {
			lang = "—"
		}

		ago := timeSince(r.PushedAt)

		agoStyle := okStyle
		if time.Since(r.PushedAt) > 90*24*time.Hour {
			agoStyle = warnStyle
		}
		if time.Since(r.PushedAt) > 365*24*time.Hour {
			agoStyle = errStyle
		}

		fmt.Printf("%-28s %-10s %-8s ⭐%-4d %s\n",
			r.Name, lang, goVer, r.Stars, agoStyle.Render(ago))
	}

	fmt.Printf("\n%s extensions total\n", boldStyle.Render(fmt.Sprintf("%d", len(repos))))
	return nil
}

func timeSince(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return "just now"
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
	}
}
