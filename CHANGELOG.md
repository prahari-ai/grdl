# Changelog

## v0.3.0

**Law names now match the CFAIS specification.** The engine uses the seven
constitutional laws: Human Dignity, Transparency, Accountability, Human
Override, Self-Limitation, Value Alignment, Graceful Degradation. Old names
(`PRIMACY`, `FAIRNESS`, `PRIVACY`, `SAFETY`, `GRACEFUL_DEGRAD`) are mapped
automatically, so existing rule files keep working. Unknown law names are now
a load error.

**Governance parameters.** New `governance_parameters` block with
`proportionality` (α, the AI weight bound) and `reversibility` (challenge
window). Rules reference them as `params.*`; the loader validates ranges and
references. New `scale_source` condition field for bounds of the form
`actual > factor × reference`.

**DAO template: AI-agent participation.** Seven new rules implementing the
constraints described in the MLSC paper (DAWO26, §6): complete audit trail per
AI vote, proportionality bound `w_AI ≤ α·w_human`, AI voting limited to L0
without multi-layer approval, execution held during the reversibility window
when AI votes were decisive, and human re-votes that exclude AI participants.

**Fix.** The compiled `governance_laws` map previously marked every law
inactive because rule files used upper-case names. Names are now normalised,
so active laws are reported correctly.

## v0.2.1

Agentic-payments template (16 rules).

## v0.2.0

Multi-backend compiler: openshell, docker, standalone.
