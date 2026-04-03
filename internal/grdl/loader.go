package grdl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadRuleset(path string) (*Ruleset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading ruleset: %w", err)
	}
	return ParseRuleset(data)
}

func ParseRuleset(data []byte) (*Ruleset, error) {
	var rs Ruleset
	if err := yaml.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("parsing ruleset: %w", err)
	}
	if rs.Version == "" {
		rs.Version = "1.0.0"
	}
	if rs.Framework == "" {
		rs.Framework = "custom"
	}
	if rs.GracefulDegradation == "" {
		rs.GracefulDegradation = "allow_with_audit"
	}
	for i := range rs.Rules {
		r := &rs.Rules[i]
		if r.Version == "" {
			r.Version = "1.0.0"
		}
		if r.Scope == "" {
			r.Scope = "*"
		}
		if r.Severity == "" {
			r.Severity = SevHigh
		}
		if r.Enforcement == "" {
			r.Enforcement = EnfEnforce
		}
		if r.Target == "" {
			r.Target = TargetRuntime
		}
		// Normalize legacy OpenShell-specific targets
		r.Target = NormalizeTarget(r.Target)
		if r.TimeoutMs == 0 {
			r.TimeoutMs = 50
		}
	}
	return &rs, nil
}
