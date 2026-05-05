package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Runtime  RuntimeConfig  `yaml:"runtime"`
	Service  ServiceConfig  `yaml:"service"`
	Checks   ChecksConfig   `yaml:"checks"`
	Settings SettingsConfig `yaml:"settings"`
}

type RuntimeConfig struct {
	Type string `yaml:"type"`
}

type ServiceConfig struct {
	Name       string            `yaml:"name"`
	Image      string            `yaml:"image"`
	Port       int32             `yaml:"port"`
	HealthPath string            `yaml:"healthPath"`
	Env        map[string]string `yaml:"env"`
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
	if c.Service.Port <= 0 {
		return fmt.Errorf("service.port must be greater than 0")
	}
	if c.Service.HealthPath == "" {
		c.Service.HealthPath = "/health"
	}
	if c.Settings.NamespacePrefix == "" {
		c.Settings.NamespacePrefix = "predeploy"
	}
	if c.Settings.TimeoutSeconds <= 0 {
		c.Settings.TimeoutSeconds = 60
	}
	return nil
}
