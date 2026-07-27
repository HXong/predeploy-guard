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
	Gateway           GatewaySummary      `json:"gateway"`
	SmokeChecks       []SmokeCheckSummary `json:"smokeChecks"`
	Workloads         []WorkloadSummary   `json:"workloads"`
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

type GatewaySummary struct {
	Enabled bool                  `json:"enabled"`
	BaseURL string                `json:"baseUrl,omitempty"`
	Routes  []GatewayRouteSummary `json:"routes"`
}

type GatewayRouteSummary struct {
	Name           string `json:"name"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	ExpectedStatus int    `json:"expectedStatus"`
	CompareDirect  bool   `json:"compareDirect"`
}

type WorkloadSummary struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Enabled        bool   `json:"enabled"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	Duration       string `json:"duration"`
	RatePerSecond  int    `json:"ratePerSecond"`
	ExpectedStatus int    `json:"expectedStatus"`
	FailurePolicy  string `json:"failurePolicy"`
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
	GatewayRoutes        int `json:"gatewayRoutes"`
	Workloads            int `json:"workloads"`
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
	gatewayRoutes := s.gatewayRouteSummaries()
	smokeChecks := s.smokeCheckSummaries()
	workloads := s.workloadSummaries()
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
		Gateway: GatewaySummary{
			Enabled: s.cfg.Gateway.Enabled,
			BaseURL: s.cfg.Gateway.BaseURL,
			Routes:  gatewayRoutes,
		},
		SmokeChecks: smokeChecks,
		Workloads:   workloads,
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
			GatewayRoutes:        len(gatewayRoutes),
			Workloads:            len(workloads),
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
			s.explainGatewayChecks(),
			s.explainSmokeChecks(),
			s.explainWorkloads(),
			s.explainPerformance(),
			s.explainReports(),
			s.explainCleanup(),
		},
	}
}

