package grdl

import (
	"fmt"
	"os"
	"strings"

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
		for j, l := range r.Laws {
			norm, ok := NormalizeLaw(l)
			if !ok {
				return nil, fmt.Errorf("rule %s: unknown governance law %q", r.ID, l)
			}
			r.Laws[j] = norm
		}
		if r.TimeoutMs == 0 {
			r.TimeoutMs = 50
		}
	}
	if err := validateParameters(&rs); err != nil {
		return nil, err
	}
	return &rs, nil
}

// validateParameters checks governance parameter ranges and makes sure every
// params.* reference in a rule resolves to a configured parameter, so a
// missing parameter is a load-time error rather than a silently skipped rule.
func validateParameters(rs *Ruleset) error {
	gp := rs.GovernanceParameters
	if gp != nil && gp.Proportionality != nil {
		a := gp.Proportionality.AIWeightBound
		if a <= 0 || a >= 1 {
			return fmt.Errorf("governance_parameters.proportionality.ai_weight_bound must be in (0,1), got %v", a)
		}
	}
	if gp != nil && gp.Reversibility != nil && gp.Reversibility.ChallengeWindowS <= 0 {
		return fmt.Errorf("governance_parameters.reversibility.challenge_window_s must be > 0")
	}
	params := gp.ToMap()
	for _, r := range rs.Rules {
		for _, ref := range conditionRefs(r.Condition) {
			if !strings.HasPrefix(ref, "params.") {
				continue
			}
			if lookupPath(params, strings.TrimPrefix(ref, "params.")) == nil {
				return fmt.Errorf("rule %s references %s, which is not set in governance_parameters", r.ID, ref)
			}
		}
	}
	return nil
}

func conditionRefs(c *Condition) []string {
	if c == nil {
		return nil
	}
	refs := []string{c.Field, c.ValueSource, c.ScaleSource}
	for i := range c.Conditions {
		refs = append(refs, conditionRefs(&c.Conditions[i])...)
	}
	return refs
}

func lookupPath(m map[string]interface{}, path string) interface{} {
	var cur interface{} = m
	for _, key := range strings.Split(path, ".") {
		cm, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur = cm[key]
	}
	return cur
}

// ToMap returns the parameters as a plain map for the CFAIS runtime config.
func (gp *GovernanceParameters) ToMap() map[string]interface{} {
	out := map[string]interface{}{}
	if gp == nil {
		return out
	}
	if p := gp.Proportionality; p != nil {
		out["proportionality"] = map[string]interface{}{"ai_weight_bound": p.AIWeightBound, "scope": p.Scope}
	}
	if r := gp.Reversibility; r != nil {
		out["reversibility"] = map[string]interface{}{"challenge_window_s": r.ChallengeWindowS, "scope": r.Scope}
	}
	return out
}
