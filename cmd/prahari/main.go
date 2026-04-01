// prahari is the CLI for the Prahari governance policy engine.
//
// Usage:
//
//	prahari compile <ruleset.grdl.yaml> [--output-dir <dir>]
//	prahari validate <ruleset.grdl.yaml>
//	prahari evaluate <ruleset.grdl.yaml> <action.json>
//	prahari serve <ruleset.grdl.yaml> [--addr :9700]
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
		fmt.Println("prahari 0.1.0")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`prahari — governance policy engine for AI agents

Usage:
  prahari compile  <ruleset.grdl.yaml> [--output-dir <dir>]
  prahari validate <ruleset.grdl.yaml>
  prahari evaluate <ruleset.grdl.yaml> <action.json>
  prahari serve    <ruleset.grdl.yaml> [--addr :9700]
  prahari version`)
}

func cmdCompile(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: prahari compile <ruleset.grdl.yaml> [--output-dir <dir>]")
		os.Exit(1)
	}

	outDir := "./compiled"
	for i, a := range args[1:] {
		if a == "--output-dir" && i+2 < len(args) {
			outDir = args[i+2]
		}
	}

	rs, err := grdl.LoadRuleset(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result := grdl.Compile(rs)

	os.MkdirAll(outDir, 0o755)

	osYAML, _ := result.ToOpenShellYAML()
	os.WriteFile(filepath.Join(outDir, "openshell-policy.yaml"), osYAML, 0o644)

	cfaisYAML, _ := result.ToCFAISYAML()
	os.WriteFile(filepath.Join(outDir, "cfais-runtime.yaml"), cfaisYAML, 0o644)

	fmt.Printf("Compiled %d rules:\n", result.Stats.TotalRules)
	fmt.Printf("  Static  (OpenShell YAML):  %d\n", result.Stats.StaticRules)
	fmt.Printf("  Dynamic (network policies): %d\n", result.Stats.DynamicRules)
	fmt.Printf("  Runtime (CFAIS sidecar):    %d\n", result.Stats.RuntimeRules)
	fmt.Printf("  Laws covered: %v\n", result.Stats.LawsCovered)
	fmt.Printf("\nOutputs:\n  %s/openshell-policy.yaml\n  %s/cfais-runtime.yaml\n", outDir, outDir)

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w)
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
		fmt.Fprintf(os.Stderr, "✗ Invalid ruleset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Valid GRDL ruleset: %s\n", rs.Name)
	fmt.Printf("  ID: %s\n", rs.ID)
	fmt.Printf("  Version: %s\n", rs.Version)
	fmt.Printf("  Framework: %s\n", rs.Framework)
	fmt.Printf("  Rules: %d\n", len(rs.Rules))
	for _, r := range rs.Rules {
		fmt.Printf("    - %s [%s] → %s\n", r.ID, r.Severity, r.Target)
	}
}

func cmdEvaluate(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: prahari evaluate <ruleset.grdl.yaml> <action.json>")
		os.Exit(1)
	}

	rs, err := grdl.LoadRuleset(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading ruleset: %v\n", err)
		os.Exit(1)
	}

	compiled := grdl.Compile(rs)
	engine, err := cfais.NewEngine(compiled.CFAISConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating engine: %v\n", err)
		os.Exit(1)
	}

	actionData, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading action: %v\n", err)
		os.Exit(1)
	}

	var ctx cfais.EvaluationContext
	if err := json.Unmarshal(actionData, &ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing action: %v\n", err)
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
		fmt.Fprintf(os.Stderr, "error loading ruleset: %v\n", err)
		os.Exit(1)
	}

	compiled := grdl.Compile(rs)
	engine, err := cfais.NewEngine(compiled.CFAISConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating engine: %v\n", err)
		os.Exit(1)
	}

	srv := sidecar.NewServer(engine)
	if err := srv.ListenAndServe(addr); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
