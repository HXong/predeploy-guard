package main

import (
	"fmt"
	"os"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/runner"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "predeploy",
		Short: "PreDeploy Guard validates backend services before deployment",
	}

	runCmd := &cobra.Command{
		Use:   "run <config-file>",
		Short: "Run pre-deployment checks using a YAML config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := args[0]

			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			fmt.Println("Loaded config successfully")
			fmt.Printf("Service: %s\n", cfg.Service.Name)
			fmt.Printf("Image: %s\n", cfg.Service.Image)
			fmt.Printf("Port: %d\n", cfg.Service.Port)
			fmt.Printf("Health path: %s\n", cfg.Service.HealthPath)

			return runner.Run(cfg)
		},
	}

	validateCmd := &cobra.Command{
		Use:   "validate <config-path>",
		Short: "Validate a PreDeploy Guard YAML config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return err
			}

			fmt.Println("Config validation passed")
			fmt.Printf("Runtime: %s\n", cfg.Runtime.Type)
			fmt.Printf("Service: %s\n", cfg.Service.Name)
			fmt.Printf("Image: %s\n", cfg.Service.Image)

			if cfg.Service.Build.Context != "" {
				fmt.Printf("Build context: %s\n", cfg.Service.Build.Context)
				fmt.Printf("Dockerfile: %s\n", cfg.Service.Build.Dockerfile)
			}

			fmt.Printf("Dependencies: %d\n", len(cfg.Dependencies))
			fmt.Printf("Smoke checks: %d\n", len(cfg.Checks.Smoke))
			fmt.Printf("Performance enabled: %t\n", cfg.Performance.Enabled)

			return nil
		},
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
