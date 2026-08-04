package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/doctor"
)

func TestPrintDoctorReportGroupsChecksAndPrintsSummary(t *testing.T) {
	report := doctor.Report{Results: []doctor.CheckResult{
		{
			Category: doctor.CategoryLocalEnvironment,
			Status:   doctor.StatusPass,
			Message:  "Current directory is writable",
		},
		{
			Category: doctor.CategoryDocker,
			Status:   doctor.StatusWarn,
			Message:  "Docker CLI not found",
			Details:  "Doctor never installs tools.",
		},
		{
			Category: doctor.CategoryApplication,
			Status:   doctor.StatusFail,
			Message:  "App path does not exist: ./missing",
		},
	}, Recommendations: []doctor.Recommendation{{
		Message: "Run guided init:",
		Command: "predeploy init --interactive",
	}}}

	var output bytes.Buffer
	printDoctorReport(&output, report)

	for _, expected := range []string{
		"PreDeploy Guard Doctor",
		"Local Environment\n[PASS] Current directory is writable",
		"Docker\n[WARN] Docker CLI not found",
		"Doctor never installs tools.",
		"Application\n[FAIL] App path does not exist: ./missing",
		"Recommendations\n- Run guided init:\n  predeploy init --interactive",
		"Summary: 1 passed, 1 warning, 1 failed",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output = %q, want it to contain %q", output.String(), expected)
		}
	}
}
