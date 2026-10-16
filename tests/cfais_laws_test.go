package tests

import (
	"strings"
	"testing"

	"github.com/prahari-ai/grdl/internal/cfais"
	"github.com/prahari-ai/grdl/internal/grdl"
)

// ── Seven laws ────────────────────────────────────────────────────────────

func TestLegacyLawNamesNormalize(t *testing.T) {
	cases := map[string]grdl.GovernanceLaw{
		"PRIMACY":         grdl.LawHumanOverride,
		"FAIRNESS":        grdl.LawHumanDignity,
		"PRIVACY":         grdl.LawHumanDignity,
		"SAFETY":          grdl.LawSelfLimitation,
		"GRACEFUL_DEGRAD": grdl.LawGracefulDegrad,
		"Transparency":    grdl.LawTransparency,
		"VALUE_ALIGNMENT": grdl.LawValueAlignment,
	}
	for in, want := range cases {
		got, ok := grdl.NormalizeLaw(grdl.GovernanceLaw(in))
		if !ok || got != want {
			t.Errorf("%s: got %s (ok=%v), want %s", in, got, ok, want)
		}
	}
	if _, ok := grdl.NormalizeLaw("COMPLIANCE"); ok {
		t.Error("unknown law should not normalize")
	}
}

func TestUnknownLawRejected(t *testing.T) {
	_, err := grdl.ParseRuleset([]byte(`
id: x
name: x
rules:
  - id: r1
    name: r1
    laws: [CONTINUOUS_LEARNING]
`))
	if err == nil || !strings.Contains(err.Error(), "unknown governance law") {
		t.Fatalf("expected unknown-law error, got %v", err)
	}
}

