# OpenShell Community Sandbox PR Description

**Ready to paste into GitHub when submitting PR to NVIDIA/OpenShell**

---

## PR Title

feat(community): add prahari-constitutional governance sandbox

## Description

Adds a community sandbox that provides deterministic governance policy enforcement for AI agents. While OpenShell controls what agents can access (filesystem, network, process), this sandbox adds a layer for controlling what decisions agents should be allowed to make.

### What it does

A CFAIS (Constitutional Framework for Autonomous Intelligent Systems) sidecar runs inside the sandbox on port 9700. OpenShell's network policies route agent actions through the sidecar, which evaluates them against loaded governance rules and returns allow/deny verdicts with full explanations.

Rules are written in GRDL (Governance Rule Definition Language) — structured YAML that compiles into both OpenShell policy fragments and runtime evaluation configuration. The engine is deterministic: mathematical comparisons and threshold checks only, no ML inference in the governance path.

### Use cases

- **Enterprise agent fleets**: Budget caps, privilege escalation prevention, cascading failure depth limits, PII access controls, human-in-the-loop for high-impact decisions
- **DAO/cooperative governance**: Treasury transaction limits, quorum enforcement, voting weight caps
- **AI safety**: Tool call allowlists, output content validation, action rate limiting

### What's included

```
sandboxes/prahari-constitutional/
├── Dockerfile
├── policy.yaml          # OpenShell baseline policy with CFAIS sidecar routing
├── README.md
└── skills/              # (empty, governance rules loaded via GRDL)
```

### Key design decisions

1. **Out-of-process governance**: Rules live in `/sandbox/cfais-rules` (read-only to the agent via OpenShell filesystem policy). The agent cannot modify its own governance constraints.

2. **OpenShell-native**: GRDL compiles to OpenShell's policy schema v1. The sidecar is reached via standard network policies. No OpenShell modifications required.

3. **Graceful degradation**: If the sidecar fails, the sandbox doesn't break the agent. Configurable fallback: `deny_all` (conservative) or `allow_with_audit` (permissive with logging).

4. **Zero agent changes**: Works with any agent (OpenClaw, Claude Code, Codex, custom). The agent doesn't know it's being governed — it just sees network policy enforcement.

### Performance

- Rule evaluation: 0.012ms average (12 microseconds)
- 13 tests passing across 3 governance templates
- Health check endpoint for OpenShell monitoring

### How to test

```bash
openshell sandbox create --from prahari-constitutional
openshell sandbox connect my-sandbox
# Inside sandbox:
curl http://127.0.0.1:9700/healthz
curl -X POST http://127.0.0.1:9700/evaluate \
  -H "Content-Type: application/json" \
  -d '{"action_id":"test","action_type":"access_control","payload":{"action":"grant_permission"}}'
```

### Related

- Project: https://github.com/prahari-ai/prahari-openshell
- License: Apache 2.0 (this sandbox and core engine)
