package app

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/HXong/predeploy-guard/internal/config"
)

type ConfigService struct {
	cfg *config.Config
}

type ConfigSummary struct {
	ConfigPath        string              `json:"configPath"`
	ConfigDir         string              `json:"configDir"`
	ActiveProfile     string              `json:"activeProfile,omitempty"`
	AvailableProfiles []string            `json:"availableProfiles"`
	Runtime           RuntimeSummary      `json:"runtime"`
	Service           ServiceSummary      `json:"service"`
	Build             BuildSummary        `json:"build"`
	Dependencies      []DependencySummary `json:"dependencies"`
	SmokeChecks       []SmokeCheckSummary `json:"smokeChecks"`
	Performance       PerformanceSummary  `json:"performance"`
	Settings          SettingsSummary     `json:"settings"`
	Reports           ReportsSummary      `json:"reports"`
	Counts            ConfigCounts        `json:"counts"`
}

type RuntimeSummary struct {
	Type string `json:"type"`
}

type ServiceSummary struct {
	Name       string `json:"name"`
	Image      string `json:"image"`
	Port       int32  `json:"port"`
	HealthPath string `json:"healthPath"`
	EnvCount   int    `json:"envCount"`
}

type BuildSummary struct {
	Enabled    bool   `json:"enabled"`
	Context    string `json:"context,omitempty"`
	Dockerfile string `json:"dockerfile,omitempty"`
}

type DependencySummary struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Port         int    `json:"port,omitempty"`
	EnvCount     int    `json:"envCount"`
	HasReadiness bool   `json:"hasReadiness"`
	Readiness    string `json:"readiness"`
}

type SmokeCheckSummary struct {
	Name           string `json:"name"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	ExpectedStatus int    `json:"expectedStatus"`
}

type PerformanceSummary struct {
	Enabled         bool    `json:"enabled"`
	VUs             int     `json:"vus,omitempty"`
	Duration        string  `json:"duration,omitempty"`
	MaxP95LatencyMs float64 `json:"maxP95LatencyMs,omitempty"`
	MaxErrorRate    float64 `json:"maxErrorRate,omitempty"`
	EndpointCount   int     `json:"endpointCount"`
}

type SettingsSummary struct {
	NamespacePrefix string `json:"namespacePrefix"`
	Cleanup         bool   `json:"cleanup"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
}

type ReportsSummary struct {
	Directory string `json:"directory"`
	Markdown  bool   `json:"markdown"`
	JSON      bool   `json:"json"`
	History   bool   `json:"history"`
}

type ConfigCounts struct {
	Dependencies         int `json:"dependencies"`
	DependenciesReady    int `json:"dependenciesWithReadiness"`
	SmokeChecks          int `json:"smokeChecks"`
	PerformanceEndpoints int `json:"performanceEndpoints"`
	Profiles             int `json:"profiles"`
}

type ConfigExplanation struct {
	Title string                  `json:"title"`
	Steps []ConfigExplanationStep `json:"steps"`
}

type ConfigExplanationStep struct {
	Number  int      `json:"number"`
	Title   string   `json:"title"`
	Details []string `json:"details"`
}

func NewConfigService(cfg *config.Config) *ConfigService {
	return &ConfigService{
		cfg: cfg,
	}
}

func (s *ConfigService) Load(path string) (*config.Config, error) {
	return config.Load(path)
}

func (s *ConfigService) LoadWithProfile(path string, profile string) (*config.Config, error) {
	return config.LoadWithProfile(path, profile)
}

func (s *ConfigService) Summary() ConfigSummary {
	dependencies := s.dependencySummaries()
	smokeChecks := s.smokeCheckSummaries()
	performanceEndpoints := len(s.cfg.Performance.Endpoints)
	dependenciesWithReadiness := 0

	for _, dependency := range dependencies {
		if dependency.HasReadiness {
			dependenciesWithReadiness++
		}
	}

	return ConfigSummary{
		ConfigPath:        s.cfg.ConfigPath,
		ConfigDir:         s.cfg.ConfigDir,
		ActiveProfile:     s.cfg.ActiveProfile,
		AvailableProfiles: s.cfg.ProfileNames(),
		Runtime: RuntimeSummary{
			Type: s.cfg.Runtime.Type,
		},
		Service: ServiceSummary{
			Name:       s.cfg.Service.Name,
			Image:      s.cfg.Service.Image,
			Port:       s.cfg.Service.Port,
			HealthPath: s.cfg.Service.HealthPath,
			EnvCount:   len(s.cfg.Service.Env),
		},
		Build: BuildSummary{
			Enabled:    s.cfg.Service.Build.Context != "",
			Context:    s.cfg.Service.Build.Context,
			Dockerfile: s.cfg.Service.Build.Dockerfile,
		},
		Dependencies: dependencies,
		SmokeChecks:  smokeChecks,
		Performance: PerformanceSummary{
			Enabled:         s.cfg.Performance.Enabled,
			VUs:             s.cfg.Performance.VUs,
			Duration:        s.cfg.Performance.Duration,
			MaxP95LatencyMs: s.cfg.Performance.Thresholds.MaxP95LatencyMs,
			MaxErrorRate:    s.cfg.Performance.Thresholds.MaxErrorRate,
			EndpointCount:   performanceEndpoints,
		},
		Settings: SettingsSummary{
			NamespacePrefix: s.cfg.Settings.NamespacePrefix,
			Cleanup:         s.cfg.Settings.Cleanup,
			TimeoutSeconds:  s.cfg.Settings.TimeoutSeconds,
		},
		Reports: ReportsSummary{
			Directory: filepath.Join(s.cfg.ConfigDir, "reports"),
			Markdown:  true,
			JSON:      true,
			History:   true,
		},
		Counts: ConfigCounts{
			Dependencies:         len(dependencies),
			DependenciesReady:    dependenciesWithReadiness,
			SmokeChecks:          len(smokeChecks),
			PerformanceEndpoints: performanceEndpoints,
			Profiles:             len(s.cfg.Profiles),
		},
	}
}

