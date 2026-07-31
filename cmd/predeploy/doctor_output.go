package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/HXong/predeploy-guard/internal/doctor"
	"github.com/spf13/cobra"
)

var errDoctorChecksFailed = errors.New("doctor checks failed")

func newDoctorCommand() *cobra.Command {
	var configPath string
	var appPath string

	command := &cobra.Command{
		Use:   "doctor",
		Short: "Check whether the local environment is ready for PreDeploy Guard",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := doctor.Run(cmd.Context(), doctor.Options{
				ConfigPath: configPath,
				AppPath:    appPath,
			})
			printDoctorReport(cmd.OutOrStdout(), report)
			if report.HasFailures() {
				cmd.Root().SilenceErrors = true
				cmd.Root().SilenceUsage = true
				return errDoctorChecksFailed
			}
			return nil
		},
	}

	command.Flags().StringVar(
		&configPath,
		"config",
		"",
		"Path to a PreDeploy Guard YAML config",
	)
	command.Flags().StringVar(
		&appPath,
		"app",
		"",
		"Path to an application directory to inspect",
	)

	return command
}

func printDoctorReport(writer io.Writer, report doctor.Report) {
	fmt.Fprintln(writer, "PreDeploy Guard Doctor")

	category := ""
	for _, result := range report.Results {
		if result.Category != category {
			category = result.Category
			fmt.Fprintf(writer, "\n%s\n", category)
		}
		fmt.Fprintf(writer, "[%s] %s\n", result.Status, result.Message)
		if result.Details != "" {
			fmt.Fprintf(writer, "       %s\n", result.Details)
		}
	}

	passed, warned, failed := report.Counts()
	warningLabel := "warnings"
	if warned == 1 {
		warningLabel = "warning"
	}
	fmt.Fprintf(writer, "\nSummary: %d passed, %d %s, %d failed\n",
		passed, warned, warningLabel, failed)
}
