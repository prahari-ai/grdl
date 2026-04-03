package grdl

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

const (
	CFAISHost = "127.0.0.1"
	CFAISPort = 9700
	AuditPath = "/sandbox/cfais-audit"
	RulesPath = "/sandbox/cfais-rules"
)

// Backend formats GRDL output for a specific runtime.
type Backend interface {
	Name() string
	CompileInfrastructure(infra []Rule, network []Rule, rs *Ruleset) (map[string]interface{}, []string)
	SidecarConfig() map[string]interface{}
}

type CompilationResult struct {
	BackendName     string                 `json:"backend"`
	InfraConfig     map[string]interface{} `json:"infrastructure_config"`
	CFAISConfig     map[string]interface{} `json:"cfais_config"`
	Stats           CompileStats           `json:"stats"`
	Warnings        []string               `json:"warnings"`
}

type CompileStats struct {
	TotalRules   int      `json:"total_rules"`
	InfraRules   int      `json:"infrastructure_rules"`
	NetworkRules int      `json:"network_rules"`
	RuntimeRules int      `json:"runtime_rules"`
	LawsCovered  []string `json:"laws_covered"`
}

func (cr *CompilationResult) ToInfraYAML() ([]byte, error) {
	return yaml.Marshal(cr.InfraConfig)
}

func (cr *CompilationResult) ToCFAISYAML() ([]byte, error) {
	return yaml.Marshal(cr.CFAISConfig)
}

// Compile transforms a Ruleset using the specified backend.
func Compile(rs *Ruleset, backend Backend) *CompilationResult {
	var infra, network, runtime []Rule
	var warnings []string

	for _, r := range rs.Rules {
		switch NormalizeTarget(r.Target) {
		case TargetInfrastructure:
			infra = append(infra, r)
		case TargetNetwork:
			network = append(network, r)
		case TargetRuntime:
			runtime = append(runtime, r)
		case TargetHybrid:
			network = append(network, r)
			runtime = append(runtime, r)
		}
	}

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

	infraConfig, infraWarnings := backend.CompileInfrastructure(infra, network, rs)
	warnings = append(warnings, infraWarnings...)

	cfaisConfig := buildCFAISConfig(runtime, rs, laws, backend)

	return &CompilationResult{
		BackendName: backend.Name(),
		InfraConfig: infraConfig,
		CFAISConfig: cfaisConfig,
		Stats: CompileStats{
			TotalRules:   len(rs.Rules),
			InfraRules:   len(infra),
			NetworkRules: len(network),
			RuntimeRules: len(runtime),
			LawsCovered:  laws,
		},
		Warnings: warnings,
	}
}

func buildCFAISConfig(runtime []Rule, rs *Ruleset, activeLaws []string, backend Backend) map[string]interface{} {
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
		govLaws[string(l)] = map[string]interface{}{"active": lawsActive[string(l)]}
	}

	rules := make([]interface{}, 0, len(runtime))
	for _, r := range runtime {
		rules = append(rules, serializeRule(r))
	}

	return map[string]interface{}{
		"cfais_engine": map[string]interface{}{
			"version":         "0.2.0",
			"ruleset_id":      rs.ID,
			"ruleset_version": rs.Version,
			"framework":       rs.Framework,
			"sidecar":         backend.SidecarConfig(),
			"graceful_degradation": map[string]interface{}{
				"policy":     rs.GracefulDegradation,
				"timeout_ms": 100,
				"circuit_breaker": map[string]interface{}{
					"failure_threshold": 5, "recovery_s": 30,
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
		"id": r.ID, "name": r.Name, "description": r.Description,
		"scope": r.Scope, "severity": string(r.Severity),
		"enforcement": string(r.Enforcement), "laws": laws, "timeout_ms": r.TimeoutMs,
	}
	if r.Condition != nil {
		entry["condition"] = serializeCondition(*r.Condition)
	}
	if r.Remedy != nil {
		entry["remedy"] = map[string]interface{}{
			"action": r.Remedy.Action, "message": r.Remedy.Message,
			"alternative": r.Remedy.Alternative, "escalation": r.Remedy.Escalation,
			"audit_log": r.Remedy.AuditLog,
		}
	}
	if len(r.PolicyRefs) > 0 {
		refs := make([]map[string]interface{}, len(r.PolicyRefs))
		for i, p := range r.PolicyRefs {
			refs[i] = map[string]interface{}{
				"framework": p.Framework, "provision_id": p.ProvisionID,
				"description": p.Description,
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
		return map[string]interface{}{"logic": c.Logic, "conditions": subs}
	}
	m := map[string]interface{}{"field": c.Field, "operator": c.Operator, "value": c.Value}
	if c.ValueSource != "" {
		m["value_source"] = c.ValueSource
	}
	return m
}

// GetBackend returns a backend by name.
func GetBackend(name string) (Backend, error) {
	switch name {
	case "openshell":
		return &openshellBackend{}, nil
	case "docker":
		return &dockerBackend{}, nil
	case "standalone":
		return &standaloneBackend{}, nil
	default:
		return nil, fmt.Errorf("unknown backend: %s (available: openshell, docker, standalone)", name)
	}
}