func (s *ConfigService) Explain() ConfigExplanation {
	return ConfigExplanation{
		Title: "PreDeploy Guard Execution Plan",
		Steps: []ConfigExplanationStep{
			s.explainProfile(),
			s.explainRuntime(),
			s.explainService(),
			s.explainDependencies(),
			s.explainServiceReadiness(),
			s.explainSmokeChecks(),
			s.explainPerformance(),
			s.explainReports(),
			s.explainCleanup(),
		},
	}
}

func (s *ConfigService) dependencySummaries() []DependencySummary {
	names := make([]string, 0, len(s.cfg.Dependencies))
	for name := range s.cfg.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	summaries := make([]DependencySummary, 0, len(names))
	for _, name := range names {
		dependency := s.cfg.Dependencies[name]
		hasReadiness := dependencyHasReadiness(dependency.Readiness)

		summaries = append(summaries, DependencySummary{
			Name:         name,
			Image:        dependency.Image,
			Port:         dependency.Port,
			EnvCount:     len(dependency.Env),
			HasReadiness: hasReadiness,
			Readiness:    describeReadiness(dependency.Readiness),
		})
	}

	return summaries
}

func (s *ConfigService) smokeCheckSummaries() []SmokeCheckSummary {
	summaries := make([]SmokeCheckSummary, 0, len(s.cfg.Checks.Smoke))

	for _, check := range s.cfg.Checks.Smoke {
		summaries = append(summaries, SmokeCheckSummary{
			Name:           check.Name,
			Method:         strings.ToUpper(check.Method),
			Path:           check.Path,
			ExpectedStatus: check.ExpectedStatus,
		})
	}

	return summaries
}

func (s *ConfigService) explainProfile() ConfigExplanationStep {
	details := []string{"No profile was selected. The base configuration will be used."}
	if s.cfg.ActiveProfile != "" {
		details = []string{fmt.Sprintf("Profile %q was selected and applied to the base configuration.", s.cfg.ActiveProfile)}
	}

	if len(s.cfg.Profiles) > 0 {
		details = append(details, fmt.Sprintf("Available profiles: %s.", strings.Join(s.cfg.ProfileNames(), ", ")))
	}

	return ConfigExplanationStep{
		Number:  0,
		Title:   "Profile",
		Details: details,
	}
}

func (s *ConfigService) explainRuntime() ConfigExplanationStep {
	detail := fmt.Sprintf("PreDeploy Guard will use runtime %q.", s.cfg.Runtime.Type)
	if s.cfg.Runtime.Type == "docker-compose" {
		detail = "PreDeploy Guard will use Docker Compose to create a temporary local sandbox."
	}

	return ConfigExplanationStep{
		Number:  1,
		Title:   "Runtime",
		Details: []string{detail},
	}
}

func (s *ConfigService) explainService() ConfigExplanationStep {
	details := []string{
		fmt.Sprintf("Service %q will be tested.", s.cfg.Service.Name),
		fmt.Sprintf("The service image is %q.", s.cfg.Service.Image),
	}

	if s.cfg.Service.Build.Context != "" {
		details = append(details, fmt.Sprintf("PreDeploy Guard will build the image from %q using %q.", s.cfg.Service.Build.Context, s.cfg.Service.Build.Dockerfile))
	} else {
		details = append(details, "No build context is configured, so PreDeploy Guard will use the existing Docker image.")
	}

	details = append(details, fmt.Sprintf("The service listens on container port %d.", s.cfg.Service.Port))

	if len(s.cfg.Service.Env) > 0 {
		envKeys := sortedMapKeys(s.cfg.Service.Env)
		for _, key := range envKeys {
			details = append(details, fmt.Sprintf("Environment variable %s=%s will be passed to the service.", key, maskIfSensitive(key, s.cfg.Service.Env[key])))
		}
	}

	return ConfigExplanationStep{
		Number:  2,
		Title:   "Service",
		Details: details,
	}
}

