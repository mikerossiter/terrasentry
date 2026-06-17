// Package report renders scan results, either as a human-readable terminal
// report or as JSON for machines and CI.
package report

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mikejrossiter/terrasentry/config"
	"github.com/mikejrossiter/terrasentry/rules"
	"github.com/mikejrossiter/terrasentry/rules/portability"
)

const rule = "──────────────────────────────────────────────────────────"

// Terminal writes a human-readable report.
func Terminal(w io.Writer, rep *portability.Report, findings []rules.Finding, cfg *config.Config) {
	fmt.Fprintln(w, "Terrasentry — IaC Governance Scan")
	fmt.Fprintln(w, rule)

	// --- Cloud portability (the wedge) ---
	fmt.Fprintln(w, "\nCLOUD PORTABILITY  (lock-in score: 1.00 = portable, 0.00 = locked-in)")
	if len(rep.Resources) == 0 {
		fmt.Fprintln(w, "  Repo score : n/a   (no managed resources in this plan to score)")
	} else {
		fmt.Fprintf(w, "  Repo score : %s %.2f / 1.00   (policy min %.2f)  %s\n",
			bar(rep.RepoScore), rep.RepoScore, cfg.Portability.MinScore, passLabel(rep.RepoScore >= cfg.Portability.MinScore))
	}

	if len(rep.Modules) > 1 {
		fmt.Fprintln(w, "\n  Per module:")
		for _, m := range rep.Modules {
			fmt.Fprintf(w, "    %-26s %s %.2f  (%d resources)\n", m.Module, bar(m.Score), m.Score, m.Count)
		}
	}

	if len(rep.Resources) > 0 {
		fmt.Fprintln(w, "\n  Most locked-in resources:")
		const limit = 8
		for i, rs := range rep.Resources {
			if i >= limit {
				fmt.Fprintf(w, "    … and %d more\n", len(rep.Resources)-limit)
				break
			}
			fmt.Fprintf(w, "    %.2f  %-44s %s\n", rs.Score, truncate(rs.Address, 44), rs.Provider)
		}
	}

	if len(rep.Unknown) > 0 {
		fmt.Fprintf(w, "\n  Note: %d resource type(s) had no portability data (scored neutral): %s\n",
			len(rep.Unknown), strings.Join(rep.Unknown, ", "))
	}

	// --- Findings, grouped by category ---
	fmt.Fprintln(w, "\n"+rule)
	fmt.Fprintln(w, "\nPOLICY FINDINGS")
	if len(findings) == 0 {
		fmt.Fprintln(w, "  No violations. ✔")
	} else {
		for _, cat := range categoriesOf(findings) {
			fmt.Fprintf(w, "\n  [%s]\n", strings.ToUpper(cat))
			for _, f := range findings {
				if f.Category != cat {
					continue
				}
				loc := f.Resource
				if loc == "" {
					loc = "(repo-wide)"
				}
				fmt.Fprintf(w, "    %-7s %s\n", "["+f.Severity+"]", loc)
				fmt.Fprintf(w, "            %s\n", f.Message)
				if f.Suggestion != "" {
					fmt.Fprintf(w, "            fix: %s\n", f.Suggestion)
				}
			}
		}
	}

	// --- Verdict ---
	pass := verdict(rep, findings, cfg)
	fmt.Fprintln(w, "\n"+rule)
	counts := severityCounts(findings)
	fmt.Fprintf(w, "Summary: %d resources scored · portability %.2f · %d findings (high %d, medium %d, low %d)\n",
		len(rep.Resources), rep.RepoScore, len(findings), counts[rules.SeverityHigh], counts[rules.SeverityMedium], counts[rules.SeverityLow])
	if pass {
		fmt.Fprintln(w, "Result : PASS ✔")
	} else {
		fmt.Fprintln(w, "Result : FAIL ✘   (a CI gate via `terrasentry ci` would block this — roadmap)")
	}
}

// bar renders a 10-cell progress bar for a 0..1 score.
func bar(score float64) string {
	const cells = 10
	filled := int(score*cells + 0.5)
	if filled < 0 {
		filled = 0
	}
	if filled > cells {
		filled = cells
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", cells-filled) + "]"
}

func passLabel(ok bool) string {
	if ok {
		return "PASS"
	}
	return "BELOW MIN"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func categoriesOf(fs []rules.Finding) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, f := range fs {
		if _, ok := seen[f.Category]; ok {
			continue
		}
		seen[f.Category] = struct{}{}
		out = append(out, f.Category)
	}
	sort.Strings(out)
	return out
}

func severityCounts(fs []rules.Finding) map[string]int {
	m := map[string]int{}
	for _, f := range fs {
		m[f.Severity]++
	}
	return m
}

// verdict is shared by both renderers: fail if the repo score is below the
// minimum or any high-severity finding exists.
func verdict(rep *portability.Report, findings []rules.Finding, cfg *config.Config) bool {
	// An empty or data-only plan has nothing to score; the repo gate does not
	// apply (its 0.00 is "nothing here", not lock-in).
	if len(rep.Resources) > 0 && rep.RepoScore < cfg.Portability.MinScore {
		return false
	}
	for _, f := range findings {
		if f.Severity == rules.SeverityHigh {
			return false
		}
	}
	return true
}
