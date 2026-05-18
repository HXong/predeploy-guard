package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ConfigPath string `yaml:"-" json:"configPath"`
	ConfigDir  string `yaml:"-" json:"configDir"`

	Runtime      RuntimeConfig               `yaml:"runtime"`
	Service      ServiceConfig               `yaml:"service"`
	Dependencies map[string]DependencyConfig `yaml:"dependencies" json:"dependencies"`
	Checks       ChecksConfig                `yaml:"checks" json:"checks"`
	Performance  PerformanceConfig           `yaml:"performance" json:"performance"`
	Settings     SettingsConfig              `yaml:"settings" json:"settings"`

	Profiles      map[string]ProfileConfig `yaml:"profiles" json:"profiles"`
	ActiveProfile string                   `yaml:"-" json:"activeProfile,omitempty"`
}

type RuntimeConfig struct {
	Type string `yaml:"type" json:"type"`
}

type ServiceConfig struct {
	Name       string            `yaml:"name" json:"name"`
	Image      string            `yaml:"image" json:"image"`
	Build      BuildConfig       `yaml:"build" json:"build"`
	Port       int32             `yaml:"port" json:"port"`
	HealthPath string            `yaml:"healthPath" json:"healthPath"`
	Env        map[string]string `yaml:"env" json:"env"`
}

type BuildConfig struct {
	Context    string `yaml:"context" json:"context"`
	Dockerfile string `yaml:"dockerfile" json:"dockerfile"`
}

type DependencyConfig struct {
	Image     string            `yaml:"image" json:"image"`
	Port      int               `yaml:"port" json:"port"`
	Env       map[string]string `yaml:"env" json:"env"`
	Readiness ReadinessConfig   `yaml:"readiness" json:"readiness"`
}

type ReadinessConfig struct {
	Command         []string `yaml:"command" json:"command"`
	Shell           string   `yaml:"shell" json:"shell"`
	IntervalSeconds int      `yaml:"intervalSeconds" json:"intervalSeconds"`
	TimeoutSeconds  int      `yaml:"timeoutSeconds" json:"timeoutSeconds"`
}

type ChecksConfig struct {
	Smoke []SmokeCheck `yaml:"smoke" json:"smoke"`
}

type SmokeCheck struct {
	Name           string `yaml:"name" json:"name"`
	Method         string `yaml:"method" json:"method"`
	Path           string `yaml:"path" json:"path"`
	ExpectedStatus int    `yaml:"expectedStatus" json:"expectedStatus"`
}

type PerformanceConfig struct {
	Enabled    bool                  `yaml:"enabled" json:"enabled"`
	VUs        int                   `yaml:"vus" json:"vus"`
	Duration   string                `yaml:"duration" json:"duration"`
	Thresholds PerformanceThresholds `yaml:"thresholds" json:"thresholds"`
	Endpoints  []PerformanceEndpoint `yaml:"endpoints" json:"endpoints"`
}

type PerformanceThresholds struct {
	MaxP95LatencyMs float64 `yaml:"maxP95LatencyMs" json:"maxP95LatencyMs"`
	MaxErrorRate    float64 `yaml:"maxErrorRate" json:"maxErrorRate"`
}

type PerformanceEndpoint struct {
	Name   string `yaml:"name" json:"name"`
	Method string `yaml:"method" json:"method"`
	Path   string `yaml:"path" json:"path"`
}

type SettingsConfig struct {
	NamespacePrefix string `yaml:"namespacePrefix" json:"namespacePrefix"`
	Cleanup         bool   `yaml:"cleanup" json:"cleanup"`
	TimeoutSeconds  int    `yaml:"timeoutSeconds" json:"timeoutSeconds"`
}

type ProfileConfig struct {
	Checks      *ChecksConfig      `yaml:"checks" json:"checks,omitempty"`
	Performance *PerformanceConfig `yaml:"performance" json:"performance,omitempty"`
	Settings    *SettingsConfig    `yaml:"settings" json:"settings,omitempty"`
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

	if c.Runtime.Type != "docker-compose" {
		return fmt.Errorf("unsupported runtime.type: %s", c.Runtime.Type)
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
