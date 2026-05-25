package app

import (
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/runner"
)

type RunService struct {
	configPath string
}

type RunValidationResult struct {
	Passed  bool   `json:"passed"`
	Profile string `json:"profile"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewRunService(configPath string) *RunService {
	return &RunService{
		configPath: configPath,
	}
}

func (s *RunService) Run(profile string) (RunValidationResult, error) {
	cfg, err := config.LoadWithProfile(s.configPath, profile)
	if err != nil {
		return RunValidationResult{}, err
	}

	profileName := cfg.ActiveProfile
	if profileName == "" {
		profileName = "base"
	}

	result := RunValidationResult{
		Passed:  true,
		Profile: profileName,
		Message: "validation run completed",
	}

	if err := runner.Run(cfg); err != nil {
		result.Passed = false
		result.Message = ""
		result.Error = err.Error()
	}

	return result, nil
}
