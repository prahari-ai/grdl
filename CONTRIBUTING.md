# Contributing to Prahari OpenShell

Thank you for your interest in constitutional governance for AI agents.

## How to contribute

### GRDL rulesets

The most valuable contributions are new GRDL rulesets for different
jurisdictions and domains:

- EU AI Act compliance rules
- GDPR data protection rules
- US labour law (FLSA, OSHA)
- Financial services regulations (RBI, SEC, FCA)
- Healthcare compliance (HIPAA, DPDP Act)

Place new rulesets in `examples/` with the naming convention:
`{jurisdiction}-{domain}.grdl.yaml`

### Code contributions

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run the test suite: `python cli.py validate examples/*.grdl.yaml`
5. Submit a pull request

### Reporting issues

Open an issue on GitHub with:
- GRDL ruleset that triggered the issue
- Input action JSON
- Expected vs actual result
- Python version and OS

## Code of conduct

We follow the Contributor Covenant Code of Conduct. Constitutional
governance tools must themselves be governed constitutionally.

## Licensing

Contributions to `core/` are under Apache 2.0.
Contributions to `enterprise/` are under BSL 1.1.
By submitting a PR, you agree to license your contribution accordingly.
