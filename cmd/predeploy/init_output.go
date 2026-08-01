package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/HXong/predeploy-guard/internal/scaffold"
)

func printInitResult(writer io.Writer, options scaffold.InitOptions, result scaffold.InitResult) {
	output := options.OutputPath
	if output == "" {
		output = scaffold.DefaultConfigFilename
	}

	fmt.Fprintf(writer, "Created %s\n", output)
	if options.Dependencies != "" {
		fmt.Fprintf(writer, "Included dependency presets: %s\n", options.Dependencies)
	}

	if strings.TrimSpace(options.AppPath) == "" {
		printDefaultInitNextSteps(writer, output)
		return
	}

	indicators := strings.Join(result.Detection.ProjectIndicators, ", ")
	if indicators == "" {
		indicators = "none"
	}
	dockerfile := "not found"
	if result.Detection.HasDockerfile {
		dockerfile = "found"
	}

	fmt.Fprintln(writer, "\nDetected application:")
	fmt.Fprintf(writer, "  Path: %s\n", options.AppPath)
	fmt.Fprintf(writer, "  Service name: %s\n", result.ServiceName)
	fmt.Fprintf(writer, "  Project indicators: %s\n", indicators)
	fmt.Fprintf(writer, "  Dockerfile: %s\n", dockerfile)

	fmt.Fprintln(writer, "\nGenerated config:")
	fmt.Fprintf(writer, "  Runtime: %s\n", result.Runtime)
	fmt.Fprintf(writer, "  Image: %s\n", result.Image)
	if result.BuildConfigured {
		fmt.Fprintf(writer, "  Build context: %s\n", result.BuildContext)
	} else if options.NoBuild {
		fmt.Fprintln(writer, "  Build context: not configured (--no-build)")
	} else {
		fmt.Fprintln(writer, "  Build context: not configured")
	}
	fmt.Fprintf(writer, "  Port: %d\n", result.Port)
	fmt.Fprintf(writer, "  Health path: %s\n", result.HealthPath)

	for _, warning := range result.Warnings {
		fmt.Fprintf(writer, "\n[WARN] %s\n", warning.Message)
		if warning.Details != "" {
			fmt.Fprintf(writer, "       %s\n", warning.Details)
		}
	}

	fmt.Fprintln(writer, "\nNext steps:")
	fmt.Fprintf(writer, "  1. Review %s\n", output)
	fmt.Fprintf(writer, "  2. Run: predeploy doctor --config %s --app %s\n", output, options.AppPath)
	fmt.Fprintf(writer, "  3. Run: predeploy validate %s\n", output)
	fmt.Fprintf(writer, "  4. Run: predeploy run %s\n", output)
}

func printDefaultInitNextSteps(writer io.Writer, output string) {
	fmt.Fprintln(writer, "Next steps:")
	fmt.Fprintf(writer, "  1. Edit %s for your service\n", output)
	fmt.Fprintf(writer, "  2. Run: predeploy validate %s\n", output)
	fmt.Fprintf(writer, "  3. Run: predeploy explain %s\n", output)
	fmt.Fprintf(writer, "  4. Run: predeploy run %s\n", output)

	fmt.Fprintln(writer, "Optional profile commands:")
	fmt.Fprintf(writer, "  - Run smoke only: predeploy run %s --profile smoke-only\n", output)
	fmt.Fprintf(writer, "  - Run light load: predeploy run %s --profile light-load\n", output)
	fmt.Fprintf(writer, "  - Run stress test: predeploy run %s --profile stress-test\n", output)

	fmt.Fprintln(writer, "After running:")
	fmt.Fprintf(writer, "  - View history: predeploy history %s\n", output)
	fmt.Fprintf(writer, "  - Show a run: predeploy show %s <run-id>\n", output)
	fmt.Fprintf(writer, "  - Compare runs: predeploy compare %s <base-run-id> <target-run-id>\n", output)
}
