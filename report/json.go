package report

import (
	"encoding/json"
	"io"

	"github.com/mikejrossiter/terrasentry/config"
	"github.com/mikejrossiter/terrasentry/rules"
	"github.com/mikejrossiter/terrasentry/rules/portability"
)

// Output is the top-level JSON document emitted by `scan --json`.
type Output struct {
	Summary     Summary             `json:"summary"`
	Portability *portability.Report `json:"portability"`
	Findings    []rules.Finding     `json:"findings"`
}

// Summary is the at-a-glance result a CI gate or dashboard reads first.
type Summary struct {
	RepoPortability float64 `json:"repo_portability"`
	MinScore        float64 `json:"min_score"`
	TotalResources  int     `json:"total_resources"`
	Findings        int     `json:"findings"`
	HighSeverity    int     `json:"high_severity"`
	Pass            bool    `json:"pass"`
}

// JSON writes the full result as indented JSON.
func JSON(w io.Writer, rep *portability.Report, findings []rules.Finding, cfg *config.Config) error {
	counts := severityCounts(findings)
	out := Output{
		Summary: Summary{
			RepoPortability: rep.RepoScore,
			MinScore:        cfg.Portability.MinScore,
			TotalResources:  len(rep.Resources),
			Findings:        len(findings),
			HighSeverity:    counts[rules.SeverityHigh],
			Pass:            verdict(rep, findings, cfg),
		},
		Portability: rep,
		Findings:    findings,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
