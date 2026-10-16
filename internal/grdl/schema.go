// Package grdl implements the Governance Rule Definition Language.
//
// GRDL defines governance rules for autonomous AI agents. It is
// runtime-agnostic — rules compile into any enforcement backend
// (OpenShell, Docker, Kubernetes, standalone).
//
// Patent reference: 202641000868 (NIYAM/GRDL)
package grdl

import "strings"

// GovernanceLaw is one of the seven CFAIS constitutional laws
// (patent 202641001109). Rule files may use any letter case; the loader
// normalizes names and maps legacy v0.1/v0.2 names via NormalizeLaw.
type GovernanceLaw string

const (
	LawHumanDignity   GovernanceLaw = "human_dignity"        // 1. no action may violate human dignity, privacy or equal treatment
	LawTransparency   GovernanceLaw = "transparency"         // 2. complete audit trail: reasoning, inputs, criteria
	LawAccountability GovernanceLaw = "accountability"       // 3. every action traceable to a responsible party
	LawHumanOverride  GovernanceLaw = "human_override"       // 4. authorised humans can override and reverse
	LawSelfLimitation GovernanceLaw = "self_limitation"      // 5. no capability, budget or authority beyond scope
	LawValueAlignment GovernanceLaw = "value_alignment"      // 6. actions stay aligned with encoded human values
	LawGracefulDegrad GovernanceLaw = "graceful_degradation" // 7. fail to a verified safe state, never open
)

// AllLaws lists the seven laws in canonical order.
var AllLaws = []GovernanceLaw{
	LawHumanDignity, LawTransparency, LawAccountability, LawHumanOverride,
	LawSelfLimitation, LawValueAlignment, LawGracefulDegrad,
}

// legacyLaws maps law names used before v0.3.0 to the canonical seven.
var legacyLaws = map[string]GovernanceLaw{
	"primacy":         LawHumanOverride,
	"fairness":        LawHumanDignity,
	"privacy":         LawHumanDignity,
	"safety":          LawSelfLimitation,
	"graceful_degrad": LawGracefulDegrad,
}

// NormalizeLaw lower-cases a law name and maps legacy names. It returns
// false if the name is not a known law.
func NormalizeLaw(l GovernanceLaw) (GovernanceLaw, bool) {
	name := GovernanceLaw(strings.ToLower(strings.TrimSpace(string(l))))
	if mapped, ok := legacyLaws[string(name)]; ok {
		return mapped, true
	}
	for _, known := range AllLaws {
		if name == known {
			return name, true
		}
	}
	return name, false
}

// GovernanceParameters are ruleset-wide settings that laws operate over.
// Rules reference them as params.<name>.<field>, e.g.
// params.proportionality.ai_weight_bound.
type GovernanceParameters struct {
	// Proportionality bounds AI influence relative to humans:
	// w_AI(a) <= alpha * w_human(s_a). Enforced under Laws 4 and 5.
	Proportionality *Proportionality `yaml:"proportionality,omitempty" json:"proportionality,omitempty"`
	// Reversibility keeps decisions where AI votes were decisive open to an
	// AI-excluded human re-vote for a fixed window. Enforced under Law 4.
	Reversibility *Reversibility `yaml:"reversibility,omitempty" json:"reversibility,omitempty"`
}

type Proportionality struct {
	AIWeightBound float64  `yaml:"ai_weight_bound" json:"ai_weight_bound"` // alpha, in (0,1)
	Scope         []string `yaml:"scope,omitempty" json:"scope,omitempty"`
}

type Reversibility struct {
	ChallengeWindowS int      `yaml:"challenge_window_s" json:"challenge_window_s"` // > 0
	Scope            []string `yaml:"scope,omitempty" json:"scope,omitempty"`
}

type Severity string

const (
	SevCritical Severity = "critical"
	SevHigh     Severity = "high"
	SevMedium   Severity = "medium"
	SevLow      Severity = "low"
	SevAdvisory Severity = "advisory"
)

type Enforcement string

const (
	EnfEnforce Enforcement = "enforce"
	EnfAudit   Enforcement = "audit"
	EnfShadow  Enforcement = "shadow"
)

// Target specifies WHERE a rule is enforced. These are runtime-agnostic
// categories — each backend maps them to its own enforcement mechanism.
//
//	infrastructure → OpenShell: filesystem+process policy
//	                 Docker: seccomp+mount policy
//	                 Standalone: advisory only
//	network        → OpenShell: network_policies with REST inspection
//	                 Docker: nftables rules
//	                 Standalone: advisory only
//	runtime        → All backends: CFAIS sidecar evaluation per request
//	hybrid         → Both network + runtime
type Target string

