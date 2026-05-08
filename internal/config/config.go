package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Runtime      RuntimeConfig               `yaml:"runtime"`
	Service      ServiceConfig               `yaml:"service"`
	Dependencies map[string]DependencyConfig `yaml:"dependencies"`
	Checks       ChecksConfig                `yaml:"checks"`
	Performance  PerformanceConfig           `yaml:"performance"`
	Settings     SettingsConfig              `yaml:"settings"`
}

type RuntimeConfig struct {
	Type string `yaml:"type"`
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

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
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
