// Package cfais implements the Constitutional Framework for Autonomous
// Intelligent Systems — a deterministic governance rule evaluator.
//
// Performance target: <1μs per evaluation under concurrent load.
// The engine holds no mutable state between evaluations (thread-safe).
package cfais

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Verdict is the result of a governance evaluation.
type Verdict string

const (
	Allow          Verdict = "allow"
	Deny           Verdict = "deny"
	AllowWithAudit Verdict = "allow_with_audit"
	Modify         Verdict = "modify"
	Escalate       Verdict = "escalate"
)

// EvaluationContext holds all data for evaluating an agent action.
type EvaluationContext struct {
	ActionID   string                 `json:"action_id"`
	ActionType string                 `json:"action_type"`
	AgentID    string                 `json:"agent_id"`
	Payload    map[string]interface{} `json:"payload"`
	Context    map[string]interface{} `json:"context"`
	HTTPMethod string                `json:"http_method"`
	HTTPPath   string                `json:"http_path"`
	TargetHost string                `json:"target_host"`
	TargetPort int                   `json:"target_port"`
}

// EvaluationResult holds the verdict with full transparency data.
type EvaluationResult struct {
	Verdict          Verdict  `json:"verdict"`
	RuleID           string   `json:"rule_id,omitempty"`
	RuleName         string   `json:"rule_name,omitempty"`
	LawsInvoked      []string `json:"laws_invoked"`
	PolicyRefsCited  []string `json:"policy_refs_cited,omitempty"`
	Explanation      string   `json:"explanation"`
	RemedyAction     string   `json:"remedy_action,omitempty"`
	RemedyMessage    string   `json:"remedy_message,omitempty"`
	Alternative      string   `json:"alternative,omitempty"`
	EvalTimeUs       float64  `json:"evaluation_time_us"`
	AuditLog         bool     `json:"audit_log"`
	FormallyVerified bool     `json:"formally_verified,omitempty"`
	FormalProof      string   `json:"formal_proof,omitempty"`
}

// ruleEntry is an internal representation of a compiled rule for fast evaluation.
type ruleEntry struct {
	ID          string
	Name        string
	Scope       string
	Severity    string
	Enforcement string
	Laws        []string
	Condition   *conditionNode
	Remedy      *remedyEntry
	PolicyRefs  []string
	FormalProof string
}

type conditionNode struct {
	// Atomic
	Field       string
	Operator    string
	Value       interface{}
	ValueSource string
	// Composite
	Logic    string // all_of, any_of, none_of
	Children []*conditionNode
}

type remedyEntry struct {
	Action      string
	Message     string
	Alternative string
	Escalation  string
	AuditLog    bool
}

// Engine is the CFAIS governance evaluator. Safe for concurrent use.
type Engine struct {
	scopeIndex    map[string][]*ruleEntry // rules indexed by scope
	wildcardRules []*ruleEntry            // rules with scope "*"
	gracefulPolicy string
	config        map[string]interface{}
}

// NewEngine creates an engine from compiled CFAIS runtime config.
func NewEngine(config map[string]interface{}) (*Engine, error) {
	e := &Engine{
		scopeIndex: make(map[string][]*ruleEntry),
		config:     config,
	}

	// Extract graceful degradation policy
	if eng, ok := config["cfais_engine"].(map[string]interface{}); ok {
		if gd, ok := eng["graceful_degradation"].(map[string]interface{}); ok {
			if p, ok := gd["policy"].(string); ok {
				e.gracefulPolicy = p
			}
		}
	}
	if e.gracefulPolicy == "" {
		e.gracefulPolicy = "allow_with_audit"
	}

	// Parse rules
	rawRules, ok := config["rules"].([]interface{})
	if !ok {
		return e, nil // no rules is valid
	}

	for _, raw := range rawRules {
		rm, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		rule := parseRuleEntry(rm)
		if rule.Scope == "*" {
			e.wildcardRules = append(e.wildcardRules, rule)
		} else {
			e.scopeIndex[rule.Scope] = append(e.scopeIndex[rule.Scope], rule)
		}
	}

	return e, nil
}

// RuleCount returns the number of loaded rules.
func (e *Engine) RuleCount() int {
	count := len(e.wildcardRules)
	for _, rules := range e.scopeIndex {
		count += len(rules)
	}
	return count
}

// Evaluate checks an agent action against all applicable rules.
// First-match-deny semantics. Thread-safe.
func (e *Engine) Evaluate(ctx *EvaluationContext) *EvaluationResult {
	start := time.Now()

	result := e.evaluateRules(ctx)

	result.EvalTimeUs = float64(time.Since(start).Nanoseconds()) / 1000.0
	return result
}

