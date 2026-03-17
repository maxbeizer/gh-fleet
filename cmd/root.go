package cmd

import (
	"flag"
	"fmt"
	"os"
)

func Execute(version string) error {
	fs := flag.NewFlagSet("gh-fleet", flag.ContinueOnError)
	showVersion := fs.Bool("v", false, "show version")
	showHelp := fs.Bool("h", false, "show help")

	// Parse only known flags before subcommand
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *showVersion {
		fmt.Printf("gh-fleet v%s\n", version)
		return nil
	}

	if *showHelp || fs.NArg() == 0 {
		printUsage(version)
		return nil
	}

	subcmd := fs.Arg(0)
	args := fs.Args()[1:]

	switch subcmd {
	case "catalog":
		return runCatalog(args)
	case "drift":
		return runDrift(args)
	case "sync":
		return runSync(args)
	case "status":
		return runStatus(args)
	case "settings":
		return runSettings(args)
	default:
		return fmt.Errorf("unknown command: %s\nRun 'gh fleet -h' for usage", subcmd)
	}
}

func printUsage(version string) {
	fmt.Printf(`gh-fleet v%s — Command center for your gh CLI extensions

Usage:
  gh fleet <command> [flags]

Commands:
  catalog    Regenerate README with extension catalog
  drift      Detect configuration drift across repos
  settings   Enforce repo settings across the fleet
  sync       Push canonical files to out-of-sync repos
  status     Quick health matrix across all extension repos

Flags:
  -h         Show help
  -v         Show version
`, version)
}
