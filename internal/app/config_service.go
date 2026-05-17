package app

import "github.com/HXong/predeploy-guard/internal/config"

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) Load(path string) (*config.Config, error) {
	return config.Load(path)
}

func (s *ConfigService) LoadWithProfile(path string, profile string) (*config.Config, error) {
	return config.LoadWithProfile(path, profile)
}
