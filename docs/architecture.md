# Architecture

## Design principles

1. **Determinism over probabilism**: Core governance logic uses mathematical comparisons, set operations, and game-theoretic verification. AI/ML is used ONLY for accessibility (voice interfaces, NLP). Never for governance decisions.

2. **Out-of-process enforcement**: Constitutional rules cannot be overridden by the agent. The CFAIS sidecar runs in the same sandbox but is protected by OpenShell's filesystem policy (agent cannot modify rules or engine).

3. **Graceful degradation**: If the CFAIS engine fails, the system never silently allows. It falls back to a configurable policy (deny-all or allow-with-audit) and logs the failure.

4. **Transparency by default**: Every evaluation produces a complete explanation citing specific constitutional articles, laws invoked, and alternative suggestions. Law 2 mandates it.

## Data flow

```
1. Agent initiates action (HTTP request to external API)
2. OpenShell gateway intercepts the request
3. Network policy routes request through CFAIS sidecar (127.0.0.1:9700)
4. Sidecar evaluates action against loaded GRDL rules
5. Verdict returned: ALLOW / DENY / ALLOW_WITH_AUDIT / MODIFY / ESCALATE
6. If ALLOW: gateway forwards request to target
7. If DENY: gateway returns 403 with constitutional explanation
8. Audit log entry appended (immutable, append-only)
```

## Formal verification

9 MLSC theorems formally verified in Lean 4 via Goedel-Prover, including DSIC (Dominant Strategy Incentive Compatibility) and Nash Equilibrium properties.
