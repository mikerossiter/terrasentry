package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/mikejrossiter/terrasentry/config"
	"github.com/mikejrossiter/terrasentry/parser"
	"github.com/mikejrossiter/terrasentry/report"
	"github.com/mikejrossiter/terrasentry/rules"
	"github.com/mikejrossiter/terrasentry/rules/portability"
)

// runScan implements `terrasentry scan`: parse a plan, run the rule engine,
// and print a human or JSON report.
func runScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	planPath := fs.String("plan", "", "path to `terraform show -json` output (required)")
	cfgPath := fs.String("config", "terrasentry.yaml", "path to the terrasentry.yaml policy file")
	asJSON := fs.Bool("json", false, "emit machine-readable JSON instead of a terminal report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *planPath == "" {
		fmt.Fprintln(os.Stderr, "error: --plan is required")
		fmt.Fprintln(os.Stderr, "hint: terraform show -json tfplan > plan.json")
		return 2
	}

	// Config is optional: a missing file falls back to sane defaults so the
	// tool runs with zero setup.
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	plan, err := parser.ParseFile(*planPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// Dataset comes from the embedded seed unless the config overrides the path.
	ds, err := portability.LoadDataset(cfg.Portability.DatasetPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	// The engine is the pluggable core: portability is the live rule today;
	// security, cost, and conventions are registered stubs so the architecture
	// is exercised end-to-end ahead of their implementations.
	portRule := portability.NewRule(ds)
	engine := rules.New(
		portRule,
		rules.SecurityRule{},
		rules.CostRule{},
		rules.ConventionsRule{},
	)
	findings := engine.Run(plan, cfg)
	rep := portRule.Report()

	if *asJSON {
		if err := report.JSON(os.Stdout, rep, findings, cfg); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		return 0
	}
	report.Terminal(os.Stdout, rep, findings, cfg)
	return 0
}
