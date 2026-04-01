# NVIDIA Inception Application — Prahari.ai

**Prepared for:** NVIDIA Inception Portal (https://programs.nvidia.com/phoenix/application)
**Company:** JKBKSK Private Limited (DBA Prahari.ai)
**Applicant:** Jagan Mohan Kataru, Founder & Chief Architect

---

## Company name

JKBKSK Private Limited (DBA Prahari.ai)

## Website

https://prahari.ai

## One-line description

Deterministic governance policy engine for autonomous AI agents, providing constitutional-level decision controls that complement NVIDIA OpenShell's infrastructure security.

## Product description (what does your product do?)

Prahari is an open-source governance policy engine that adds decision-level controls to AI agents running inside NVIDIA OpenShell sandboxes.

OpenShell solves infrastructure security — what files, networks, and processes an agent can access. Prahari solves the layer above: what decisions an agent should be allowed to make. Should this agent approve a financial transaction exceeding its budget? Can it escalate its own permissions? Is the delegation chain too deep, risking cascading failures?

The core is GRDL (Governance Rule Definition Language), a structured YAML format for expressing governance rules. Rules compile into two outputs: OpenShell-compatible YAML policies for infrastructure enforcement, and a runtime configuration for a CFAIS (Constitutional Framework for Autonomous Intelligent Systems) sidecar that evaluates agent actions in real-time.

The engine is deterministic — mathematical comparisons, set operations, threshold checks. No ML in the governance path. This means governance decisions are formally verifiable, reproducible, and auditable. We have 9 theorems formally verified in Lean 4 covering core governance properties including dominant strategy incentive compatibility and Nash equilibrium stability.

Three governance templates ship out of the box: Enterprise Agent Governance (budget caps, privilege escalation prevention, cascading failure limits, PII access control), DAO/Cooperative Governance (treasury limits, quorum enforcement, voting weight caps), and AI Safety (tool allowlists, output validation, rate limiting). Organizations clone a template and customize for their specific policies.

## How does your product use NVIDIA technology?

Prahari integrates directly with NVIDIA OpenShell as a community sandbox and policy extension:

1. **OpenShell sandbox**: Our Dockerfile and policy.yaml deploy as a community sandbox via `openshell sandbox create --from prahari-constitutional`. The CFAIS sidecar runs inside the sandbox as a governance co-processor.

2. **OpenShell policy compilation**: GRDL rules compile to OpenShell's YAML policy schema (version 1), producing filesystem_policy, network_policies, process, and landlock configurations that merge with existing sandbox policies.

3. **OpenShell network policy routing**: Agent API calls are routed through the CFAIS sidecar via OpenShell's REST protocol inspection, enabling per-request governance evaluation without modifying the agent.

4. **NemoClaw / Agent Toolkit compatibility**: Prahari extends the NemoClaw stack by adding semantic governance on top of OpenShell's infrastructure guardrails. Any agent running in NemoClaw (OpenClaw, Claude Code, Codex, etc.) benefits from governance rules without code changes.

We plan to use NVIDIA Nemotron models for the accessibility layer — voice interfaces and natural language rule authoring — while keeping the core governance engine deterministic and model-free.

## What stage is your product at?

Working prototype. Core engine functional with 13 passing tests, sub-millisecond evaluation latency (0.012ms average), three governance templates, CLI tools, and FastAPI sidecar server. Ready for OpenShell community sandbox PR submission.

## Industry / vertical

Enterprise AI Governance, AI Agent Security, Regulatory Technology

## Technology area

AI Agent Infrastructure, Policy Engines, Formal Verification, Governance Frameworks

## What is your business model?

Open-core with BSL 1.1 for enterprise modules:

- Core engine (GRDL compiler, CFAIS evaluator, sidecar, templates): Apache 2.0, free forever
- Enterprise modules (RITAM multi-law integrity, advanced constitutional auditing, formal verification bridge, regulatory compliance templates): BSL 1.1, free for ≤10 agents in production, commercial pricing above that, converts to Apache 2.0 after 4 years

## Team size

Founder-stage. 1 full-time (founder/architect with 30+ years IT infrastructure experience, CISSP/CISA certified). Active partnership with Camplight/Patio (102-member global worker-owned tech cooperative) for development and deployment.

## Funding

Bootstrapped. Patent portfolio of 23 Indian patent filings covering the core innovations.

## Country

India

## Additional notes

We are submitting a community sandbox PR to the NVIDIA OpenShell repository concurrent with this application. Our work directly addresses the governance gap identified by Futurum Group at GTC 2026: "OpenShell and NemoClaw are a necessary component of agent trust, not a complete solution for it." We are building that complete solution — and we want to build it within the NVIDIA ecosystem.

Our DAWO26 (European DAO Workshop, Zurich) paper on formally verified cooperative governance has been accepted, demonstrating academic validation of the underlying framework.
