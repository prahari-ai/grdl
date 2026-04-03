package tests

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/prahari-ai/grdl/internal/cfais"
	"github.com/prahari-ai/grdl/internal/grdl"
)

func templatePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "examples", "templates", name)
}

func loadEngine(t *testing.T, template string) *cfais.Engine {
	t.Helper()
	rs, err := grdl.LoadRuleset(templatePath(template))
	if err != nil {
		t.Fatalf("loading: %v", err)
	}
	backend, _ := grdl.GetBackend("standalone")
	compiled := grdl.Compile(rs, backend)
	engine, err := cfais.NewEngine(compiled.CFAISConfig)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	return engine
}

func TestLoadAllTemplates(t *testing.T) {
	for _, name := range []string{
		"enterprise-agent-governance.grdl.yaml",
		"dao-governance.grdl.yaml",
		"ai-safety.grdl.yaml",
	} {
		rs, err := grdl.LoadRuleset(templatePath(name))
		if err != nil {
			t.Errorf("load %s: %v", name, err)
			continue
		}
		if len(rs.Rules) == 0 {
			t.Errorf("%s has no rules", name)
		}
	}
}

func TestCompileAllBackends(t *testing.T) {
	rs, _ := grdl.LoadRuleset(templatePath("enterprise-agent-governance.grdl.yaml"))
	for _, name := range []string{"openshell", "docker", "standalone"} {
		backend, err := grdl.GetBackend(name)
		if err != nil {
			t.Errorf("backend %s: %v", name, err)
			continue
		}
		result := grdl.Compile(rs, backend)
		if result.BackendName != name {
			t.Errorf("expected backend %s, got %s", name, result.BackendName)
		}
		if result.Stats.TotalRules != 11 {
			t.Errorf("%s: expected 11 rules, got %d", name, result.Stats.TotalRules)
		}
		if result.Stats.RuntimeRules != 9 {
			t.Errorf("%s: expected 9 runtime, got %d", name, result.Stats.RuntimeRules)
		}
		t.Logf("%s backend: %d total, %d infra, %d net, %d runtime, %d warnings",
			name, result.Stats.TotalRules, result.Stats.InfraRules,
			result.Stats.NetworkRules, result.Stats.RuntimeRules, len(result.Warnings))
	}
}

func TestOpenShellProducesValidPolicy(t *testing.T) {
	rs, _ := grdl.LoadRuleset(templatePath("enterprise-agent-governance.grdl.yaml"))
	backend, _ := grdl.GetBackend("openshell")
	result := grdl.Compile(rs, backend)

	pol := result.InfraConfig
	if pol["version"] != 1 {
		t.Error("version should be 1")
	}
	proc, ok := pol["process"].(map[string]interface{})
	if !ok || proc["run_as_user"] != "sandbox" {
		t.Error("process.run_as_user should be sandbox")
	}
	np, ok := pol["network_policies"].(map[string]interface{})
	if !ok {
		t.Fatal("missing network_policies")
	}
	if _, ok := np["cfais_sidecar"]; !ok {
		t.Error("missing cfais_sidecar")
	}
}

func TestStandaloneAdvisoryMode(t *testing.T) {
	rs, _ := grdl.LoadRuleset(templatePath("enterprise-agent-governance.grdl.yaml"))
	backend, _ := grdl.GetBackend("standalone")
	result := grdl.Compile(rs, backend)

	if result.InfraConfig["mode"] != "advisory" {
		t.Error("standalone should be advisory mode")
	}
	if len(result.Warnings) == 0 {
		t.Error("standalone should warn about advisory-only enforcement")
	}
}

