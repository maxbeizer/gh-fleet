package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	configDir := fs.String("config", ".", "directory containing fleet.toml")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(*configDir)
	if err != nil {
		return err
	}

	fmt.Println(boldStyle.Render("Fleet Doctor"))
	fmt.Println()

	failed := false
	pass := func(msg string) {
		fmt.Printf("  %s %s\n", okStyle.Render("✅"), msg)
	}
	fail := func(msg string) {
		fmt.Printf("  %s %s\n", errStyle.Render("❌"), msg)
		failed = true
	}

	// owner is non-empty
	if cfg.Owner != "" {
		pass("owner is set")
	} else {
		fail("owner is empty")
	}

	// catalog.output is set
	if cfg.Catalog.Output != "" {
		pass("catalog.output is set")
	} else {
		fail("catalog.output is not set")
	}

	// catalog.header file exists if configured
	if cfg.Catalog.Header != "" {
		headerPath := filepath.Join(cfg.Dir, cfg.Catalog.Header)
		if _, err := os.Stat(headerPath); err == nil {
			pass(fmt.Sprintf("catalog.header file exists (%s)", cfg.Catalog.Header))
		} else {
			fail(fmt.Sprintf("catalog.header file missing: %s", cfg.Catalog.Header))
		}
	}

	// All canonical files exist on disk
	for _, sf := range cfg.Sync.Files {
		canonPath := filepath.Join(cfg.Dir, sf.Canon)
		if _, err := os.Stat(canonPath); err == nil {
			pass(fmt.Sprintf("canon file exists: %s", sf.Canon))
		} else {
			fail(fmt.Sprintf("canon file missing: %s", sf.Canon))
		}
	}

	// Template files contain the "extension-template" placeholder
	for _, sf := range cfg.Sync.Files {
		if !sf.Template {
			continue
		}
		canonPath := filepath.Join(cfg.Dir, sf.Canon)
		data, err := os.ReadFile(canonPath)
		if err != nil {
			continue // already reported as missing above
		}
		if strings.Contains(string(data), "extension-template") {
			pass(fmt.Sprintf("template has extension-template placeholder: %s", sf.Canon))
		} else {
			fail(fmt.Sprintf("template missing extension-template placeholder: %s", sf.Canon))
		}
	}

	// Collect all template file contents for variable checks
	varRefPattern := regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)
	referencedVars := map[string]bool{}
	for _, sf := range cfg.Sync.Files {
		if !sf.Template {
			continue
		}
		canonPath := filepath.Join(cfg.Dir, sf.Canon)
		data, err := os.ReadFile(canonPath)
		if err != nil {
			continue
		}
		for _, match := range varRefPattern.FindAllStringSubmatch(string(data), -1) {
			referencedVars[match[1]] = true
		}
	}

	// All template_vars keys are used in at least one template file
	for key := range cfg.Sync.TemplateVars {
		if referencedVars[key] {
			pass(fmt.Sprintf("template_var %s is used", key))
		} else {
			fail(fmt.Sprintf("template_var %s is defined but never referenced in any template", key))
		}
	}

	// No template files reference variables not defined in template_vars
	for varName := range referencedVars {
		if _, ok := cfg.Sync.TemplateVars[varName]; ok {
			pass(fmt.Sprintf("template ref ${%s} is defined", varName))
		} else {
			fail(fmt.Sprintf("template ref ${%s} is not defined in template_vars", varName))
		}
	}

	fmt.Println()
	if failed {
		fmt.Println(errStyle.Render("Some checks failed."))
		return fmt.Errorf("doctor found issues")
	}
	fmt.Println(okStyle.Render("All checks passed."))
	return nil
}
