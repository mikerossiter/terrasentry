package rules

import (
	"github.com/mikejrossiter/terrasentry/config"
	"github.com/mikejrossiter/terrasentry/parser"
)

// The rules below are roadmap placeholders. They satisfy the Rule interface so
// the engine wiring is exercised today and adding real logic later is a single
// in-file change — no plumbing. Each returns no findings for now.

// SecurityRule will flag misconfigurations (public buckets, open security
// groups, unencrypted volumes) and may wrap Trivy/Checkov when present.
type SecurityRule struct{}

func (SecurityRule) ID() string                                   { return "security.placeholder" }
func (SecurityRule) Category() string                             { return "security" }
func (SecurityRule) Check(*parser.Plan, *config.Config) []Finding { return nil }

// CostRule will flag obviously expensive resources using rough heuristics
// (no live pricing API required for the MVP).
type CostRule struct{}

func (CostRule) ID() string                                   { return "cost.placeholder" }
func (CostRule) Category() string                             { return "cost" }
func (CostRule) Check(*parser.Plan, *config.Config) []Finding { return nil }

// ConventionsRule will enforce naming patterns, required tags, and approved
// module sources defined in terrasentry.yaml.
type ConventionsRule struct{}

func (ConventionsRule) ID() string                                   { return "conventions.placeholder" }
func (ConventionsRule) Category() string                             { return "conventions" }
func (ConventionsRule) Check(*parser.Plan, *config.Config) []Finding { return nil }
