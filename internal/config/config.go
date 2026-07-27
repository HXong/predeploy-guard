package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigPath string `yaml:"-"`
	ConfigDir  string `yaml:"-"`

	Runtime      RuntimeConfig               `yaml:"runtime"`
	Service      ServiceConfig               `yaml:"service"`
	Dependencies map[string]DependencyConfig `yaml:"dependencies"`
	Checks       ChecksConfig                `yaml:"checks"`
	Gateway      GatewayConfig               `yaml:"gateway"`
	Workloads    []WorkloadConfig            `yaml:"workloads"`
	Performance  PerformanceConfig           `yaml:"performance"`
	Settings     SettingsConfig              `yaml:"settings"`

	Profiles      map[string]ProfileConfig `yaml:"profiles"`
	ActiveProfile string                   `yaml:"-"`
}

type RuntimeConfig struct {
	Type    string `yaml:"type"`
	Context string `yaml:"context"`
}

type ServiceConfig struct {
	Name       string            `yaml:"name"`
	Image      string            `yaml:"image"`
	Build      BuildConfig       `yaml:"build"`
	Port       int32             `yaml:"port"`
	HealthPath string            `yaml:"healthPath"`
	Env        map[string]string `yaml:"env"`
}

type BuildConfig struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

type DependencyConfig struct {
	Image     string            `yaml:"image"`
	Port      int               `yaml:"port"`
	Env       map[string]string `yaml:"env"`
	Readiness ReadinessConfig   `yaml:"readiness"`
}

type ReadinessConfig struct {
	Command         []string `yaml:"command"`
	Shell           string   `yaml:"shell"`
	IntervalSeconds int      `yaml:"intervalSeconds"`
	TimeoutSeconds  int      `yaml:"timeoutSeconds"`
}

type ChecksConfig struct {
	Smoke []SmokeCheck `yaml:"smoke"`
}

type SmokeCheck struct {
	Name           string `yaml:"name"`
	Method         string `yaml:"method"`
	Path           string `yaml:"path"`
	ExpectedStatus int    `yaml:"expectedStatus"`
}

type GatewayConfig struct {
	Enabled bool           `yaml:"enabled"`
	BaseURL string         `yaml:"baseURL"`
	Routes  []GatewayRoute `yaml:"routes"`
}

type GatewayRoute struct {
	Name           string `yaml:"name"`
	Method         string `yaml:"method"`
	Path           string `yaml:"path"`
	ExpectedStatus int    `yaml:"expectedStatus"`
	CompareDirect  *bool  `yaml:"compareDirect"`
}

type WorkloadConfig struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
	Enabled        *bool  `yaml:"enabled"`
	Method         string `yaml:"method"`
	Path           string `yaml:"path"`
	Duration       string `yaml:"duration"`
	RatePerSecond  int    `yaml:"ratePerSecond"`
	ExpectedStatus int    `yaml:"expectedStatus"`
	FailurePolicy  string `yaml:"failurePolicy"`
}

type PerformanceConfig struct {
	Enabled    bool                  `yaml:"enabled"`
	VUs        int                   `yaml:"vus"`
	Duration   string                `yaml:"duration"`
	Thresholds PerformanceThresholds `yaml:"thresholds"`
	Endpoints  []PerformanceEndpoint `yaml:"endpoints"`
}

type PerformanceThresholds struct {
	MaxP95LatencyMs float64 `yaml:"maxP95LatencyMs"`
	MaxErrorRate    float64 `yaml:"maxErrorRate"`
}

