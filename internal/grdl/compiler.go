package grdl

import (
	"gopkg.in/yaml.v3"
)

const (
	CFAISHost     = "127.0.0.1"
	CFAISPort     = 9700
	AuditPath     = "/sandbox/cfais-audit"
	RulesPath     = "/sandbox/cfais-rules"
)

// CompilationResult holds both OpenShell policy and CFAIS runtime config.
type CompilationResult struct {
	OpenShellPolicy map[string]interface{} `json:"openshell_policy"`
	CFAISConfig     map[string]interface{} `json:"cfais_config"`
	Stats           CompileStats           `json:"stats"`
	Warnings        []string               `json:"warnings"`
}

type CompileStats struct {
	TotalRules   int      `json:"total_rules"`
	StaticRules  int      `json:"static_rules"`
	DynamicRules int      `json:"dynamic_rules"`
	RuntimeRules int      `json:"runtime_rules"`
	LawsCovered  []string `json:"laws_covered"`
}

// ToOpenShellYAML serializes the OpenShell policy to YAML bytes.
func (cr *CompilationResult) ToOpenShellYAML() ([]byte, error) {
	return yaml.Marshal(cr.OpenShellPolicy)
}

// ToCFAISYAML serializes the CFAIS runtime config to YAML bytes.
func (cr *CompilationResult) ToCFAISYAML() ([]byte, error) {
	return yaml.Marshal(cr.CFAISConfig)
}

// Compile transforms a Ruleset into OpenShell policies and CFAIS runtime config.
func Compile(rs *Ruleset) *CompilationResult {
	var static, dynamic, runtime []Rule
	var warnings []string

	for _, r := range rs.Rules {
		switch r.Target {
		case TargetOpenShellStatic:
			static = append(static, r)
		case TargetOpenShellDynamic:
			dynamic = append(dynamic, r)
		case TargetRuntime:
			runtime = append(runtime, r)
		case TargetHybrid:
			dynamic = append(dynamic, r)
			runtime = append(runtime, r)
		}
	}

	// Collect unique laws
	lawSet := make(map[string]bool)
	for _, r := range rs.Rules {
		for _, l := range r.Laws {
			lawSet[string(l)] = true
		}
	}
	var laws []string
	for l := range lawSet {
		laws = append(laws, l)
	}

	osPolicy := buildOpenShellPolicy(dynamic, &warnings)
	cfaisConfig := buildCFAISConfig(runtime, rs, laws)

	return &CompilationResult{
		OpenShellPolicy: osPolicy,
		CFAISConfig:     cfaisConfig,
		Stats: CompileStats{
			TotalRules:   len(rs.Rules),
			StaticRules:  len(static),
			DynamicRules: len(dynamic),
			RuntimeRules: len(runtime),
			LawsCovered:  laws,
		},
		Warnings: warnings,
	}
}

func buildOpenShellPolicy(dynamic []Rule, warnings *[]string) map[string]interface{} {
	networkPolicies := map[string]interface{}{
		"cfais_sidecar": map[string]interface{}{
			"name": "cfais-governance-engine",
			"endpoints": []map[string]interface{}{
				{
					"host":        CFAISHost,
					"port":        CFAISPort,
					"protocol":    "rest",
					"tls":         "passthrough",
					"enforcement": "enforce",
					"access":      "full",
				},
			},
			"binaries": []map[string]interface{}{
				{"path": "/sandbox/**"},
			},
		},
	}

	return map[string]interface{}{
		"version": 1,
		"filesystem_policy": map[string]interface{}{
			"include_workdir": true,
			"read_only":  []string{"/usr", "/lib", "/proc", "/dev/urandom", "/etc", RulesPath},
			"read_write": []string{"/sandbox", "/tmp", "/dev/null", AuditPath},
		},
		"process": map[string]interface{}{
			"run_as_user":  "sandbox",
			"run_as_group": "sandbox",
		},
		"landlock": map[string]interface{}{
			"compatibility": "best_effort",
		},
		"network_policies": networkPolicies,
	}
}

func buildCFAISConfig(runtime []Rule, rs *Ruleset, activeLaws []string) map[string]interface{} {
	lawsActive := make(map[string]bool)
	for _, l := range activeLaws {
		lawsActive[l] = true
	}

	govLaws := map[string]interface{}{}
	allLaws := []GovernanceLaw{
		LawPrimacy, LawTransparency, LawAccountability,
		LawFairness, LawSafety, LawPrivacy, LawGracefulDegrad,
	}
	for _, l := range allLaws {
		govLaws[string(l)] = map[string]interface{}{
			"active": lawsActive[string(l)],
		}
	}

	rules := make([]interface{}, 0, len(runtime))
	for _, r := range runtime {
		rules = append(rules, serializeRule(r))
	}

	return map[string]interface{}{
		"cfais_engine": map[string]interface{}{
			"version":         "0.1.0",
			"ruleset_id":      rs.ID,
			"ruleset_version": rs.Version,
			"framework":       rs.Framework,
			"sidecar":         map[string]interface{}{"host": CFAISHost, "port": CFAISPort},
			"graceful_degradation": map[string]interface{}{
				"policy":     rs.GracefulDegradation,
				"timeout_ms": 100,
				"circuit_breaker": map[string]interface{}{
					"failure_threshold": 5,
					"recovery_s":        30,
				},
			},
			"audit": map[string]interface{}{"log_path": AuditPath, "enabled": true},
		},
		"governance_laws": govLaws,
		"rules":           rules,
	}
}

func serializeRule(r Rule) map[string]interface{} {
	laws := make([]string, len(r.Laws))
	for i, l := range r.Laws {
		laws[i] = string(l)
	}

	entry := map[string]interface{}{
		"id":          r.ID,
		"name":        r.Name,
		"description": r.Description,
		"scope":       r.Scope,
		"severity":    string(r.Severity),
		"enforcement": string(r.Enforcement),
		"laws":        laws,
		"timeout_ms":  r.TimeoutMs,
	}

	if r.Condition != nil {
		entry["condition"] = serializeCondition(*r.Condition)
	}
	if r.Remedy != nil {
		entry["remedy"] = map[string]interface{}{
			"action":      r.Remedy.Action,
			"message":     r.Remedy.Message,
			"alternative": r.Remedy.Alternative,
			"escalation":  r.Remedy.Escalation,
			"audit_log":   r.Remedy.AuditLog,
		}
	}
	if len(r.PolicyRefs) > 0 {
		refs := make([]map[string]interface{}, len(r.PolicyRefs))
		for i, p := range r.PolicyRefs {
			refs[i] = map[string]interface{}{
				"framework":    p.Framework,
				"provision_id": p.ProvisionID,
				"description":  p.Description,
			}
		}
		entry["policy_refs"] = refs
	}
	if r.FormalProof != "" {
		entry["formal_proof"] = r.FormalProof
	}
	if r.FallbackRule != "" {
		entry["fallback_rule"] = r.FallbackRule
	}
	return entry
}

func serializeCondition(c Condition) map[string]interface{} {
	if c.IsComposite() {
		subs := make([]interface{}, len(c.Conditions))
		for i, sub := range c.Conditions {
			subs[i] = serializeCondition(sub)
		}
		return map[string]interface{}{
			"logic":      c.Logic,
			"conditions": subs,
		}
	}
	m := map[string]interface{}{
		"field":    c.Field,
		"operator": c.Operator,
		"value":    c.Value,
	}
	if c.ValueSource != "" {
		m["value_source"] = c.ValueSource
	}
	return m
}
