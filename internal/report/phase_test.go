package report

import "testing"

func TestSkippedPhase(t *testing.T) {
	phase := SkippedPhase(PhasePerformanceChecks, "performance is disabled")

	if phase.Name != PhasePerformanceChecks {
		t.Fatalf("name = %q, want %q", phase.Name, PhasePerformanceChecks)
	}
	if phase.Status != PhaseStatusSkipped {
		t.Fatalf("status = %q, want %q", phase.Status, PhaseStatusSkipped)
	}
	if phase.Duration != 0 {
		t.Fatalf("duration = %s, want 0", phase.Duration)
	}
	if phase.Error != "performance is disabled" {
		t.Fatalf("error = %q, want skip reason", phase.Error)
	}
	if !phase.StartedAt.Equal(phase.FinishedAt) {
		t.Fatalf("skipped phase timestamps differ: %s != %s", phase.StartedAt, phase.FinishedAt)
	}
}

func TestRunPhaseRecorderFinishesPassed(t *testing.T) {
	phase := StartPhase(PhaseSmokeChecks).FinishPassed()

	if phase.Status != PhaseStatusPassed {
		t.Fatalf("status = %q, want %q", phase.Status, PhaseStatusPassed)
	}
	if phase.Duration < 0 {
		t.Fatalf("duration = %s, want non-negative", phase.Duration)
	}
	if phase.FinishedAt.Before(phase.StartedAt) {
		t.Fatalf("finishedAt %s is before startedAt %s", phase.FinishedAt, phase.StartedAt)
	}
}
