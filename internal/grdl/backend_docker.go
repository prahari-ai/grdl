package grdl

type dockerBackend struct{}

func (d *dockerBackend) Name() string { return "docker" }

func (d *dockerBackend) SidecarConfig() map[string]interface{} {
	return map[string]interface{}{"host": "127.0.0.1", "port": 9700}
}

func (d *dockerBackend) CompileInfrastructure(infra []Rule, network []Rule, rs *Ruleset) (map[string]interface{}, []string) {
	var warnings []string

	sidecar := map[string]interface{}{
		"image": "prahari-cfais:latest", "ports": []string{"9700:9700"},
		"volumes":      []string{"./cfais-rules:/sandbox/cfais-rules:ro", "./cfais-audit:/sandbox/cfais-audit"},
		"user":         "1000:1000",
		"read_only":    true,
		"security_opt": []string{"no-new-privileges:true"},
	}

	seccomp := map[string]interface{}{
		"defaultAction": "SCMP_ACT_ALLOW",
		"syscalls": []map[string]interface{}{
			{"names": []string{"mount", "umount2", "pivot_root", "chroot"}, "action": "SCMP_ACT_ERRNO"},
			{"names": []string{"ptrace", "personality", "init_module"}, "action": "SCMP_ACT_ERRNO"},
		},
	}

	warnings = append(warnings, "Docker backend: network-level enforcement requires manual nftables configuration")

	return map[string]interface{}{
		"docker_compose": map[string]interface{}{
			"version": "3.8", "services": map[string]interface{}{"cfais-sidecar": sidecar},
		},
		"seccomp_profile": seccomp,
	}, warnings
}
