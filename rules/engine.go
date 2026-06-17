// Package rules defines the pluggable policy engine. Each policy dimension
// (portability, security, cost, conventions) is a Rule; the engine simply runs
// every registered rule and collects their findings. New rules are added by
// implementing the interface and registering them — no engine changes needed.
package rules

import (
	"github.com/mikejrossiter/terrasentry/config"
	"github.com/mikejrossiter/terrasentry/parser"
)

// Severity levels, ordered low to high.
const (
	SeverityInfo   = "info"
	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"
)

// Finding is a single policy result. Suggestion is an optional HCL hint/diff
// that addresses the "auto-fix in the PR, don't just flag" gap.
type Finding struct {
	RuleID     string `json:"rule_id"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Resource   string `json:"resource,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Rule is one policy check across the plan.
type Rule interface {
	ID() string
	Category() string
	Check(p *parser.Plan, cfg *config.Config) []Finding
}

// Engine runs a set of rules over a plan.
type Engine struct {
	rules []Rule
}

// New builds an engine from the given rules.
func New(rs ...Rule) *Engine {
	return &Engine{rules: rs}
}

// Register adds a rule to the engine.
func (e *Engine) Register(r Rule) {
	e.rules = append(e.rules, r)
}

// Run executes every rule and returns the combined findings.
func (e *Engine) Run(p *parser.Plan, cfg *config.Config) []Finding {
	var out []Finding
	for _, r := range e.rules {
		out = append(out, r.Check(p, cfg)...)
	}
	return out
}
