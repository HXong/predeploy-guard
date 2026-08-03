package main

import (
	"fmt"
	"io"

	"github.com/HXong/predeploy-guard/internal/scaffold"
)

func runInteractiveInit(input io.Reader, output io.Writer, initial scaffold.InitOptions) error {
	wizardResult, err := scaffold.RunInitWizard(scaffold.WizardOptions{
		Initial: initial,
		Input:   input,
		Output:  output,
	})
	if err != nil {
		return err
	}
	if wizardResult.Cancelled {
		fmt.Fprintln(output, "Init cancelled. No files were written.")
		return nil
	}

	result, err := scaffold.WriteConfig(wizardResult.Options)
	if err != nil {
		return err
	}
	printInitResult(output, wizardResult.Options, result)
	return nil
}
