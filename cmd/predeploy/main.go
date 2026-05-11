package main

import (
	"fmt"
	"os"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/runner"
	"github.com/HXong/predeploy-guard/internal/scaffold"
	"github.com/spf13/cobra"
)

func main() {
	var initOutputPath string
	var initForce bool

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create a starter predeploy.yaml config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := scaffold.WriteDefaultConfig(initOutputPath, initForce); err != nil {
				return err
			}

			output := initOutputPath
			if output == "" {
				output = scaffold.DefaultConfigFilename
			}

			fmt.Printf("Created %s\n", output)
			fmt.Println("Next steps:")
			fmt.Printf("  1. Edit %s for your service\n", output)
			fmt.Printf("  2. Run: predeploy validate %s\n", output)
			fmt.Printf("  3. Run: predeploy run %s\n", output)

			return nil
		},
	}

	initCmd.Flags().StringVarP(&initOutputPath, "output", "o", scaffold.DefaultConfigFilename, "Output path for generated config")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing config file")

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

			printValidationSummary(cfg)
			return nil
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
