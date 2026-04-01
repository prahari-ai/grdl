# Prahari — Governance Policy Engine for AI Agents

**Deterministic governance for autonomous AI agents running in [NVIDIA OpenShell](https://github.com/NVIDIA/OpenShell).**

Single static binary. No runtime dependencies. ~5MB memory footprint.

OpenShell controls what agents *can* do (infrastructure).
Prahari controls what agents *should* do (governance).

---

## Build

```bash
# Requires Go 1.22+
go build -o prahari ./cmd/prahari

# Or use make
make build
```

## Quick start

```bash
# Validate a ruleset
./prahari validate examples/templates/enterprise-agent-governance.grdl.yaml

# Compile GRDL → OpenShell YAML + CFAIS config
./prahari compile examples/templates/enterprise-agent-governance.grdl.yaml

# Test an action against the rules
./prahari evaluate examples/templates/enterprise-agent-governance.grdl.yaml examples/test-action-denied.json

# Start the sidecar server
./prahari serve examples/templates/enterprise-agent-governance.grdl.yaml --addr :9700
```

## OpenShell integration

```bash
# Build Docker image (multi-stage, ~15MB final)
make docker

# Or directly in OpenShell
openshell sandbox create --from prahari-constitutional
```

## Governance templates

| Template | Use case |
|----------|----------|
| enterprise-agent-governance | Budget caps, privilege escalation, PII controls, cascading depth |
| dao-governance | Treasury limits, quorum, voting weight caps |
| ai-safety | Tool allowlists, output validation, rate limiting |

Clone a template, edit thresholds, compile, deploy.

## Test

```bash
make test     # unit + integration tests
make bench    # benchmark (ns/op per evaluation)
make smoke    # full CLI smoke test
```

## Architecture

```
Agent → OpenShell Gateway → CFAIS Sidecar (:9700) → Verdict → Target
                                   │
                            GRDL Rules Engine
                          (deterministic, <1μs)
```

The engine is deterministic. Mathematical comparisons and set operations only.
No ML. No probabilistic decisions. Formally verifiable.

## The Seven Governance Laws

| # | Law | Addresses |
|---|-----|-----------|
| 1 | Primacy | Decision authority boundaries, human override |
| 2 | Transparency | Explainability, audit trails, traceability |
| 3 | Accountability | Agent identity, ownership, lifecycle |
| 4 | Fairness | Bias prevention, equitable access |
| 5 | Safety | Cascading failures, privilege limits, budget runaway |
| 6 | Privacy | Data minimisation, PII controls, data sovereignty |
| 7 | Graceful degradation | Deterministic fallback, circuit breaking |

## License

Core (`cmd/`, `internal/`): Apache 2.0
Enterprise (`enterprise/`): BSL 1.1

Built by [Prahari.ai](https://prahari.ai)
