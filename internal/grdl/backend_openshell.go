package grdl

type openshellBackend struct{}

func (o *openshellBackend) Name() string { return "openshell" }

func (o *openshellBackend) SidecarConfig() map[string]interface{} {
	return map[string]interface{}{"host": "127.0.0.1", "port": 9700}
}

func (o *openshellBackend) CompileInfrastructure(infra []Rule, network []Rule, rs *Ruleset) (map[string]interface{}, []string) {
	var warnings []string

	np := map[string]interface{}{
		"cfais_sidecar": map[string]interface{}{
			"name": "cfais-governance-engine",
			"endpoints": []map[string]interface{}{
				{"host": "127.0.0.1", "port": 9700, "protocol": "rest",
					"tls": "passthrough", "enforcement": "enforce", "access": "full"},
			},
			"binaries": []map[string]interface{}{{"path": "/sandbox/**"}},
		},
	}

	return map[string]interface{}{
		"version": 1,
		"filesystem_policy": map[string]interface{}{
			"include_workdir": true,
			"read_only":       []string{"/usr", "/lib", "/proc", "/dev/urandom", "/etc", RulesPath},
			"read_write":      []string{"/sandbox", "/tmp", "/dev/null", AuditPath},
		},
		"process":          map[string]interface{}{"run_as_user": "sandbox", "run_as_group": "sandbox"},
		"landlock":         map[string]interface{}{"compatibility": "best_effort"},
		"network_policies": np,
	}, warnings
}
