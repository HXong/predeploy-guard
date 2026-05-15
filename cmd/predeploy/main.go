package main

import (
	"fmt"
	"os"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/history"
	"github.com/HXong/predeploy-guard/internal/runner"
	"github.com/HXong/predeploy-guard/internal/scaffold"
	"github.com/spf13/cobra"
)

func main() {
	var initOutputPath string
	var initForce bool
	var initDependencies string
	var runProfile string
	var validateProfile string
	var explainProfile string

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create a starter predeploy.yaml config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := scaffold.WriteDefaultConfig(scaffold.InitOptions{
				OutputPath:   initOutputPath,
				Overwrite:    initForce,
				Dependencies: initDependencies,
			}); err != nil {
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
			fmt.Printf("  3. Run: predeploy explain %s\n", output)
			fmt.Printf("  4. Run: predeploy run %s\n", output)

			return nil
		},
	}

	initCmd.Flags().StringVarP(
		&initOutputPath,
		"output",
		"o",
		scaffold.DefaultConfigFilename,
		"Output path for generated config",
	)

	initCmd.Flags().BoolVarP(
		&initForce,
		"force",
		"f",
		false,
		"Overwrite existing config file",
	)

	initCmd.Flags().StringVar(
		&initDependencies,
		"with",
		"",
		"Comma-separated dependency presets to include, e.g. postgres,redis",
	)

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

			cfg, err := config.LoadWithProfile(configPath, runProfile)
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

	runCmd.Flags().StringVarP(
		&runProfile,
		"profile",
		"p",
		"",
		"Profile to apply from the config file",
	)

	validateCmd := &cobra.Command{
		Use:   "validate <config-path>",
		Short: "Validate a PreDeploy Guard YAML config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithProfile(args[0], validateProfile)
			if err != nil {
				return err
			}

			printValidationSummary(cfg)
			return nil
		},
	}

	validateCmd.Flags().StringVarP(
		&validateProfile,
		"profile",
		"p",
		"",
		"Profile to apply from the config file",
	)

	explainCmd := &cobra.Command{
		Use:   "explain <config-path>",
		Short: "Explain what a PreDeploy Guard config will do",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadWithProfile(args[0], explainProfile)
			if err != nil {
				return err
			}

			printExplainSummary(cfg)
			return nil
		},
	}

	explainCmd.Flags().StringVarP(
		&explainProfile,
		"profile",
		"p",
		"",
		"Profile to apply from the config file",
	)

	historyCmd := &cobra.Command{
		Use:   "history <config-path>",
		Short: "Show previous PreDeploy Guard runs for a config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return err
			}

			summaries, err := history.ReadHistory(cfg.ConfigDir)
			if err != nil {
				return err
			}

			printHistory(summaries)
			return nil
		},
	}

	showCmd := &cobra.Command{
		Use:   "show <config-path> <run-id>",
		Short: "Show details for a previous PreDeploy Guard run",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(args[0])
			if err != nil {
				return err
			}

			summary, err := history.FindRun(cfg.ConfigDir, args[1])
			if err != nil {
				return err
			}

			printRunSummary(summary)
			return nil
		},
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(showCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