const (
	TargetInfrastructure Target = "infrastructure" // filesystem, process, kernel-level
	TargetNetwork        Target = "network"        // network endpoint control
	TargetRuntime        Target = "runtime"        // CFAIS sidecar per-request evaluation
	TargetHybrid         Target = "hybrid"         // network + runtime

	// Legacy aliases (accepted by loader, mapped to new values)
	TargetOpenShellStatic  Target = "openshell_static"
	TargetOpenShellDynamic Target = "openshell_dynamic"
)

// NormalizeTarget maps legacy target values to the new runtime-agnostic ones.
func NormalizeTarget(t Target) Target {
	switch t {
	case TargetOpenShellStatic:
		return TargetInfrastructure
	case TargetOpenShellDynamic:
		return TargetNetwork
	default:
		return t
	}
}

// PolicyReference points to any governance document. Jurisdiction-agnostic.
// Implements patent claim [0041]: constitutional hierarchy binding.
type PolicyReference struct {
	Framework   string   `yaml:"framework" json:"framework"`
	ProvisionID string   `yaml:"provision_id" json:"provision_id"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	References  []string `yaml:"references,omitempty" json:"references,omitempty"`
}

// Condition is a deterministic predicate. Never probabilistic.
// Patent claim [0072]: mathematical certainty, not probabilistic assessments.
type Condition struct {
	Field       string      `yaml:"field,omitempty" json:"field,omitempty"`
	Operator    string      `yaml:"operator,omitempty" json:"operator,omitempty"`
	Value       interface{} `yaml:"value,omitempty" json:"value,omitempty"`
	ValueSource string      `yaml:"value_source,omitempty" json:"value_source,omitempty"`
	// ScaleSource multiplies the comparison value by another resolved field,
	// so a rule can express "actual > factor * reference" (e.g. the
	// proportionality bound w_AI <= alpha * w_human).
	ScaleSource string      `yaml:"scale_source,omitempty" json:"scale_source,omitempty"`
	Logic       string      `yaml:"logic,omitempty" json:"logic,omitempty"`
	Conditions  []Condition `yaml:"conditions,omitempty" json:"conditions,omitempty"`
}

func (c *Condition) IsComposite() bool { return c.Logic != "" }

type Remedy struct {
	Action      string `yaml:"action" json:"action"`
	Message     string `yaml:"message" json:"message"`
	Alternative string `yaml:"alternative,omitempty" json:"alternative,omitempty"`
	Escalation  string `yaml:"escalation,omitempty" json:"escalation,omitempty"`
	AuditLog    bool   `yaml:"audit_log,omitempty" json:"audit_log"`
}

type Rule struct {
	ID           string            `yaml:"id" json:"id"`
	Name         string            `yaml:"name" json:"name"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	Version      string            `yaml:"version,omitempty" json:"version,omitempty"`
	Laws         []GovernanceLaw   `yaml:"laws,omitempty" json:"laws,omitempty"`
	PolicyRefs   []PolicyReference `yaml:"policy_refs,omitempty" json:"policy_refs,omitempty"`
	Scope        string            `yaml:"scope,omitempty" json:"scope,omitempty"`
	Condition    *Condition        `yaml:"condition,omitempty" json:"condition,omitempty"`
	Severity     Severity          `yaml:"severity,omitempty" json:"severity,omitempty"`
	Enforcement  Enforcement       `yaml:"enforcement,omitempty" json:"enforcement,omitempty"`
	Target       Target            `yaml:"target,omitempty" json:"target,omitempty"`
	Remedy       *Remedy           `yaml:"remedy,omitempty" json:"remedy,omitempty"`
	FormalProof  string            `yaml:"formal_proof,omitempty" json:"formal_proof,omitempty"`
	FallbackRule string            `yaml:"fallback_rule,omitempty" json:"fallback_rule,omitempty"`
	TimeoutMs    int               `yaml:"timeout_ms,omitempty" json:"timeout_ms,omitempty"`
}

// Ruleset is the top-level compilation unit.
// Patent [0046]-[0051]: metadata, constitutional_hierarchy, stakeholder_layers,
// constraints, verification_bindings, portability.
type Ruleset struct {
	ID                   string                `yaml:"id" json:"id"`
	Name                 string                `yaml:"name" json:"name"`
	Description          string                `yaml:"description,omitempty" json:"description,omitempty"`
	Version              string                `yaml:"version,omitempty" json:"version,omitempty"`
	Framework            string                `yaml:"framework,omitempty" json:"framework,omitempty"`
	Rules                []Rule                `yaml:"rules" json:"rules"`
	DefaultEnforcement   Enforcement           `yaml:"default_enforcement,omitempty" json:"default_enforcement,omitempty"`
	DefaultSeverity      Severity              `yaml:"default_severity,omitempty" json:"default_severity,omitempty"`
	GracefulDegradation  string                `yaml:"graceful_degradation,omitempty" json:"graceful_degradation,omitempty"`
	GovernanceParameters *GovernanceParameters `yaml:"governance_parameters,omitempty" json:"governance_parameters,omitempty"`
}
