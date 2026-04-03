package grdl

type standaloneBackend struct{}

func (s *standaloneBackend) Name() string { return "standalone" }

func (s *standaloneBackend) SidecarConfig() map[string]interface{} {
	return map[string]interface{}{"host": "127.0.0.1", "port": 9700}
}

func (s *standaloneBackend) CompileInfrastructure(infra []Rule, network []Rule, rs *Ruleset) (map[string]interface{}, []string) {
	var warnings []string

	if len(infra) > 0 {
		warnings = append(warnings,
			"Standalone backend: infrastructure rules are advisory only. "+
				"Use openshell or docker backend for kernel-level enforcement.")
	}
	if len(network) > 0 {
		warnings = append(warnings,
			"Standalone backend: network rules are advisory only. "+
				"Agent must respect CFAIS sidecar verdicts voluntarily.")
	}

	return map[string]interface{}{
		"mode": "advisory",
		"note": "CFAIS sidecar evaluates rules via HTTP. Agent application must call POST /evaluate before each action and respect the verdict.",
		"sidecar": map[string]interface{}{
			"command": "prahari serve <ruleset.grdl.yaml> --addr :9700",
		},
		"integration": map[string]interface{}{
			"gemma4_ollama":   "Route Gemma 4 function calls through http://localhost:9700/evaluate before execution",
			"langchain":       "Add CFAIS as a tool-call interceptor in your agent chain",
			"openclaw":        "Configure OpenClaw to proxy tool calls via the sidecar",
			"custom":          "HTTP POST to /evaluate with action JSON, check verdict before proceeding",
		},
	}, warnings
}
