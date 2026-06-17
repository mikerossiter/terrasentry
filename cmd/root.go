// Package cmd implements terrasentry's command-line interface. It uses the
// standard library flag package (no external CLI framework) to keep the binary
// small and dependency-light.
package cmd

import (
	"fmt"
	"os"
)

// version is the build version. Kept simple for the MVP; wire to ldflags later.
const version = "0.1.0"

// Main dispatches a subcommand and returns the process exit code.
func Main(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "scan":
		return runScan(args[1:])
	case "version", "--version", "-v":
		fmt.Println("terrasentry", version)
		return 0
	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Print(`terrasentry ` + version + ` — affordable IaC governance & AI policy guardrail

Usage:
  terrasentry scan --plan <plan.json> [--config terrasentry.yaml] [--json]
  terrasentry version

Commands:
  scan      Score a Terraform plan for cloud portability and policy violations.
  version   Print the version.

Roadmap commands (not yet implemented): ci, mcp.

Generate a plan to scan with:
  terraform plan -out tfplan
  terraform show -json tfplan > plan.json
  terrasentry scan --plan plan.json
`)
}
