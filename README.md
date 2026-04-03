# GRDL — Governance Rule Definition Language

**The YAML standard for AI agent governance. Runtime-agnostic. Deterministic. Sub-microsecond.**

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

The CFAIS engine is deterministic — mathematical comparisons, not ML. Sub-microsecond evaluation. Single static binary.

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
| `dao-governance` | 4 | Treasury limits, quorum, voting weight caps |
| `ai-safety` | 4 | Tool allowlists, output validation, rate limiting |

## The Seven Governance Laws

| # | Law | Addresses |
|---|-----|-----------|
| 1 | Primacy | Human authority overrides agent decisions |
| 2 | Transparency | All decisions explainable and auditable |
| 3 | Accountability | Every action traced to an entity |
| 4 | Fairness | No bias in agent decisions |
| 5 | Safety | Cascading failures, privilege limits, budget runaway |
| 6 | Privacy | Data minimisation, PII controls |
| 7 | Graceful degradation | Deterministic fallback on failure |

## License

Core: Apache 2.0 | Enterprise modules: BSL 1.1

Built by [Prahari.ai](https://prahari.ai)
