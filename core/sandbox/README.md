# prahari-constitutional

**Deterministic governance policy engine for AI agents in OpenShell.**

Bridges the gap between infrastructure security (what agents *can* do)
and governance intelligence (what agents *should* do).

## What it does

- Evaluates agent actions against governance rules before execution
- Blocks actions that violate configured policies (budget caps, privilege escalation, data access)
- Produces transparency reports for every decision (immutable audit trail)
- Degrades gracefully — never silently allows, never crashes the agent

## Architecture

```
Agent -> OpenShell Gateway -> CFAIS Sidecar (:9700) -> Verdict -> Target
                                    |
                             GRDL Rules Engine
                            (deterministic, <1ms)
```

The engine is NOT probabilistic. Mathematical comparisons and rule
evaluation only. No ML inference in the governance path.

## Quick start

```bash
openshell sandbox create --from prahari-constitutional
```

## Governance templates included

- **Enterprise Agent Governance** — Budget caps, privilege escalation prevention,
  cascading failure limits, PII access control, human-in-the-loop for high-impact actions
- **DAO Governance** — Treasury limits, quorum enforcement, voting weight caps
- **AI Safety** — Tool allowlists, output validation, rate limiting

Clone a template and customize for your organization.

## The Seven Governance Laws (CFAIS)

| # | Law | What it governs |
|---|-----|-----------------|
| 1 | Primacy | Human authority overrides agent decisions |
| 2 | Transparency | All decisions explainable and auditable |
| 3 | Accountability | Every action traced to an entity |
| 4 | Fairness | No bias in agent decisions |
| 5 | Safety | Prevent harm; halt on uncertainty |
| 6 | Privacy | Minimum data; respect data sovereignty |
| 7 | Graceful degradation | Deterministic fallback on failure |

## License

Core engine: Apache 2.0 | Enterprise modules: BSL 1.1
