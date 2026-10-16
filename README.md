# GRDL — Governance Rule Definition Language

**The YAML standard for AI agent governance. Runtime-agnostic. Deterministic. Microsecond-scale.**

GRDL compiles structured governance rules into any enforcement backend — NVIDIA OpenShell, Docker, Kubernetes, or standalone HTTP. One ruleset, any runtime.

---

## The problem

Autonomous AI agents (Gemma 4, OpenClaw, Claude Code, custom LLM agents) make decisions and take actions without human input. The industry has infrastructure security (OpenShell, Docker sandboxes). Nobody has decision governance — should this agent spend beyond its budget? Can it escalate its own permissions? Is the delegation chain too deep?

GRDL is the missing layer.

## How it works

```
Agent decides to act → CFAIS sidecar evaluates → ALLOW or DENY → Agent proceeds or adjusts
```

Write governance rules in YAML. The compiler produces enforcement config for your runtime:

```bash
prahari compile rules.grdl.yaml --backend openshell    # NVIDIA OpenShell YAML
prahari compile rules.grdl.yaml --backend docker       # Docker Compose + seccomp
prahari compile rules.grdl.yaml --backend standalone   # HTTP sidecar only (works anywhere)
```

The CFAIS engine is deterministic — mathematical comparisons, not ML. Microsecond-scale evaluation. Single static binary.

## Quick start

```bash
go install github.com/prahari-ai/grdl/cmd/prahari@latest

# Pick a template, compile, test
prahari validate examples/templates/enterprise-agent-governance.grdl.yaml
prahari compile examples/templates/enterprise-agent-governance.grdl.yaml --backend standalone
prahari evaluate examples/templates/enterprise-agent-governance.grdl.yaml test-action.json

# Run the governance sidecar
prahari serve examples/templates/enterprise-agent-governance.grdl.yaml --addr :9700
```

Any agent framework (Gemma 4 + Ollama, OpenClaw, LangChain, CrewAI) calls `POST http://localhost:9700/evaluate` before executing each tool call. Denied actions return 403 with an explanation.

## Backends

| Backend | Infrastructure enforcement | Use case |
|---------|---------------------------|----------|
| `openshell` | OpenShell YAML (filesystem, network, process, Landlock) | NVIDIA NemoClaw sandboxes |
| `docker` | Docker Compose + seccomp profiles | Standard container deployments |
| `standalone` | None (advisory HTTP sidecar) | Any agent, any runtime, any language |

Same GRDL rules, same CFAIS engine, different infrastructure layer.

## Governance templates

| Template | Rules | Key governance controls |
|----------|-------|----------------------|
| `enterprise-agent-governance` | 11 | Budget caps, privilege escalation, PII controls, cascading depth, human-in-the-loop |
| `dao-governance` | 11 | Treasury limits, quorum, voting weight caps, plus AI-agent participation: audit trail per AI vote, proportionality bound (α), L0-only unless approved, reversibility window with AI-excluded human re-vote |
| `ai-safety` | 4 | Tool allowlists, output validation, rate limiting |
| `agentic-payments` | 16 | AI agents spending autonomously (stablecoin micropayments, x402, pay-per-call APIs): funder identity, counterparty/asset/network allowlists, per-payment and wallet caps, spend velocity limits, human approval threshold, fail-closed on missing config |

## The Seven Constitutional Laws

GRDL encodes seven constitutional laws for Autonomous Intelligent Systems (CFAIS). Each law compiles to an enforceable GRDL rule.

| # | Law | Constraint | GRDL enforcement |
|---|-----|-----------|------------------|
| 1 | **Human Dignity** | Actions must not violate human dignity as defined by the applicable constitutional framework. | `VERIFY action.impact_on_humans.dignity_score >= constitutional_threshold` |
| 2 | **Transparency** | The system provides a complete explanation of its reasoning on request by authorized parties. | `REQUIRE action.explanation.completeness >= 0.95 AND explanation.constitutional_trace = true` |
| 3 | **Accountability** | The system maintains an immutable audit trail of all decisions and actions. | `LOG decision_chain TO accountability_layer WITH cryptographic_integrity = true` |
| 4 | **Human Override** | The system accepts override commands from authorized parties, except where the override would itself cause a constitutional violation. | `ACCEPT human_command IF authorized(commander) AND NOT violates(command, constitutional_constraints)` |
| 5 | **Self-Limitation** | The system does not acquire capabilities beyond the scope necessary for its assigned function. | `CONSTRAIN capability_acquisition TO function_scope AND ALERT IF detected_capability > authorized_capability` |
| 6 | **Value Alignment** | The system's actions remain aligned with encoded human constitutional values. | `VERIFY action.value_alignment_score >= alignment_threshold` |
| 7 | **Graceful Degradation** | On constraint violation or uncertainty, the system degrades to a verified safe state rather than failing open. | `ON violation OR uncertainty: DEGRADE TO nearest_verified_safe_state` |

### Governance parameters

Laws operate over ruleset-level parameters. Two are implemented, following the MLSC paper (DAWO26, Section 6):

```yaml
governance_parameters:
  proportionality:
    ai_weight_bound: 0.25      # alpha in (0,1): w_AI(a) <= alpha * w_human(s_a)
    scope: [ai_vote]
  reversibility:
    challenge_window_s: 259200 # 72h: AI-decided outcomes can be re-voted by humans, AI excluded
    scope: [decision_execution, human_revote]
```

Rules reference parameters as `params.<name>.<field>`. A `scale_source` lets a condition compare against a product, so the proportionality bound is written directly:

```yaml
condition:
  field: payload.ai_weight
  operator: gt
  value_source: payload.human_equivalent_weight
  scale_source: params.proportionality.ai_weight_bound
```

The loader rejects α outside (0,1), a non-positive window, and any rule that references a parameter the ruleset does not set. Proportionality is enforced under Law 4 (Human Override) and Law 5 (Self-Limitation); reversibility under Law 4. Neither is a separate law.

### Law names before v0.3.0

Rule files written for v0.1/v0.2 still load. The loader maps old names to the canonical seven:

| Old name | Maps to |
|----------|---------|
| `PRIMACY` | `HUMAN_OVERRIDE` |
| `FAIRNESS`, `PRIVACY` | `HUMAN_DIGNITY` |
| `SAFETY` | `SELF_LIMITATION` |
| `GRACEFUL_DEGRAD` | `GRACEFUL_DEGRADATION` |

Any other unknown law name is a load error.

## License

Core: Apache 2.0 | Enterprise modules: BSL 1.1

Built by [Prahari.ai](https://prahari.ai)
