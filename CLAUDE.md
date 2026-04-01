# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
make build          # Compiles ./cmd/prahari → ./prahari binary
make test           # Runs all tests: go test -v ./tests/
make bench          # Benchmarks with memory profiling
make smoke          # Integration smoke test (validate + compile + evaluate all templates)
make docker         # Builds multi-stage Docker image prahari-cfais:0.1.0
make clean          # Removes compiled artifacts
```

Run a single test:
```bash
go test -v -run TestEvaluateDenyPrivilegeEscalation ./tests/
```

CLI usage:
```bash
./prahari validate examples/templates/enterprise-agent-governance.grdl.yaml
./prahari compile examples/templates/enterprise-agent-governance.grdl.yaml [--output-dir dir]
./prahari evaluate examples/templates/enterprise-agent-governance.grdl.yaml examples/test-action-denied.json
./prahari serve examples/templates/enterprise-agent-governance.grdl.yaml [--addr :9700]
```

## Architecture

Prahari is a deterministic governance policy engine for AI agents running in NVIDIA OpenShell. It enforces rules defined in GRDL (Governance Rule Definition Language) YAML files through a sidecar HTTP server.

**Data flow:** Agent action → OpenShell gateway → CFAIS sidecar (`:9700`) → rule evaluation → verdict (ALLOW/DENY/ALLOW_WITH_AUDIT/MODIFY/ESCALATE)

**Four packages, single pipeline:**

- **`cmd/prahari/`** — CLI entry point. Dispatches to validate/compile/evaluate/serve commands. Each command follows: load ruleset → compile → create engine → execute.
- **`internal/grdl/`** — GRDL schema types, YAML loader, and three-pass compiler. The compiler partitions rules by target (static/dynamic/runtime/hybrid), extracts governance laws, and serializes to OpenShell YAML + CFAIS runtime config.
- **`internal/cfais/`** — Deterministic rule evaluation engine (`engine.go`, 549 lines). Uses first-match-deny semantics with O(1) rule lookup by scope. Supports atomic operators (eq, ne, gt, gte, lt, lte, in, not_in, contains, exists, between) and composite logic (all_of, any_of, none_of). Graceful degradation: type mismatches return false, never error.
- **`internal/sidecar/`** — HTTP server wrapping the engine. Endpoints: `POST /evaluate`, `GET /healthz`, `GET /metrics`, `GET /rules`.

**Only external dependency:** `gopkg.in/yaml.v3`. Everything else is stdlib.

## GRDL Rulesets

Three templates in `examples/templates/`:
- `enterprise-agent-governance.grdl.yaml` — 11 rules covering human primacy, transparency, safety, privacy
- `dao-governance.grdl.yaml` — 5 rules for DAO treasury/quorum/voting
- `ai-safety.grdl.yaml` — 4 rules for tool allowlists, output validation, rate limiting

Test action payloads in `examples/test-action-allowed.json` and `examples/test-action-denied.json`.

## Testing

All 13 tests live in `tests/pipeline_test.go` — a single integration test file that covers the full pipeline (load → compile → evaluate). Tests cover all three rulesets, denial paths (privilege escalation, budget overrun, PII access, cascading depth, DAO treasury, tool allowlist), and graceful degradation on type mismatches.

## Key Design Decisions

- **Deterministic, not probabilistic:** No ML in the governance evaluation path. All comparisons are mathematical/logical.
- **First-match-deny semantics:** Rules are evaluated in order; first matching deny rule wins.
- **Graceful degradation:** Governed by the `graceful_degradation` field in rulesets. Type mismatches and missing fields fail safe (condition not met), never panic.
- **Seven Governance Laws:** Primacy, Transparency, Accountability, Fairness, Safety, Privacy, Graceful Degradation — referenced by rules via the `laws` field.
