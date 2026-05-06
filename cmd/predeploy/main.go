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

	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
