// Package portability implements terrasentry's wedge feature: scoring how
// cloud-locked a Terraform configuration is. Every resource type carries a
// score from 0.0 (fully proprietary, hard to move off this cloud) to 1.0
// (open/standard, easy to move). We aggregate these into a per-module and
// per-repo "lock-in score" — a metric no existing tool produces.
package portability

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/mikejrossiter/terrasentry/config"
	"github.com/mikejrossiter/terrasentry/data"
	"github.com/mikejrossiter/terrasentry/parser"
	"github.com/mikejrossiter/terrasentry/rules"
)

// Dataset maps a Terraform resource type to a portability score. It is loaded
// from the embedded seed or a user-supplied YAML file with the same shape.
type Dataset struct {
	Defaults struct {
		Unknown float64 `yaml:"unknown"`
	} `yaml:"defaults"`
	Resources map[string]Entry `yaml:"resources"`
}

// Entry is the scoring record for one resource type.
type Entry struct {
	Score  float64 `yaml:"score"`
	Reason string  `yaml:"reason"`
}

// LoadDataset reads the dataset from path, or the embedded seed if path is "".
func LoadDataset(path string) (*Dataset, error) {
	raw := data.Portability
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read portability dataset: %w", err)
		}
		raw = b
	}
	var ds Dataset
	if err := yaml.Unmarshal(raw, &ds); err != nil {
		return nil, fmt.Errorf("parse portability dataset: %w", err)
	}
	if ds.Defaults.Unknown == 0 {
		ds.Defaults.Unknown = 0.5 // neutral default for unscored types
	}
	return &ds, nil
}

// ScoreOf returns the score, reason, and whether the type was found in the
// dataset (unknown types fall back to the neutral default).
func (ds *Dataset) ScoreOf(typ string) (score float64, reason string, known bool) {
	if e, ok := ds.Resources[typ]; ok {
		return e.Score, e.Reason, true
	}
	return ds.Defaults.Unknown, "no portability data for this resource type yet", false
}

// ResourceScore is the portability result for one resource.
type ResourceScore struct {
	Address  string  `json:"address"`
	Type     string  `json:"type"`
	Provider string  `json:"provider"`
	Score    float64 `json:"score"`
	Reason   string  `json:"reason"`
	Known    bool    `json:"known"`
}

// ModuleScore is the aggregate for one module.
type ModuleScore struct {
	Module string  `json:"module"`
	Score  float64 `json:"score"`
	Count  int     `json:"count"`
}

// Report is the full portability picture for a plan.
type Report struct {
	RepoScore float64         `json:"repo_score"`
	Resources []ResourceScore `json:"resources"` // sorted worst (most locked) first
	Modules   []ModuleScore   `json:"modules"`
	Unknown   []string        `json:"unknown_types,omitempty"`
}

type modAcc struct {
	sum float64
	n   int
}

// Score computes the portability report. Managed resources are scored; data
// sources are skipped (they describe, not provision, infrastructure).
func Score(p *parser.Plan, ds *Dataset) *Report {
	rep := &Report{}
	mods := map[string]*modAcc{}
	var sum float64

	for _, r := range p.Resources {
		if r.Mode == "data" {
			continue
		}
		s, reason, known := ds.ScoreOf(r.Type)
		rep.Resources = append(rep.Resources, ResourceScore{
			Address:  r.Address,
			Type:     r.Type,
			Provider: r.Provider,
			Score:    s,
			Reason:   reason,
			Known:    known,
		})
		sum += s

		mod := r.Module
		if mod == "" {
			mod = "root"
		}
		a := mods[mod]
		if a == nil {
			a = &modAcc{}
			mods[mod] = a
		}
		a.sum += s
		a.n++

		if !known {
			rep.Unknown = append(rep.Unknown, r.Type)
		}
	}

	if n := len(rep.Resources); n > 0 {
		rep.RepoScore = sum / float64(n)
	}
	for name, a := range mods {
		rep.Modules = append(rep.Modules, ModuleScore{Module: name, Score: a.sum / float64(a.n), Count: a.n})
	}

	sort.Slice(rep.Modules, func(i, j int) bool { return rep.Modules[i].Module < rep.Modules[j].Module })
	sort.SliceStable(rep.Resources, func(i, j int) bool { return rep.Resources[i].Score < rep.Resources[j].Score })
	rep.Unknown = dedup(rep.Unknown)
	return rep
}

// Findings converts a report into engine findings using the config thresholds:
// one per resource below warn_below, plus an aggregate finding if the repo
// score falls under min_score (the value a CI gate would act on).
func (rep *Report) Findings(cfg *config.Config) []rules.Finding {
	var fs []rules.Finding
	for _, rs := range rep.Resources {
		if rs.Score < cfg.Portability.WarnBelow {
			fs = append(fs, rules.Finding{
				RuleID:   "portability.low_score",
				Category: "portability",
				Severity: severityFor(rs.Score),
				Resource: rs.Address,
				Message:  fmt.Sprintf("%s is highly cloud-locked (score %.2f): %s", rs.Type, rs.Score, rs.Reason),
			})
		}
	}
	// Only gate on the repo score when there is something to score. An empty or
	// data-only plan has no managed resources, so a 0.00 score is an artifact of
	// "nothing here", not real lock-in — flagging it would be a false alarm.
	if len(rep.Resources) > 0 && rep.RepoScore < cfg.Portability.MinScore {
		fs = append(fs, rules.Finding{
			RuleID:   "portability.repo_threshold",
			Category: "portability",
			Severity: rules.SeverityHigh,
			Message: fmt.Sprintf("repo lock-in score %.2f is below the required minimum %.2f",
				rep.RepoScore, cfg.Portability.MinScore),
		})
	}
	return fs
}

// Rule adapts the scorer to the engine's Rule interface while retaining the
// structured Report for the renderer (Check stores it; Report returns it).
type Rule struct {
	ds   *Dataset
	last *Report
}

// NewRule builds a portability rule over the given dataset.
func NewRule(ds *Dataset) *Rule { return &Rule{ds: ds} }

func (r *Rule) ID() string       { return "portability" }
func (r *Rule) Category() string { return "portability" }

// Check scores the plan, caches the report, and returns threshold findings.
func (r *Rule) Check(p *parser.Plan, cfg *config.Config) []rules.Finding {
	r.last = Score(p, r.ds)
	return r.last.Findings(cfg)
}

// Report returns the report from the most recent Check.
func (r *Rule) Report() *Report { return r.last }

func severityFor(score float64) string {
	switch {
	case score < 0.15:
		return rules.SeverityHigh
	case score < 0.30:
		return rules.SeverityMedium
	default:
		return rules.SeverityLow
	}
}

func dedup(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
