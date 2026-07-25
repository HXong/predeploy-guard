package report

import "time"

const (
	PhaseStatusPassed  = "passed"
	PhaseStatusFailed  = "failed"
	PhaseStatusSkipped = "skipped"

	PhaseImageBuild          = "image-build"
	PhaseRuntimePrepare      = "runtime-prepare"
	PhaseRuntimeStart        = "runtime-start"
	PhaseRuntimeReadiness    = "runtime-readiness"
	PhaseServiceReadiness    = "service-readiness"
	PhaseSmokeChecks         = "smoke-checks"
	PhaseExperimentWorkloads = "experiment-workloads"
	PhasePerformanceChecks   = "performance-checks"
	PhaseDiagnostics         = "diagnostics"
)

type RunPhaseRecorder struct {
	name      string
	startedAt time.Time
}

func StartPhase(name string) *RunPhaseRecorder {
	return &RunPhaseRecorder{
		name:      name,
		startedAt: time.Now(),
	}
}

func (r *RunPhaseRecorder) FinishPassed() RunPhase {
	return r.finish(PhaseStatusPassed, "")
}

func (r *RunPhaseRecorder) FinishFailed(errorMessage string) RunPhase {
	return r.finish(PhaseStatusFailed, errorMessage)
}

func (r *RunPhaseRecorder) FinishSkipped(reason string) RunPhase {
	return r.finish(PhaseStatusSkipped, reason)
}

func SkippedPhase(name string, reason string) RunPhase {
	now := time.Now()
	return RunPhase{
		Name:       name,
		Status:     PhaseStatusSkipped,
		StartedAt:  now,
		FinishedAt: now,
		Error:      reason,
	}
}

func (r *RunPhaseRecorder) finish(status string, errorMessage string) RunPhase {
	finishedAt := time.Now()
	return RunPhase{
		Name:       r.name,
		Status:     status,
		StartedAt:  r.startedAt,
		FinishedAt: finishedAt,
		Duration:   finishedAt.Sub(r.startedAt),
		Error:      errorMessage,
	}
}
