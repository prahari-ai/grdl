// Package grdl implements the Governance Rule Definition Language.
//
// GRDL defines governance rules for autonomous AI agents. It is
// runtime-agnostic — rules compile into any enforcement backend
// (OpenShell, Docker, Kubernetes, standalone).
//
// Patent reference: 202641000868 (NIYAM/GRDL)
package grdl

// GovernanceLaw represents one of the Seven Governance Laws from CFAIS.
type GovernanceLaw string

const (
	LawPrimacy        GovernanceLaw = "primacy"
	LawTransparency   GovernanceLaw = "transparency"
	LawAccountability GovernanceLaw = "accountability"
	LawFairness       GovernanceLaw = "fairness"
	LawSafety         GovernanceLaw = "safety"
	LawPrivacy        GovernanceLaw = "privacy"
	LawGracefulDegrad GovernanceLaw = "graceful_degradation"
)

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
//   infrastructure → OpenShell: filesystem+process policy
//                    Docker: seccomp+mount policy
//                    Standalone: advisory only
//   network        → OpenShell: network_policies with REST inspection
//                    Docker: nftables rules
//                    Standalone: advisory only
//   runtime        → All backends: CFAIS sidecar evaluation per request
//   hybrid         → Both network + runtime
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
	ID          string          `yaml:"id" json:"id"`
	Name        string          `yaml:"name" json:"name"`
	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string          `yaml:"version,omitempty" json:"version,omitempty"`
	Laws        []GovernanceLaw `yaml:"laws,omitempty" json:"laws,omitempty"`
	PolicyRefs  []PolicyReference `yaml:"policy_refs,omitempty" json:"policy_refs,omitempty"`
	Scope       string          `yaml:"scope,omitempty" json:"scope,omitempty"`
	Condition   *Condition      `yaml:"condition,omitempty" json:"condition,omitempty"`
	Severity    Severity        `yaml:"severity,omitempty" json:"severity,omitempty"`
	Enforcement Enforcement     `yaml:"enforcement,omitempty" json:"enforcement,omitempty"`
	Target      Target          `yaml:"target,omitempty" json:"target,omitempty"`
	Remedy      *Remedy         `yaml:"remedy,omitempty" json:"remedy,omitempty"`
	FormalProof string          `yaml:"formal_proof,omitempty" json:"formal_proof,omitempty"`
	FallbackRule string         `yaml:"fallback_rule,omitempty" json:"fallback_rule,omitempty"`
	TimeoutMs   int             `yaml:"timeout_ms,omitempty" json:"timeout_ms,omitempty"`
}

// Ruleset is the top-level compilation unit.
// Patent [0046]-[0051]: metadata, constitutional_hierarchy, stakeholder_layers,
// constraints, verification_bindings, portability.
type Ruleset struct {
	ID                  string      `yaml:"id" json:"id"`
	Name                string      `yaml:"name" json:"name"`
	Description         string      `yaml:"description,omitempty" json:"description,omitempty"`
	Version             string      `yaml:"version,omitempty" json:"version,omitempty"`
	Framework           string      `yaml:"framework,omitempty" json:"framework,omitempty"`
	Rules               []Rule      `yaml:"rules" json:"rules"`
	DefaultEnforcement  Enforcement `yaml:"default_enforcement,omitempty" json:"default_enforcement,omitempty"`
	DefaultSeverity     Severity    `yaml:"default_severity,omitempty" json:"default_severity,omitempty"`
	GracefulDegradation string     `yaml:"graceful_degradation,omitempty" json:"graceful_degradation,omitempty"`
}
