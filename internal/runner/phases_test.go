package runner

import (
	"testing"

	"github.com/HXong/predeploy-guard/internal/report"
	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
)

func TestAppendSkippedPhasesPreservesOrder(t *testing.T) {
	var phases []report.RunPhase
	names := []string{
		report.PhaseGatewayChecks,
		report.PhaseSmokeChecks,
		report.PhaseExperimentWorkloads,
		report.PhasePerformanceChecks,
	}

	appendSkippedPhases(&phases, names, "earlier phase failed")

	if len(phases) != len(names) {
		t.Fatalf("phase count = %d, want %d", len(phases), len(names))
	}
	for i, phase := range phases {
		if phase.Name != names[i] {
			t.Fatalf("phase[%d].name = %q, want %q", i, phase.Name, names[i])
		}
		if phase.Status != report.PhaseStatusSkipped {
			t.Fatalf("phase[%d].status = %q, want skipped", i, phase.Status)
		}
		if phase.Error != "earlier phase failed" {
			t.Fatalf("phase[%d].error = %q, want skip reason", i, phase.Error)
		}
	}
}

func TestRuntimeEnvironmentSummary(t *testing.T) {
	env := &predeployruntime.Environment{
		Name:            "predeploy-test-abc123",
		WorkDir:         "not-reported",
		BaseURL:         "http://127.0.0.1:8080",
		WorkloadBaseURL: "http://host.docker.internal:8080",
	}

	summary := runtimeEnvironmentSummary("kubernetes", env)

	if summary.Runtime != "kubernetes" ||
		summary.Name != env.Name ||
		summary.BaseURL != env.BaseURL ||
		summary.WorkloadBaseURL != env.WorkloadBaseURL {
		t.Fatalf("runtime environment summary = %#v, want safe environment fields", summary)
	}
}
