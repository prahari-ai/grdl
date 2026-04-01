// Package grdl implements the Governance Rule Definition Language.
//
// GRDL defines governance rules for autonomous AI agents. Rules compile into
// two enforcement layers: OpenShell YAML policies (infrastructure) and CFAIS
// runtime configuration (semantic decision governance).
package grdl

// GovernanceLaw represents one of the Seven Governance Laws from CFAIS.
type GovernanceLaw string

const (
	LawPrimacy         GovernanceLaw = "primacy"
	LawTransparency    GovernanceLaw = "transparency"
	LawAccountability  GovernanceLaw = "accountability"
	LawFairness        GovernanceLaw = "fairness"
	LawSafety          GovernanceLaw = "safety"
	LawPrivacy         GovernanceLaw = "privacy"
	LawGracefulDegrad  GovernanceLaw = "graceful_degradation"
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

type Target string

const (
	TargetOpenShellStatic  Target = "openshell_static"
	TargetOpenShellDynamic Target = "openshell_dynamic"
	TargetRuntime          Target = "runtime"
	TargetHybrid           Target = "hybrid"
)

// PolicyReference points to any governance document (constitution, regulation,
// corporate policy, DAO charter, AI safety framework). Jurisdiction-agnostic.
type PolicyReference struct {
	Framework   string   `yaml:"framework" json:"framework"`
	ProvisionID string   `yaml:"provision_id" json:"provision_id"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	References  []string `yaml:"references,omitempty" json:"references,omitempty"`
}

// Condition is a deterministic predicate on the action context.
// Never probabilistic — mathematical comparisons and set operations only.
type Condition struct {
	// Atomic condition fields
	Field       string      `yaml:"field,omitempty" json:"field,omitempty"`
	Operator    string      `yaml:"operator,omitempty" json:"operator,omitempty"`
	Value       interface{} `yaml:"value,omitempty" json:"value,omitempty"`
	ValueSource string      `yaml:"value_source,omitempty" json:"value_source,omitempty"`

	// Composite condition fields
	Logic      string      `yaml:"logic,omitempty" json:"logic,omitempty"` // all_of, any_of, none_of
	Conditions []Condition `yaml:"conditions,omitempty" json:"conditions,omitempty"`
}

// IsComposite returns true if this is a composite (AND/OR/NOT) condition.
func (c *Condition) IsComposite() bool {
	return c.Logic != ""
}

// Remedy defines the action taken on rule violation.
type Remedy struct {
	Action      string `yaml:"action" json:"action"`                                // block, modify, escalate, alert, throttle, log
	Message     string `yaml:"message" json:"message"`
	Alternative string `yaml:"alternative,omitempty" json:"alternative,omitempty"`
	Escalation  string `yaml:"escalation,omitempty" json:"escalation,omitempty"`
	AuditLog    bool   `yaml:"audit_log,omitempty" json:"audit_log"`
}

// Rule is a single governance rule.
type Rule struct {
	ID          string          `yaml:"id" json:"id"`
	Name        string          `yaml:"name" json:"name"`
	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string          `yaml:"version,omitempty" json:"version,omitempty"`
	Laws        []GovernanceLaw `yaml:"laws,omitempty" json:"laws,omitempty"`
	PolicyRefs  []PolicyReference `yaml:"policy_refs,omitempty" json:"policy_refs,omitempty"`
	Scope       string          `yaml:"scope,omitempty" json:"scope,omitempty"` // action type or "*"
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
type Ruleset struct {
	ID                   string      `yaml:"id" json:"id"`
	Name                 string      `yaml:"name" json:"name"`
	Description          string      `yaml:"description,omitempty" json:"description,omitempty"`
	Version              string      `yaml:"version,omitempty" json:"version,omitempty"`
	Framework            string      `yaml:"framework,omitempty" json:"framework,omitempty"`
	Rules                []Rule      `yaml:"rules" json:"rules"`
	DefaultEnforcement   Enforcement `yaml:"default_enforcement,omitempty" json:"default_enforcement,omitempty"`
	DefaultSeverity      Severity    `yaml:"default_severity,omitempty" json:"default_severity,omitempty"`
	GracefulDegradation  string      `yaml:"graceful_degradation,omitempty" json:"graceful_degradation,omitempty"`
	OpenShellVersionMin  string      `yaml:"openshell_version_min,omitempty" json:"openshell_version_min,omitempty"`
	CFAISVersionMin      string      `yaml:"cfais_version_min,omitempty" json:"cfais_version_min,omitempty"`
}