type PerformanceEndpoint struct {
	Name   string `yaml:"name"`
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type SettingsConfig struct {
	NamespacePrefix string `yaml:"namespacePrefix"`
	Cleanup         bool   `yaml:"cleanup"`
	TimeoutSeconds  int    `yaml:"timeoutSeconds"`
}

type ProfileConfig struct {
	Checks      *ChecksConfig      `yaml:"checks"`
	Performance *PerformanceConfig `yaml:"performance"`
	Settings    *SettingsConfig    `yaml:"settings"`
}

func Load(path string) (*Config, error) {
	return LoadWithProfile(path, "")
}

func LoadWithProfile(path string, profileName string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}

	cfg.ConfigPath = absPath
	cfg.ConfigDir = filepath.Dir(absPath)

	if err := cfg.ResolvePaths(); err != nil {
		return nil, err
	}

	if err := cfg.ApplyProfile(profileName); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Runtime.Type == "" {
		c.Runtime.Type = "docker-compose"
	}

	switch c.Runtime.Type {
	case "docker-compose", "kubernetes":
	default:
		return fmt.Errorf(
			"unsupported runtime.type %q; supported runtimes: docker-compose, kubernetes",
			c.Runtime.Type,
		)
	}
	if c.Service.Name == "" {
		return fmt.Errorf("service.name is required")
	}
	if c.Service.Image == "" {
		return fmt.Errorf("service.image is required")
	}
	if c.Service.Build.Context != "" {
		if c.Service.Build.Dockerfile == "" {
			c.Service.Build.Dockerfile = "Dockerfile"
		}

		if _, err := os.Stat(c.Service.Build.Context); err != nil {
			return fmt.Errorf("service.build.context is invalid: %w", err)
		}
	}
	if c.Service.Port <= 0 {
		return fmt.Errorf("service.port must be greater than 0")
	}
	if c.Service.HealthPath == "" {
		c.Service.HealthPath = "/health"
	}
	for name, dependency := range c.Dependencies {
		if name == "" {
			return fmt.Errorf("dependency name cannot be empty")
		}

		if dependency.Image == "" {
			return fmt.Errorf("dependencies.%s.image is required", name)
		}

		if dependency.Port < 0 {
			return fmt.Errorf("dependencies.%s.port cannot be negative", name)
		}

		hasCommand := len(dependency.Readiness.Command) > 0
		hasShell := dependency.Readiness.Shell != ""

		if hasCommand && hasShell {
			return fmt.Errorf("dependencies.%s.readiness cannot define both command and shell", name)
		}

		if hasCommand || hasShell {
			if dependency.Readiness.IntervalSeconds <= 0 {
				dependency.Readiness.IntervalSeconds = 2
			}

			if dependency.Readiness.TimeoutSeconds <= 0 {
				dependency.Readiness.TimeoutSeconds = 30
			}

			c.Dependencies[name] = dependency
		}
	}
	if c.Gateway.Enabled {
		if c.Gateway.BaseURL == "" {
			return fmt.Errorf("gateway.baseURL is required when gateway is enabled")
		}
		if !strings.HasPrefix(c.Gateway.BaseURL, "http://") &&
			!strings.HasPrefix(c.Gateway.BaseURL, "https://") {
			return fmt.Errorf("gateway.baseURL must start with http:// or https://")
		}
		if len(c.Gateway.Routes) == 0 {
			return fmt.Errorf("gateway.routes must contain at least one route when gateway is enabled")
		}

		routeNames := make(map[string]struct{}, len(c.Gateway.Routes))
		for i := range c.Gateway.Routes {
			route := &c.Gateway.Routes[i]
			if route.Name == "" {
				return fmt.Errorf("gateway.routes[%d].name is required", i)
			}
			if _, exists := routeNames[route.Name]; exists {
				return fmt.Errorf("gateway.routes[%d].name %q must be unique", i, route.Name)
			}
			routeNames[route.Name] = struct{}{}

			if route.Method == "" {
				route.Method = "GET"
			} else {
				route.Method = strings.ToUpper(route.Method)
			}
			if route.Path == "" {
				return fmt.Errorf("gateway.routes[%d].path is required", i)
			}
			if !strings.HasPrefix(route.Path, "/") {
				return fmt.Errorf("gateway.routes[%d].path must start with /", i)
			}
			if route.ExpectedStatus == 0 {
				route.ExpectedStatus = 200
			}
			if route.ExpectedStatus < 100 || route.ExpectedStatus > 599 {
				return fmt.Errorf("gateway.routes[%d].expectedStatus must be between 100 and 599", i)
			}
			if route.CompareDirect == nil {
				route.CompareDirect = boolPointer(true)
			}
		}
	}
	workloadNames := make(map[string]struct{}, len(c.Workloads))
	for i := range c.Workloads {
		workload := &c.Workloads[i]

		if workload.Name == "" {
			return fmt.Errorf("workloads[%d].name is required", i)
		}
		if _, exists := workloadNames[workload.Name]; exists {
			return fmt.Errorf("workloads[%d].name %q must be unique", i, workload.Name)
		}
		workloadNames[workload.Name] = struct{}{}

		if workload.Type == "" {
			return fmt.Errorf("workloads[%d].type is required", i)
		}
		workload.Type = strings.ToLower(workload.Type)
		if workload.Type != "http" {
			return fmt.Errorf(
				"unsupported workloads[%d].type %q; supported workload types: http",
				i,
				workload.Type,
			)
		}

		if workload.Enabled == nil {
			workload.Enabled = boolPointer(true)
		}
		if workload.Method == "" {
			workload.Method = "GET"
		} else {
			workload.Method = strings.ToUpper(workload.Method)
		}
		if workload.Path == "" {
			return fmt.Errorf("workloads[%d].path is required for HTTP workloads", i)
		}
		if !strings.HasPrefix(workload.Path, "/") {
			return fmt.Errorf("workloads[%d].path must start with /", i)
		}

		if workload.Duration == "" {
			workload.Duration = "10s"
		}
		duration, err := time.ParseDuration(workload.Duration)
		if err != nil {
			return fmt.Errorf("workloads[%d].duration must be a valid Go duration: %w", i, err)
		}
		if duration <= 0 {
			return fmt.Errorf("workloads[%d].duration must be greater than 0", i)
		}

		if workload.RatePerSecond == 0 {
			workload.RatePerSecond = 1
		}
		if workload.RatePerSecond < 0 {
			return fmt.Errorf("workloads[%d].ratePerSecond must be greater than 0", i)
		}

		if workload.ExpectedStatus == 0 {
			workload.ExpectedStatus = 200
		}
		if workload.ExpectedStatus < 100 || workload.ExpectedStatus > 599 {
			return fmt.Errorf("workloads[%d].expectedStatus must be between 100 and 599", i)
		}

		if workload.FailurePolicy == "" {
			workload.FailurePolicy = "fail"
		} else {
			workload.FailurePolicy = strings.ToLower(workload.FailurePolicy)
		}
		switch workload.FailurePolicy {
		case "fail", "warn":
		default:
			return fmt.Errorf(
				"unsupported workloads[%d].failurePolicy %q; supported failure policies: fail, warn",
				i,
				workload.FailurePolicy,
			)
		}
	}
	if c.Performance.Enabled {
		if c.Performance.VUs <= 0 {
			c.Performance.VUs = 10
		}
		if c.Performance.Duration == "" {
			c.Performance.Duration = "15s"
		}

		if c.Performance.Thresholds.MaxP95LatencyMs <= 0 {
			c.Performance.Thresholds.MaxP95LatencyMs = 500
		}

		if c.Performance.Thresholds.MaxErrorRate < 0 {
			return fmt.Errorf("performance.thresholds.maxErrorRate cannot be negative")
		}

		if c.Performance.Thresholds.MaxErrorRate == 0 {
			c.Performance.Thresholds.MaxErrorRate = 0.01
		}

		if len(c.Performance.Endpoints) == 0 {
			for _, smoke := range c.Checks.Smoke {
				c.Performance.Endpoints = append(c.Performance.Endpoints, PerformanceEndpoint{
					Name:   smoke.Name,
					Method: smoke.Method,
					Path:   smoke.Path,
				})
			}
		}

		for i, endpoint := range c.Performance.Endpoints {
			if endpoint.Name == "" {
				return fmt.Errorf("performance.endpoints[%d].name is required", i)
			}

			if endpoint.Method == "" {
				c.Performance.Endpoints[i].Method = "GET"
			}

			if endpoint.Path == "" {
				return fmt.Errorf("performance.endpoints[%d].path is required", i)
			}
		}
	}
	if c.Settings.NamespacePrefix == "" {
		c.Settings.NamespacePrefix = "predeploy"
	}
	if c.Settings.TimeoutSeconds <= 0 {
		c.Settings.TimeoutSeconds = 60
	}
	return nil
}

func boolPointer(value bool) *bool {
	return &value
}