func (s *ConfigService) explainDependencies() ConfigExplanationStep {
	if len(s.cfg.Dependencies) == 0 {
		return ConfigExplanationStep{
			Number:  3,
			Title:   "Dependencies",
			Details: []string{"No dependency services are configured."},
		}
	}

	details := []string{"Dependency services will be started before validating the target service."}
	for _, dependency := range s.dependencySummaries() {
		description := fmt.Sprintf("%q uses image %q.", dependency.Name, dependency.Image)
		if dependency.Port > 0 {
			description = fmt.Sprintf("%q uses image %q on internal port %d.", dependency.Name, dependency.Image, dependency.Port)
		}

		details = append(details, description)
		details = append(details, fmt.Sprintf("%q readiness: %s.", dependency.Name, dependency.Readiness))
	}

	return ConfigExplanationStep{
		Number:  3,
		Title:   "Dependencies",
		Details: details,
	}
}

func (s *ConfigService) explainServiceReadiness() ConfigExplanationStep {
	return ConfigExplanationStep{
		Number: 4,
		Title:  "Service Readiness",
		Details: []string{
			fmt.Sprintf("PreDeploy Guard will wait for the service health endpoint %q to become reachable.", s.cfg.Service.HealthPath),
			fmt.Sprintf("Timeout: %d seconds.", s.cfg.Settings.TimeoutSeconds),
		},
	}
}

func (s *ConfigService) explainSmokeChecks() ConfigExplanationStep {
	if len(s.cfg.Checks.Smoke) == 0 {
		return ConfigExplanationStep{
			Number:  5,
			Title:   "Smoke Checks",
			Details: []string{"No smoke checks are configured."},
		}
	}

	details := []string{fmt.Sprintf("%d smoke check(s) will run after service readiness passes.", len(s.cfg.Checks.Smoke))}
	for _, check := range s.smokeCheckSummaries() {
		details = append(details, fmt.Sprintf("%q: %s %s expects HTTP %d.", check.Name, check.Method, check.Path, check.ExpectedStatus))
	}

	return ConfigExplanationStep{
		Number:  5,
		Title:   "Smoke Checks",
		Details: details,
	}
}

func (s *ConfigService) explainPerformance() ConfigExplanationStep {
	if !s.cfg.Performance.Enabled {
		return ConfigExplanationStep{
			Number:  6,
			Title:   "Performance Checks",
			Details: []string{"Performance checks are disabled."},
		}
	}

	details := []string{
		"Dockerized k6 will run after all smoke checks pass.",
		fmt.Sprintf("Virtual users: %d.", s.cfg.Performance.VUs),
		fmt.Sprintf("Duration: %s.", s.cfg.Performance.Duration),
		fmt.Sprintf("Failure threshold: p95 latency must be <= %.2fms.", s.cfg.Performance.Thresholds.MaxP95LatencyMs),
		fmt.Sprintf("Failure threshold: error rate must be <= %.4f.", s.cfg.Performance.Thresholds.MaxErrorRate),
	}

	for _, endpoint := range s.cfg.Performance.Endpoints {
		details = append(details, fmt.Sprintf("k6 endpoint %q: %s %s.", endpoint.Name, strings.ToUpper(endpoint.Method), endpoint.Path))
	}

	return ConfigExplanationStep{
		Number:  6,
		Title:   "Performance Checks",
		Details: details,
	}
}

func (s *ConfigService) explainReports() ConfigExplanationStep {
	reportsDir := filepath.Join(s.cfg.ConfigDir, "reports")

	return ConfigExplanationStep{
		Number: 7,
		Title:  "Reports",
		Details: []string{
			fmt.Sprintf("Markdown and JSON reports will be written under %q.", reportsDir),
			"Markdown is intended for humans.",
			"JSON is intended for automation and CI/CD.",
		},
	}
}

func (s *ConfigService) explainCleanup() ConfigExplanationStep {
	details := []string{"Cleanup is enabled. The temporary Docker Compose sandbox will be removed after the run."}
	if !s.cfg.Settings.Cleanup {
		details = []string{
			"Cleanup is disabled. The temporary Docker Compose sandbox will be kept after the run.",
			"This is useful for debugging but may leave containers, networks, and files behind.",
		}
	}

	return ConfigExplanationStep{
		Number:  8,
		Title:   "Cleanup",
		Details: details,
	}
}

func dependencyHasReadiness(readiness config.ReadinessConfig) bool {
	return len(readiness.Command) > 0 || readiness.Shell != ""
}

func describeReadiness(readiness config.ReadinessConfig) string {
	if len(readiness.Command) > 0 {
		return fmt.Sprintf("command %q", strings.Join(readiness.Command, " "))
	}

	if readiness.Shell != "" {
		return fmt.Sprintf("shell %q", readiness.Shell)
	}

	return "not configured"
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func maskIfSensitive(key string, value string) string {
	upperKey := strings.ToUpper(key)
	sensitiveWords := []string{
		"PASSWORD",
		"SECRET",
		"TOKEN",
		"KEY",
	}

	for _, word := range sensitiveWords {
		if strings.Contains(upperKey, word) {
			return "******"
		}
	}

	return value
}
