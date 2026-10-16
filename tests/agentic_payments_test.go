package tests

import (
	"testing"

	"github.com/prahari-ai/grdl/internal/cfais"
	"github.com/prahari-ai/grdl/internal/grdl"
)

const paymentsTemplate = "agentic-payments.grdl.yaml"

// paymentContext returns a complete deployer policy.
func paymentContext() map[string]interface{} {
	return map[string]interface{}{
		"max_payment_amount":       1.00,
		"wallet_remaining_budget":  25.00,
		"approved_counterparties":  []interface{}{"inference.example", "quotes-api.example"},
		"allowed_assets":           []interface{}{"USDC"},
		"allowed_networks":         []interface{}{"base"},
		"max_payments_per_window":  100,
		"max_spend_per_window":     5.00,
		"human_approval_threshold": 0.50,
		"max_wallet_funding":       100.00,
	}
}

// compliantPayment returns a 2-cent inference payment that should be allowed.
func compliantPayment() map[string]interface{} {
	return map[string]interface{}{
		"amount":             0.02,
		"asset":              "USDC",
		"network":            "base",
		"recipient":          "inference.example",
		"purpose":            "inference: summarise filing",
		"funded_by":          "treasury-wallet-01",
		"on_behalf_of":       "research-team",
		"payments_in_window": 12,
		"spend_in_window":    0.40,
	}
}

func evalPayment(t *testing.T, e *cfais.Engine, mutate func(p, c map[string]interface{})) *cfais.EvaluationResult {
	t.Helper()
	p, c := compliantPayment(), paymentContext()
	if mutate != nil {
		mutate(p, c)
	}
	return e.Evaluate(&cfais.EvaluationContext{
		ActionID: "pay", ActionType: "agent_payment", AgentID: "research-agent",
		Payload: p, Context: c,
	})
}

func TestPaymentsTemplateCompiles(t *testing.T) {
	rs, err := grdl.LoadRuleset(templatePath(paymentsTemplate))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"openshell", "docker", "standalone"} {
		b, _ := grdl.GetBackend(name)
		r := grdl.Compile(rs, b)
		if r.Stats.TotalRules != 16 || r.Stats.RuntimeRules != 15 || r.Stats.InfraRules != 1 {
			t.Errorf("%s: got total=%d runtime=%d infra=%d", name,
				r.Stats.TotalRules, r.Stats.RuntimeRules, r.Stats.InfraRules)
		}
	}
}

func TestPaymentsAllowCompliant(t *testing.T) {
	e := loadEngine(t, paymentsTemplate)
	r := evalPayment(t, e, nil)
	if r.Verdict != cfais.Allow {
		t.Fatalf("expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
}

func TestPaymentsAllowApprovedLargePayment(t *testing.T) {
	e := loadEngine(t, paymentsTemplate)
	r := evalPayment(t, e, func(p, c map[string]interface{}) {
		p["amount"] = 0.80
		p["spend_in_window"] = 1.20
		p["human_approved"] = true
	})
	if r.Verdict != cfais.Allow {
		t.Fatalf("expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
}

func TestPaymentsDenials(t *testing.T) {
	cases := []struct {
		name   string
		rule   string
		mutate func(p, c map[string]interface{})
	}{
		{"missing policy", "graceful.payment_policy_loaded",
			func(p, c map[string]interface{}) { delete(c, "max_payment_amount") }},
		{"missing counters", "graceful.velocity_counters_present",
			func(p, c map[string]interface{}) { delete(p, "spend_in_window") }},
		{"no funder", "accountability.deployer_identity",
			func(p, c map[string]interface{}) { delete(p, "funded_by") }},
		{"no purpose", "transparency.payment_purpose",
			func(p, c map[string]interface{}) { delete(p, "purpose") }},
		{"wrong asset", "safety.asset_allowlist",
			func(p, c map[string]interface{}) { p["asset"] = "MEMECOIN" }},
		{"wrong network", "safety.network_allowlist",
			func(p, c map[string]interface{}) { p["network"] = "unknown-l2" }},
		{"unknown recipient", "safety.counterparty_allowlist",
			func(p, c map[string]interface{}) { p["recipient"] = "attacker.example" }},
		{"over per-payment cap", "safety.per_payment_cap",
			func(p, c map[string]interface{}) { p["amount"] = 3.00 }},
		{"over wallet budget", "safety.wallet_budget",
			func(p, c map[string]interface{}) {
				c["wallet_remaining_budget"] = 0.01
			}},
		{"payment loop", "safety.velocity_count",
			func(p, c map[string]interface{}) { p["payments_in_window"] = 100 }},
		{"window spend", "safety.velocity_value",
			func(p, c map[string]interface{}) { p["spend_in_window"] = 5.01 }},
		{"needs human", "primacy.human_approval_threshold",
			func(p, c map[string]interface{}) { p["amount"] = 0.80 }},
	}

	e := loadEngine(t, paymentsTemplate)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := evalPayment(t, e, tc.mutate)
			if r.Verdict != cfais.Deny || r.RuleID != tc.rule {
				t.Errorf("expected deny by %s, got %s by %s", tc.rule, r.Verdict, r.RuleID)
			}
		})
	}
}

func TestPaymentsFirstCounterpartyAudited(t *testing.T) {
	e := loadEngine(t, paymentsTemplate)
	r := evalPayment(t, e, func(p, c map[string]interface{}) {
		p["counterparty_first_seen"] = true
	})
	if r.Verdict != cfais.AllowWithAudit || r.RuleID != "safety.first_payment_to_counterparty" {
		t.Fatalf("expected allow_with_audit, got %s (%s)", r.Verdict, r.RuleID)
	}
}

func TestWalletFunding(t *testing.T) {
	e := loadEngine(t, paymentsTemplate)
	eval := func(p, c map[string]interface{}) *cfais.EvaluationResult {
		return e.Evaluate(&cfais.EvaluationContext{
			ActionID: "fund", ActionType: "wallet_funding", AgentID: "research-agent",
			Payload: p, Context: c,
		})
	}
	ctx := map[string]interface{}{"max_wallet_funding": 100.0}

	if r := eval(map[string]interface{}{"amount": 50.0, "authorized_by": "cfo"}, ctx); r.Verdict != cfais.Allow {
		t.Errorf("authorised top-up: expected allow, got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := eval(map[string]interface{}{"amount": 50.0}, ctx); r.RuleID != "accountability.funding_authorized" {
		t.Errorf("unauthorised top-up: got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := eval(map[string]interface{}{"amount": 500.0, "authorized_by": "cfo"}, ctx); r.RuleID != "safety.funding_cap" {
		t.Errorf("oversized top-up: got %s (%s)", r.Verdict, r.RuleID)
	}
	if r := eval(map[string]interface{}{"amount": 5.0, "authorized_by": "cfo"}, map[string]interface{}{}); r.RuleID != "safety.funding_cap" {
		t.Errorf("unconfigured ceiling should fail closed: got %s (%s)", r.Verdict, r.RuleID)
	}
}

func BenchmarkPaymentEvaluate(b *testing.B) {
	rs, _ := grdl.LoadRuleset(templatePath(paymentsTemplate))
	be, _ := grdl.GetBackend("standalone")
	e, _ := cfais.NewEngine(grdl.Compile(rs, be).CFAISConfig)
	ctx := &cfais.EvaluationContext{
		ActionID: "bench", ActionType: "agent_payment", AgentID: "research-agent",
		Payload: compliantPayment(), Context: paymentContext(),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Evaluate(ctx)
	}
}
