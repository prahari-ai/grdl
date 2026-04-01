# GitHub Discussion for NVIDIA/OpenShell

**Category:** Ideas
**Title:** Governance policy engine — decision-level controls for agents

---

## Context

OpenShell provides excellent infrastructure security: filesystem isolation, network policies, process controls, Landlock enforcement. These answer the question: **what can this agent access?**

But enterprises deploying agents are increasingly asking a different question: **what decisions should this agent be allowed to make?** The top concerns from McKinsey, Gartner, and CSA research this year:

- Budget and resource runaway (agents spending beyond limits)
- Privilege escalation (agents granting themselves permissions)
- Cascading failures (delegation chains propagating bad decisions)
- Missing audit trails (who authorized what, on whose behalf)
- Human-in-the-loop gaps (high-impact decisions without approval)

These are semantic governance concerns that YAML network/filesystem policies can't express.

## Proposal

I've built a governance sidecar that runs inside OpenShell sandboxes and evaluates agent actions against structured rules at runtime. Key properties:

- **Compiles to OpenShell-native YAML** — governance rules produce `policy.yaml` fragments that merge with existing sandbox policies
- **Sidecar architecture** — runs on 127.0.0.1:9700, reached via standard OpenShell network policies. Agents don't need modification.
- **Deterministic** — mathematical comparisons and threshold checks only. No ML in the governance path. Formally verifiable.
- **12 microsecond evaluation** — well within any reasonable latency budget
- **Graceful degradation** — configurable fallback if sidecar fails (deny-all or allow-with-audit). Never breaks the agent.

Rules are written in GRDL (Governance Rule Definition Language):

```yaml
- id: safety.budget_cap
  name: Budget enforcement
  laws: [SAFETY]
  scope: resource_consumption
  condition:
    field: payload.estimated_cost
    operator: gt
    value_source: context.remaining_budget
  severity: critical
  target: runtime
  remedy:
    action: block
    message: Cost exceeds remaining budget
    escalation: operator
```

Three template rulesets included: enterprise agent governance, DAO governance, AI safety.

## Questions for the OpenShell team

1. **Is a governance sidecar pattern something you'd want in the community catalog?** I've submitted a sandbox PR with Dockerfile + policy.yaml.

2. **Would you consider a `governance_policies` section in the policy schema?** Currently GRDL rules that need runtime evaluation are routed through network policies to the sidecar. A native governance section in the schema would make this cleaner.

3. **Interest in a `generate-governance-policy` agent skill?** Similar to the existing `generate-sandbox-policy` skill in `.agents/skills/`, but for governance rules.

Repo: https://github.com/prahari-ai/prahari-openshell