func TestAllTemplatesUseCanonicalLaws(t *testing.T) {
	canonical := map[grdl.GovernanceLaw]bool{}
	for _, l := range grdl.AllLaws {
		canonical[l] = true
	}
	for _, name := range []string{
		"enterprise-agent-governance.grdl.yaml", "dao-governance.grdl.yaml",
		"ai-safety.grdl.yaml", "agentic-payments.grdl.yaml",
	} {
		rs, err := grdl.LoadRuleset(templatePath(name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, r := range rs.Rules {
			for _, l := range r.Laws {
				if !canonical[l] {
					t.Errorf("%s/%s uses non-canonical law %s", name, r.ID, l)
				}
			}
		}
	}
}

func TestCompiledLawsMarkedActive(t *testing.T) {
	rs, _ := grdl.LoadRuleset(templatePath("dao-governance.grdl.yaml"))
	b, _ := grdl.GetBackend("standalone")
	laws := grdl.Compile(rs, b).CFAISConfig["governance_laws"].(map[string]interface{})
	for _, l := range []string{"transparency", "human_override", "self_limitation"} {
		if laws[l].(map[string]interface{})["active"] != true {
			t.Errorf("law %s should be active in compiled config", l)
		}
	}
}

// ── Governance parameters ─────────────────────────────────────────────────

func TestProportionalityBoundValidated(t *testing.T) {
	for _, alpha := range []string{"0", "1", "1.5", "-0.1"} {
		_, err := grdl.ParseRuleset([]byte(`
id: x
name: x
governance_parameters:
  proportionality:
    ai_weight_bound: ` + alpha + `
rules: []
`))
		if err == nil {
			t.Errorf("alpha=%s should be rejected", alpha)
		}
	}
}

func TestMissingParameterReferenceRejected(t *testing.T) {
	_, err := grdl.ParseRuleset([]byte(`
id: x
name: x
rules:
  - id: r1
    name: r1
    condition:
      field: payload.ai_weight
      operator: gt
      value_source: payload.human_equivalent_weight
      scale_source: params.proportionality.ai_weight_bound
`))
	if err == nil || !strings.Contains(err.Error(), "params.proportionality.ai_weight_bound") {
		t.Fatalf("expected missing-parameter error, got %v", err)
	}
}

// ── DAO template: AI agent constraints from the MLSC paper ─────────────────

func aiVote(mutate func(p map[string]interface{})) *cfais.EvaluationContext {
	p := map[string]interface{}{
		"ai_weight":               25.0,
		"human_equivalent_weight": 100.0,
		"layer":                   "L0",
		"reasoning_chain":         "budget within mandate; aligns with proposal 42 rationale",
		"data_inputs":             []interface{}{"proposal-42", "treasury-report-q3"},
		"decision_criteria":       "approve if cost <= 2% of treasury",
	}
	if mutate != nil {
		mutate(p)
	}
	return &cfais.EvaluationContext{ActionID: "v", ActionType: "ai_vote", AgentID: "delegate-bot", Payload: p}
}

func TestAIVoteAllowedAtBound(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	if r := e.Evaluate(aiVote(nil)); r.Verdict != cfais.Allow {
		t.Fatalf("w_AI = alpha*w_human should be allowed, got %s (%s)", r.Verdict, r.RuleID)
	}
}

func TestAIVoteDenials(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	cases := []struct {
		name, rule string
		mutate     func(p map[string]interface{})
	}{
		{"over alpha bound", "cfais.ai_proportionality",
			func(p map[string]interface{}) { p["ai_weight"] = 25.01 }},
		{"missing human weight", "cfais.ai_vote_inputs_present",
			func(p map[string]interface{}) { delete(p, "human_equivalent_weight") }},
		{"no reasoning chain", "cfais.ai_vote_transparency",
			func(p map[string]interface{}) { delete(p, "reasoning_chain") }},
		{"no decision criteria", "cfais.ai_vote_transparency",
			func(p map[string]interface{}) { delete(p, "decision_criteria") }},
		{"L2 without approval", "cfais.ai_layer_restriction",
			func(p map[string]interface{}) { p["layer"] = "L2" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := e.Evaluate(aiVote(tc.mutate))
			if r.Verdict != cfais.Deny || r.RuleID != tc.rule {
				t.Errorf("expected deny by %s, got %s by %s", tc.rule, r.Verdict, r.RuleID)
			}
		})
	}
}

func TestAIVoteHigherLayerWithApproval(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	r := e.Evaluate(aiVote(func(p map[string]interface{}) {
		p["layer"] = "L1"
		p["multi_layer_approval"] = true
	}))
	if r.Verdict != cfais.Allow {
		t.Fatalf("expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
}

func TestReversibilityHold(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	exec := func(p map[string]interface{}) *cfais.EvaluationResult {
		return e.Evaluate(&cfais.EvaluationContext{ActionID: "x", ActionType: "decision_execution", Payload: p})
	}
	if r := exec(map[string]interface{}{"ai_votes_decisive": true, "seconds_since_decision": 3600}); r.RuleID != "cfais.reversibility_hold" {
		t.Errorf("AI-decided, inside window: expected hold, got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := exec(map[string]interface{}{"ai_votes_decisive": true}); r.RuleID != "cfais.reversibility_hold" {
		t.Errorf("AI-decided, no elapsed time: expected hold, got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := exec(map[string]interface{}{"ai_votes_decisive": true, "seconds_since_decision": 259201}); r.Verdict != cfais.Allow {
		t.Errorf("AI-decided, window passed: expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := exec(map[string]interface{}{"ai_votes_decisive": false, "seconds_since_decision": 10}); r.Verdict != cfais.Allow {
		t.Errorf("human-decided: expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := exec(map[string]interface{}{"seconds_since_decision": 10}); r.RuleID != "cfais.reversibility_flag_required" {
		t.Errorf("missing flag: expected block, got %s (%s)", r.Verdict, r.RuleID)
	}
}

func TestHumanRevote(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	valid := func() map[string]interface{} {
		return map[string]interface{}{
			"original_ai_decisive": true, "requested_by_type": "human",
			"seconds_since_decision": 7200, "ai_participants": 0,
		}
	}
	revote := func(p map[string]interface{}) *cfais.EvaluationResult {
		return e.Evaluate(&cfais.EvaluationContext{ActionID: "rv", ActionType: "human_revote", Payload: p})
	}
	if r := revote(valid()); r.Verdict != cfais.Allow {
		t.Fatalf("valid re-vote: expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
	bad := map[string]func(p map[string]interface{}){
		"window closed":      func(p map[string]interface{}) { p["seconds_since_decision"] = 300000 },
		"AI participants":    func(p map[string]interface{}) { p["ai_participants"] = 3 },
		"requested by agent": func(p map[string]interface{}) { p["requested_by_type"] = "ai_agent" },
		"human-decided":      func(p map[string]interface{}) { p["original_ai_decisive"] = false },
		"participants unset": func(p map[string]interface{}) { delete(p, "ai_participants") },
	}
	for name, mutate := range bad {
		p := valid()
		mutate(p)
		if r := revote(p); r.Verdict != cfais.Deny || r.RuleID != "cfais.revote_eligibility" {
			t.Errorf("%s: expected deny, got %s (%s)", name, r.Verdict, r.RuleID)
		}
	}
}
