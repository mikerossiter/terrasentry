// Package config loads the terrasentry.yaml policy file that defines an
// organisation's thresholds and conventions. A missing file is not an error:
// the tool falls back to defaults so it runs with zero setup.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the full policy document.
type Config struct {
	Version     int               `yaml:"version"`
	Portability PortabilityConfig `yaml:"portability"`
	Budget      BudgetConfig      `yaml:"budget"`
	Conventions ConventionsConfig `yaml:"conventions"`
	Security    SecurityConfig    `yaml:"security"`
}

// PortabilityConfig tunes the cloud lock-in scorer (the wedge feature).
type PortabilityConfig struct {
	MinScore    float64 `yaml:"min_score"`    // repo aggregate gate (CI fails below this)
	WarnBelow   float64 `yaml:"warn_below"`   // per-resource warning threshold
	DatasetPath string  `yaml:"dataset_path"` // optional override of the seed dataset
}

// BudgetConfig is roadmap: cost-heuristic thresholds.
type BudgetConfig struct {
	MonthlyLimitUSD float64 `yaml:"monthly_limit_usd"`
}

// ConventionsConfig is roadmap: naming, tagging, and approved modules.
type ConventionsConfig struct {
	RequiredTags          []string `yaml:"required_tags"`
	NamingPattern         string   `yaml:"naming_pattern"`
	ApprovedModuleSources []string `yaml:"approved_module_sources"`
}

// SecurityConfig is roadmap: misconfiguration checks.
type SecurityConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Default returns the built-in policy used when no config file is present.
func Default() *Config {
	return &Config{
		Version:     1,
		Portability: PortabilityConfig{MinScore: 0.6, WarnBelow: 0.4},
		Security:    SecurityConfig{Enabled: true},
	}
}

// Load reads a config file, returning defaults if the path is empty or missing.
// Values present in the file override the defaults; absent values keep them.
func Load(path string) (*Config, error) {
	if path == "" {
		return Default(), nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}
