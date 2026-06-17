// Package parser turns the JSON emitted by `terraform show -json` into a flat,
// normalized list of resources the rule engine can reason about without
// caring about Terraform's nested module layout.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Resource is a normalized view of one Terraform resource.
type Resource struct {
	Address  string                 // full address, e.g. module.network.aws_vpc.main
	Module   string                 // module path ("" for the root module)
	Type     string                 // resource type, e.g. aws_s3_bucket
	Name     string                 // logical name
	Provider string                 // normalized: aws | azure | gcp | <other>
	Mode     string                 // managed | data
	Actions  []string               // plan actions: create | update | delete | no-op
	Values   map[string]interface{} // attribute values (best-effort, may be partial)
}

// Plan is the normalized result of parsing a Terraform plan.
type Plan struct {
	Resources []Resource
}

// rawPlan is the subset of the Terraform plan JSON schema we consume.
type rawPlan struct {
	PlannedValues struct {
		RootModule rawModule `json:"root_module"`
	} `json:"planned_values"`
	ResourceChanges []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

type rawModule struct {
	Address      string        `json:"address"`
	Resources    []rawResource `json:"resources"`
	ChildModules []rawModule   `json:"child_modules"`
}

type rawResource struct {
	Address      string                 `json:"address"`
	Mode         string                 `json:"mode"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	ProviderName string                 `json:"provider_name"`
	Values       map[string]interface{} `json:"values"`
}

// ParseFile reads and parses a plan JSON file.
func ParseFile(path string) (*Plan, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plan file: %w", err)
	}
	return Parse(b)
}

// Parse parses plan JSON bytes.
func Parse(b []byte) (*Plan, error) {
	var rp rawPlan
	if err := json.Unmarshal(b, &rp); err != nil {
		return nil, fmt.Errorf("parse plan JSON (is this `terraform show -json` output?): %w", err)
	}

	// Index actions by address so we can attach them while walking modules.
	actions := make(map[string][]string, len(rp.ResourceChanges))
	for _, rc := range rp.ResourceChanges {
		actions[rc.Address] = rc.Change.Actions
	}

	p := &Plan{}
	walkModule(rp.PlannedValues.RootModule, actions, p)
	return p, nil
}

// walkModule recursively flattens a module tree into Plan.Resources.
func walkModule(m rawModule, actions map[string][]string, p *Plan) {
	for _, r := range m.Resources {
		p.Resources = append(p.Resources, Resource{
			Address:  r.Address,
			Module:   moduleOf(r.Address),
			Type:     r.Type,
			Name:     r.Name,
			Provider: providerOf(r.Type, r.ProviderName),
			Mode:     r.Mode,
			Actions:  actions[r.Address],
			Values:   r.Values,
		})
	}
	for _, c := range m.ChildModules {
		walkModule(c, actions, p)
	}
}

// moduleOf extracts the module path from a resource address. A root resource
// "aws_vpc.main" returns ""; a nested "module.network.aws_vpc.main" returns
// "module.network".
func moduleOf(addr string) string {
	if !strings.HasPrefix(addr, "module.") {
		return ""
	}
	parts := strings.Split(addr, ".")
	// Last two segments are always <type>.<name>; everything before is the path.
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-2], ".")
	}
	return ""
}

// providerOf normalizes a resource to a cloud family, first by the well-known
// type prefix, then falling back to the provider_name field.
func providerOf(typ, providerName string) string {
	switch {
	case strings.HasPrefix(typ, "aws_"):
		return "aws"
	case strings.HasPrefix(typ, "azurerm_"), strings.HasPrefix(typ, "azuread_"), strings.HasPrefix(typ, "azapi_"):
		return "azure"
	case strings.HasPrefix(typ, "google_"):
		return "gcp"
	}
	if providerName != "" {
		seg := providerName
		if i := strings.LastIndex(seg, "/"); i >= 0 {
			seg = seg[i+1:]
		}
		switch seg {
		case "aws":
			return "aws"
		case "azurerm", "azuread":
			return "azure"
		case "google", "google-beta":
			return "gcp"
		default:
			return seg
		}
	}
	return "other"
}