func (e *Engine) evaluateRules(ctx *EvaluationContext) *EvaluationResult {
	// Scope-specific rules first, then wildcards
	applicable := make([]*ruleEntry, 0, 16)
	if scoped, ok := e.scopeIndex[ctx.ActionType]; ok {
		applicable = append(applicable, scoped...)
	}
	applicable = append(applicable, e.wildcardRules...)

	for _, rule := range applicable {
		if rule.Enforcement == "shadow" {
			continue
		}

		// Evaluate condition
		if rule.Condition != nil {
			match, err := e.evalCondition(rule.Condition, ctx)
			if err != nil {
				// Graceful degradation on eval error
				continue
			}
			if !match {
				continue
			}
		}

		// Condition matched — apply rule
		switch rule.Severity {
		case "critical", "high":
			msg := fmt.Sprintf("Blocked by rule: %s", rule.Name)
			var remedyAction, remedyMsg, alt string
			if rule.Remedy != nil {
				msg = rule.Remedy.Message
				remedyAction = rule.Remedy.Action
				remedyMsg = rule.Remedy.Message
				alt = rule.Remedy.Alternative
			}
			return &EvaluationResult{
				Verdict:          Deny,
				RuleID:           rule.ID,
				RuleName:         rule.Name,
				LawsInvoked:      rule.Laws,
				PolicyRefsCited:  rule.PolicyRefs,
				Explanation:      msg,
				RemedyAction:     remedyAction,
				RemedyMessage:    remedyMsg,
				Alternative:      alt,
				AuditLog:         true,
				FormallyVerified: rule.FormalProof != "",
				FormalProof:      rule.FormalProof,
			}

		case "medium":
			return &EvaluationResult{
				Verdict:     AllowWithAudit,
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				LawsInvoked: rule.Laws,
				Explanation: fmt.Sprintf("Rule '%s' flagged for review.", rule.Name),
				AuditLog:    true,
			}
		}

		// low / advisory — continue checking
		if rule.Enforcement == "audit" {
			return &EvaluationResult{
				Verdict:     AllowWithAudit,
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				LawsInvoked: rule.Laws,
				Explanation: fmt.Sprintf("Rule '%s' matched (audit mode).", rule.Name),
				AuditLog:    true,
			}
		}
	}

	return &EvaluationResult{
		Verdict:     Allow,
		Explanation: "No governance rules triggered. Action permitted.",
	}
}

// evalCondition evaluates a condition tree against the context.
func (e *Engine) evalCondition(node *conditionNode, ctx *EvaluationContext) (bool, error) {
	if node.Logic != "" {
		return e.evalComposite(node, ctx)
	}
	return e.evalAtomic(node, ctx)
}

func (e *Engine) evalComposite(node *conditionNode, ctx *EvaluationContext) (bool, error) {
	switch node.Logic {
	case "all_of":
		for _, child := range node.Children {
			match, err := e.evalCondition(child, ctx)
			if err != nil {
				return false, err
			}
			if !match {
				return false, nil
			}
		}
		return true, nil

	case "any_of":
		for _, child := range node.Children {
			match, err := e.evalCondition(child, ctx)
			if err != nil {
				continue // skip errored sub-conditions in any_of
			}
			if match {
				return true, nil
			}
		}
		return false, nil

	case "none_of":
		for _, child := range node.Children {
			match, err := e.evalCondition(child, ctx)
			if err != nil {
				continue
			}
			if match {
				return false, nil
			}
		}
		return true, nil
	}
	return false, fmt.Errorf("unknown logic: %s", node.Logic)
}

func (e *Engine) evalAtomic(node *conditionNode, ctx *EvaluationContext) (bool, error) {
	actual := e.resolveField(node.Field, ctx)
	expected := node.Value

	// Dynamic value resolution
	if node.ValueSource != "" {
		if resolved := e.resolveField(node.ValueSource, ctx); resolved != nil {
			expected = resolved
		}
	}

	switch node.Operator {
	case "eq":
		return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected), nil
	case "ne":
		return fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected), nil
	case "gt":
		a, b, err := toFloats(actual, expected)
		if err != nil {
			return false, nil // type mismatch → condition not met
		}
		return a > b, nil
	case "gte":
		a, b, err := toFloats(actual, expected)
		if err != nil {
			return false, nil
		}
		return a >= b, nil
	case "lt":
		a, b, err := toFloats(actual, expected)
		if err != nil {
			return false, nil
		}
		return a < b, nil
	case "lte":
		a, b, err := toFloats(actual, expected)
		if err != nil {
			return false, nil
		}
		return a <= b, nil
	case "in":
		return inSlice(actual, expected), nil
	case "not_in":
		return !inSlice(actual, expected), nil
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected)), nil
	case "exists":
		fieldExists := actual != nil
		if b, ok := expected.(bool); ok {
			return fieldExists == b, nil
		}
		return fieldExists, nil
	case "between":
		return evalBetween(actual, expected)
	}
	return false, fmt.Errorf("unknown operator: %s", node.Operator)
}

