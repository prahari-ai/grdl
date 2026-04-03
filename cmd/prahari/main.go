package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prahari-ai/grdl/internal/cfais"
	"github.com/prahari-ai/grdl/internal/grdl"
	"github.com/prahari-ai/grdl/internal/sidecar"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "compile":
		cmdCompile(os.Args[2:])
	case "validate":
		cmdValidate(os.Args[2:])
	case "evaluate":
		cmdEvaluate(os.Args[2:])
	case "serve":
		cmdServe(os.Args[2:])
	case "version":
		fmt.Println("prahari 0.2.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`prahari — GRDL governance policy engine for AI agents

Usage:
  prahari compile  <ruleset.grdl.yaml> [--backend openshell|docker|standalone] [--output-dir <dir>]
  prahari validate <ruleset.grdl.yaml>
  prahari evaluate <ruleset.grdl.yaml> <action.json>
  prahari serve    <ruleset.grdl.yaml> [--addr :9700]
  prahari version

Backends:
  openshell   NVIDIA OpenShell YAML policies (default)
  docker      Docker Compose + seccomp profiles
  standalone  CFAIS sidecar only, no infrastructure enforcement`)
}

func cmdCompile(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prahari compile <ruleset.grdl.yaml> [--backend openshell|docker|standalone]")
		os.Exit(1)
	}

	outDir := "./compiled"
	backendName := "openshell"
	for i, a := range args[1:] {
		if a == "--output-dir" && i+2 < len(args) {
			outDir = args[i+2]
		}
		if a == "--backend" && i+2 < len(args) {
			backendName = args[i+2]
		}
	}

	rs, err := grdl.LoadRuleset(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	backend, err := grdl.GetBackend(backendName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result := grdl.Compile(rs, backend)

	os.MkdirAll(outDir, 0o755)

	infraName := fmt.Sprintf("%s-policy.yaml", backendName)
	infraYAML, _ := result.ToInfraYAML()
	os.WriteFile(filepath.Join(outDir, infraName), infraYAML, 0o644)

	cfaisYAML, _ := result.ToCFAISYAML()
	os.WriteFile(filepath.Join(outDir, "cfais-runtime.yaml"), cfaisYAML, 0o644)

	fmt.Printf("Compiled %d rules (backend: %s):\n", result.Stats.TotalRules, result.BackendName)
	fmt.Printf("  Infrastructure:  %d\n", result.Stats.InfraRules)
	fmt.Printf("  Network:         %d\n", result.Stats.NetworkRules)
	fmt.Printf("  Runtime (CFAIS): %d\n", result.Stats.RuntimeRules)
	fmt.Printf("  Laws covered:    %v\n", result.Stats.LawsCovered)
	fmt.Printf("\nOutputs:\n  %s/%s\n  %s/cfais-runtime.yaml\n", outDir, infraName, outDir)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}
}

func cmdValidate(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prahari validate <ruleset.grdl.yaml>")
		os.Exit(1)
	}
	rs, err := grdl.LoadRuleset(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "x Invalid: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("v Valid GRDL ruleset: %s\n", rs.Name)
	fmt.Printf("  ID: %s\n", rs.ID)
	fmt.Printf("  Version: %s\n", rs.Version)
	fmt.Printf("  Framework: %s\n", rs.Framework)
	fmt.Printf("  Rules: %d\n", len(rs.Rules))
	for _, r := range rs.Rules {
		fmt.Printf("    - %s [%s] > %s\n", r.ID, r.Severity, r.Target)
	}
}

func cmdEvaluate(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: prahari evaluate <ruleset.grdl.yaml> <action.json>")
		os.Exit(1)
	}
	rs, err := grdl.LoadRuleset(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	backend, _ := grdl.GetBackend("standalone")
	compiled := grdl.Compile(rs, backend)
	engine, err := cfais.NewEngine(compiled.CFAISConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	actionData, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	var ctx cfais.EvaluationContext
	if err := json.Unmarshal(actionData, &ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	result := engine.Evaluate(&ctx)
	output, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(output))
}

func cmdServe(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prahari serve <ruleset.grdl.yaml> [--addr :9700]")
		os.Exit(1)
	}
	addr := ":9700"
	for i, a := range args[1:] {
		if a == "--addr" && i+2 < len(args) {
			addr = args[i+2]
		}
	}
	rs, err := grdl.LoadRuleset(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	backend, _ := grdl.GetBackend("standalone")
	compiled := grdl.Compile(rs, backend)
	engine, err := cfais.NewEngine(compiled.CFAISConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	srv := sidecar.NewServer(engine)
	if err := srv.ListenAndServe(addr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
