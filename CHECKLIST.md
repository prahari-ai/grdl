# Prahari OpenShell — Execution Checklist

## What's ready (delivered in this archive)

### Code (13 tests passing, 0.012ms evaluation latency)
- [x] GRDL schema — universal, jurisdiction-agnostic governance rule definitions
- [x] GRDL loader — YAML parser with validation
- [x] GRDL compiler — three-pass: static → OpenShell YAML, dynamic → network policies, runtime → CFAIS sidecar
- [x] CFAIS engine — deterministic rule evaluator with graceful degradation
- [x] FastAPI sidecar — HTTP server with /evaluate, /healthz, /metrics, /rules
- [x] CLI — compile, validate, evaluate, serve commands
- [x] OpenShell sandbox — Dockerfile, policy.yaml, README.md
- [x] Test suite — 13 tests across 3 governance templates

### Governance templates
- [x] Enterprise Agent Governance — budget caps, privilege escalation, cascading depth, PII, human-in-the-loop
- [x] DAO / Cooperative Governance — treasury limits, quorum, voting weight caps
- [x] AI Safety — tool allowlists, output validation, rate limiting

### Licensing
- [x] LICENSE-APACHE — core engine
- [x] LICENSE-BSL — enterprise modules
- [x] docs/pricing.md — FRAND terms

### Outreach materials
- [x] docs/nvidia-inception-application.md — ready to paste into Inception portal
- [x] docs/openshell-pr-description.md — ready to paste into GitHub PR
- [x] docs/openshell-discussion-post.md — ready to paste into GitHub Discussions
- [x] LinkedIn message to Kari Briski (2 variants drafted)
- [x] Email to Naska Yankova (drafted)

---

## Execution sequence

### Week 1: Foundation

| # | Task | Time | Status |
|---|------|------|--------|
| 1 | Extract archive to VPS, run `python tests/test_pipeline.py` | 30 min | |
| 2 | Create GitHub repo `prahari-ai/prahari-openshell` | 15 min | |
| 3 | Push code to GitHub | 15 min | |
| 4 | Apply for NVIDIA Inception at nvidia.com/en-us/startups/ | 45 min | |
| 5 | Send email to inceptionprogram@nvidia.com referencing application | 15 min | |

### Week 2: Integration testing

| # | Task | Time | Status |
|---|------|------|--------|
| 6 | Install OpenShell CLI on VPS: `curl -fsSL nvidia.com/openshell.sh \| bash` | 30 min | |
| 7 | Build Docker image: `docker build -t prahari-cfais -f core/sandbox/Dockerfile .` | 15 min | |
| 8 | Create OpenShell sandbox manually and test sidecar integration | 2 hrs | |
| 9 | Record 2-minute demo video showing: compile → deploy → deny → allow | 1 hr | |

### Week 3: Publish

| # | Task | Time | Status |
|---|------|------|--------|
| 10 | Fork NVIDIA/OpenShell, add sandbox to `sandboxes/prahari-constitutional/` | 30 min | |
| 11 | Submit PR using `docs/openshell-pr-description.md` | 30 min | |
| 12 | Post Discussion using `docs/openshell-discussion-post.md` | 15 min | |
| 13 | Update prahari.ai website with OpenShell integration page | 2 hrs | |

### Week 4: Outreach

| # | Task | Time | Status |
|---|------|------|--------|
| 14 | LinkedIn message to Kari Briski (AFTER PR is submitted) | 15 min | |
| 15 | Email Naska Yankova with Patio/NVIDIA angle | 15 min | |
| 16 | Weave NemoClaw integration into DAWO26 paper (due April 15) | 4 hrs | |
| 17 | Post on Moltbook agentic-governance submolt | 30 min | |

---

## Key files reference

| File | Purpose |
|------|---------|
| `cli.py` | Main entry point — compile, validate, evaluate, serve |
| `core/grdl_compiler/schema.py` | Type definitions (GovernanceLaw, Rule, Ruleset) |
| `core/grdl_compiler/compiler.py` | GRDL → OpenShell YAML + CFAIS config |
| `core/cfais_engine/engine.py` | Runtime evaluator (12μs per eval) |
| `core/sidecar.py` | FastAPI HTTP server for OpenShell integration |
| `core/sandbox/` | Dockerfile + policy.yaml for OpenShell community catalog |
| `examples/templates/` | Three governance templates ready to customize |
| `tests/test_pipeline.py` | 13-test suite covering all three templates |
| `docs/nvidia-inception-application.md` | Inception portal application text |
| `docs/openshell-pr-description.md` | GitHub PR description |
| `docs/openshell-discussion-post.md` | GitHub Discussion post |