func (s *ConfigService) gatewayRouteSummaries() []GatewayRouteSummary {
	summaries := make([]GatewayRouteSummary, 0, len(s.cfg.Gateway.Routes))
	for _, route := range s.cfg.Gateway.Routes {
		compareDirect := route.CompareDirect == nil || *route.CompareDirect
		summaries = append(summaries, GatewayRouteSummary{
			Name:           route.Name,
			Method:         strings.ToUpper(route.Method),
			Path:           route.Path,
			ExpectedStatus: route.ExpectedStatus,
			CompareDirect:  compareDirect,
		})
	}
	return summaries
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

func (s *ConfigService) workloadSummaries() []WorkloadSummary {
	summaries := make([]WorkloadSummary, 0, len(s.cfg.Workloads))

	for _, workload := range s.cfg.Workloads {
		enabled := workload.Enabled == nil || *workload.Enabled
		summaries = append(summaries, WorkloadSummary{
			Name:           workload.Name,
			Type:           workload.Type,
			Enabled:        enabled,
			Method:         strings.ToUpper(workload.Method),
			Path:           workload.Path,
			Duration:       workload.Duration,
			RatePerSecond:  workload.RatePerSecond,
			ExpectedStatus: workload.ExpectedStatus,
			FailurePolicy:  workload.FailurePolicy,
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
	switch s.cfg.Runtime.Type {
	case "docker-compose":
		detail = "PreDeploy Guard will use Docker Compose to create a temporary local sandbox."
	case "kubernetes":
		if s.cfg.Runtime.Context == "" {
			detail = "PreDeploy Guard will use the current kubeconfig context to create a temporary Kubernetes namespace."
		} else {
			detail = fmt.Sprintf(
				"PreDeploy Guard will use kubeconfig context %q to create a temporary Kubernetes namespace.",
				s.cfg.Runtime.Context,
			)
		}
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
		details = append(details, "No build context is configured, so PreDeploy Guard will use the configured image.")
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
			Number:  6,
			Title:   "Smoke Checks",
			Details: []string{"No smoke checks are configured."},
		}
	}

	details := []string{fmt.Sprintf("%d smoke check(s) will run after service readiness passes.", len(s.cfg.Checks.Smoke))}
	for _, check := range s.smokeCheckSummaries() {
		details = append(details, fmt.Sprintf("%q: %s %s expects HTTP %d.", check.Name, check.Method, check.Path, check.ExpectedStatus))
	}

	return ConfigExplanationStep{
		Number:  6,
		Title:   "Smoke Checks",
		Details: details,
	}
}

func (s *ConfigService) explainGatewayChecks() ConfigExplanationStep {
	if !s.cfg.Gateway.Enabled {
		return ConfigExplanationStep{
			Number:  5,
			Title:   "Gateway Checks",
			Details: []string{"Gateway checks are disabled."},
		}
	}

	details := []string{
		"Gateway checks run after service readiness.",
		fmt.Sprintf("%d gateway route(s) will be checked through %q.", len(s.cfg.Gateway.Routes), s.cfg.Gateway.BaseURL),
		"Gateway checks compare the configured gateway route to the direct service route when compareDirect is enabled.",
	}
	for _, route := range s.gatewayRouteSummaries() {
		details = append(
			details,
			fmt.Sprintf(
				"%q: %s %s expects HTTP %d; compareDirect is %t.",
				route.Name,
				route.Method,
				route.Path,
				route.ExpectedStatus,
				route.CompareDirect,
			),
		)
	}

	return ConfigExplanationStep{
		Number:  5,
		Title:   "Gateway Checks",
		Details: details,
	}
}

func (s *ConfigService) explainPerformance() ConfigExplanationStep {
	if !s.cfg.Performance.Enabled {
		return ConfigExplanationStep{
			Number:  8,
			Title:   "Performance Checks",
			Details: []string{"Performance checks are disabled."},
		}
	}

	details := []string{
		"Dockerized k6 will run after smoke checks and experiment workloads complete.",
		fmt.Sprintf("Virtual users: %d.", s.cfg.Performance.VUs),
		fmt.Sprintf("Duration: %s.", s.cfg.Performance.Duration),
		fmt.Sprintf("Failure threshold: p95 latency must be <= %.2fms.", s.cfg.Performance.Thresholds.MaxP95LatencyMs),
		fmt.Sprintf("Failure threshold: error rate must be <= %.4f.", s.cfg.Performance.Thresholds.MaxErrorRate),
	}

	for _, endpoint := range s.cfg.Performance.Endpoints {
		details = append(details, fmt.Sprintf("k6 endpoint %q: %s %s.", endpoint.Name, strings.ToUpper(endpoint.Method), endpoint.Path))
	}

	return ConfigExplanationStep{
		Number:  8,
		Title:   "Performance Checks",
		Details: details,
	}
}

func (s *ConfigService) explainWorkloads() ConfigExplanationStep {
	if len(s.cfg.Workloads) == 0 {
		return ConfigExplanationStep{
			Number:  7,
			Title:   "Experiment Workloads",
			Details: []string{"No experiment workloads are configured."},
		}
	}

	details := []string{
		fmt.Sprintf("%d experiment workload(s) are configured to run after smoke checks.", len(s.cfg.Workloads)),
	}
	for _, workload := range s.workloadSummaries() {
		state := "enabled"
		if !workload.Enabled {
			state = "disabled"
		}
		details = append(
			details,
			fmt.Sprintf(
				"%q (%s): %s %s at %d request(s)/second for %s; expects HTTP %d; failure policy %s.",
				workload.Name,
				state,
				workload.Method,
				workload.Path,
				workload.RatePerSecond,
				workload.Duration,
				workload.ExpectedStatus,
				workload.FailurePolicy,
			),
		)
	}

	return ConfigExplanationStep{
		Number:  7,
		Title:   "Experiment Workloads",
		Details: details,
	}
}

func (s *ConfigService) explainReports() ConfigExplanationStep {
	reportsDir := filepath.Join(s.cfg.ConfigDir, "reports")

	return ConfigExplanationStep{
		Number: 9,
		Title:  "Reports",
		Details: []string{
			fmt.Sprintf("Markdown and JSON reports will be written under %q.", reportsDir),
			"Reports include a runtime environment summary and phase timeline.",
			"Runtime diagnostics are included when collected for a failed or incomplete run.",
			"Markdown is intended for humans.",
			"JSON is intended for automation and CI/CD.",
		},
	}
}

func (s *ConfigService) explainCleanup() ConfigExplanationStep {
	details := []string{"Cleanup is enabled. The temporary Docker Compose sandbox will be removed after the run."}
	if s.cfg.Runtime.Type == "kubernetes" {
		details = []string{"Cleanup is enabled. The temporary Kubernetes namespace and local runtime files will be removed after the run."}
	}
	if !s.cfg.Settings.Cleanup {
		if s.cfg.Runtime.Type == "kubernetes" {
			details = []string{
				"Cleanup is disabled. The temporary Kubernetes namespace and local runtime files will be kept after the run.",
				"This is useful for debugging but may leave cluster resources, port-forward processes, and files behind.",
			}
		} else {
			details = []string{
				"Cleanup is disabled. The temporary Docker Compose sandbox will be kept after the run.",
				"This is useful for debugging but may leave containers, networks, and files behind.",
			}
		}
	}

	return ConfigExplanationStep{
		Number:  10,
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