func TestDenyPrivilegeEscalation(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t1", ActionType: "access_control",
		Payload: map[string]interface{}{"action": "grant_permission", "target_is_self": true},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestDenyBudgetOverrun(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t2", ActionType: "resource_consumption", AgentID: "x",
		Payload: map[string]interface{}{"estimated_cost": 5000, "on_behalf_of": "ops"},
		Context: map[string]interface{}{"remaining_budget": 1000},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestAllowWithinBudget(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t3", ActionType: "resource_consumption", AgentID: "agent-a",
		Payload: map[string]interface{}{"estimated_cost": 500, "on_behalf_of": "ops"},
		Context: map[string]interface{}{"remaining_budget": 1000},
	})
	if r.Verdict != cfais.Allow {
		t.Errorf("expected allow, got %s (rule: %s)", r.Verdict, r.RuleID)
	}
}

func TestDenyPIIWithoutJustification(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t4", ActionType: "data_access", AgentID: "agent-s",
		Payload: map[string]interface{}{"contains_pii": true, "on_behalf_of": "team", "fields_requested_count": 3},
		Context: map[string]interface{}{"max_fields_per_request": 20},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestAllowPIIWithJustification(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t5", ActionType: "data_access", AgentID: "agent-s",
		Payload: map[string]interface{}{
			"contains_pii": true, "purpose_justification": "ticket #1234",
			"on_behalf_of": "team", "fields_requested_count": 3,
		},
		Context: map[string]interface{}{"max_fields_per_request": 10},
	})
	if r.Verdict != cfais.Allow {
		t.Errorf("expected allow, got %s (rule: %s)", r.Verdict, r.RuleID)
	}
}

func TestDenyCascadingDepth(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t6", ActionType: "agent_delegation", AgentID: "orch",
		Payload: map[string]interface{}{"delegation_depth": 8, "on_behalf_of": "pipe"},
		Context: map[string]interface{}{"max_delegation_depth": 5},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestDAOTreasuryLimit(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t7", ActionType: "treasury_transaction",
		Payload: map[string]interface{}{"amount": 50000},
		Context: map[string]interface{}{"treasury_auto_limit": 10000},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestDAOQuorum(t *testing.T) {
	e := loadEngine(t, "dao-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t8", ActionType: "governance_decision",
		Payload: map[string]interface{}{"participation_rate": 0.15, "max_voter_weight": 0.05},
		Context: map[string]interface{}{"quorum_threshold": 0.33, "max_voting_weight": 0.10},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestSafetyToolAllowlist(t *testing.T) {
	e := loadEngine(t, "ai-safety.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t9", ActionType: "tool_invocation",
		Payload: map[string]interface{}{"tool_name": "execute_shell"},
		Context: map[string]interface{}{"allowed_tools": []interface{}{"search", "calculator"}},
	})
	if r.Verdict != cfais.Deny {
		t.Errorf("expected deny, got %s", r.Verdict)
	}
}

func TestGracefulDegradation(t *testing.T) {
	e := loadEngine(t, "enterprise-agent-governance.grdl.yaml")
	r := e.Evaluate(&cfais.EvaluationContext{
		ActionID: "t10", ActionType: "resource_consumption", AgentID: "x",
		Payload: map[string]interface{}{"estimated_cost": "not_a_number", "on_behalf_of": "t"},
		Context: map[string]interface{}{"remaining_budget": 1000},
	})
	if r.Verdict != cfais.Allow && r.Verdict != cfais.AllowWithAudit {
		t.Errorf("expected graceful fallback, got %s", r.Verdict)
	}
}

func BenchmarkEvaluate(b *testing.B) {
	rs, _ := grdl.LoadRuleset(templatePath("enterprise-agent-governance.grdl.yaml"))
	backend, _ := grdl.GetBackend("standalone")
	compiled := grdl.Compile(rs, backend)
	engine, _ := cfais.NewEngine(compiled.CFAISConfig)
	ctx := &cfais.EvaluationContext{
		ActionID: "bench", ActionType: "data_retrieval", AgentID: "perf",
		Payload: map[string]interface{}{"query": "x", "on_behalf_of": "test"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Evaluate(ctx)
	}
}