// resolveField resolves a dotted path against the evaluation context.
func (e *Engine) resolveField(path string, ctx *EvaluationContext) interface{} {
	parts := strings.SplitN(path, ".", 2)
	root := parts[0]

	// Direct context attributes
	switch root {
	case "action_type":
		return ctx.ActionType
	case "agent_id":
		return ctx.AgentID
	case "http_method":
		return ctx.HTTPMethod
	case "http_path":
		return ctx.HTTPPath
	case "target_host":
		return ctx.TargetHost
	}

	// Nested map lookups
	var source map[string]interface{}
	switch root {
	case "payload":
		source = ctx.Payload
	case "context":
		source = ctx.Context
	case "config":
		source = e.config
	default:
		// Try payload as default
		source = ctx.Payload
		// Use full path since root wasn't a known prefix
		return nestedGet(source, path)
	}

	if len(parts) < 2 {
		return source
	}
	return nestedGet(source, parts[1])
}

// --- helpers ---

func nestedGet(m map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = m
	for _, key := range parts {
		cm, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = cm[key]
	}
	return current
}

func toFloats(a, b interface{}) (float64, float64, error) {
	af, err := toFloat(a)
	if err != nil {
		return 0, 0, err
	}
	bf, err := toFloat(b)
	if err != nil {
		return 0, 0, err
	}
	return af, bf, nil
}

func toFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	case json.Number:
		return val.Float64()
	case nil:
		return 0, fmt.Errorf("nil value")
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
	}
}

func inSlice(actual interface{}, expected interface{}) bool {
	actualStr := fmt.Sprintf("%v", actual)
	switch exp := expected.(type) {
	case []interface{}:
		for _, item := range exp {
			if fmt.Sprintf("%v", item) == actualStr {
				return true
			}
		}
	case []string:
		for _, item := range exp {
			if item == actualStr {
				return true
			}
		}
	}
	return false
}

func evalBetween(actual, expected interface{}) (bool, error) {
	bounds, ok := expected.([]interface{})
	if !ok || len(bounds) != 2 {
		return false, fmt.Errorf("between requires [low, high]")
	}
	a, err := toFloat(actual)
	if err != nil {
		return false, nil
	}
	low, err := toFloat(bounds[0])
	if err != nil {
		return false, nil
	}
	high, err := toFloat(bounds[1])
	if err != nil {
		return false, nil
	}
	return a >= low && a <= high, nil
}

// parseRuleEntry converts a map[string]interface{} into a ruleEntry.
func parseRuleEntry(m map[string]interface{}) *ruleEntry {
	r := &ruleEntry{
		ID:          getString(m, "id"),
		Name:        getString(m, "name"),
		Scope:       getString(m, "scope"),
		Severity:    getString(m, "severity"),
		Enforcement: getString(m, "enforcement"),
		FormalProof: getString(m, "formal_proof"),
	}

	if laws, ok := m["laws"].([]interface{}); ok {
		for _, l := range laws {
			r.Laws = append(r.Laws, fmt.Sprintf("%v", l))
		}
	}

	if cond, ok := m["condition"].(map[string]interface{}); ok {
		r.Condition = parseConditionNode(cond)
	}

	if rem, ok := m["remedy"].(map[string]interface{}); ok {
		r.Remedy = &remedyEntry{
			Action:      getString(rem, "action"),
			Message:     getString(rem, "message"),
			Alternative: getString(rem, "alternative"),
			Escalation:  getString(rem, "escalation"),
			AuditLog:    getBool(rem, "audit_log"),
		}
	}

	if refs, ok := m["policy_refs"].([]interface{}); ok {
		for _, ref := range refs {
			if rm, ok := ref.(map[string]interface{}); ok {
				r.PolicyRefs = append(r.PolicyRefs, getString(rm, "provision_id"))
			}
		}
	}

	return r
}

func parseConditionNode(m map[string]interface{}) *conditionNode {
	node := &conditionNode{
		Logic:       getString(m, "logic"),
		Field:       getString(m, "field"),
		Operator:    getString(m, "operator"),
		Value:       m["value"],
		ValueSource: getString(m, "value_source"),
	}

	if subs, ok := m["conditions"].([]interface{}); ok {
		for _, sub := range subs {
			if sm, ok := sub.(map[string]interface{}); ok {
				node.Children = append(node.Children, parseConditionNode(sm))
			}
		}
	}
	return node
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
