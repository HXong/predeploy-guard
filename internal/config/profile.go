package config

import (
	"fmt"
	"sort"
)

func (c *Config) ApplyProfile(profileName string) error {
	if profileName == "" {
		return nil
	}

	profile, ok := c.Profiles[profileName]
	if !ok {
		return fmt.Errorf("profile %q not found; available profiles: %v", profileName, c.ProfileNames())
	}

	if profile.Checks != nil {
		c.Checks = *profile.Checks
	}

	if profile.Performance != nil {
		c.Performance = *profile.Performance
	}

	if profile.Settings != nil {
		c.Settings = *profile.Settings
	}

	c.ActiveProfile = profileName
	return nil
}

func (c *Config) ProfileNames() []string {
	names := make([]string, 0, len(c.Profiles))

	for name := range c.Profiles {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
