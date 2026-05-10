package config

import "path/filepath"

func (c *Config) ResolvePaths() error {
	if c.Service.Build.Context == "" {
		return nil
	}

	if filepath.IsAbs(c.Service.Build.Context) {
		c.Service.Build.Context = filepath.Clean(c.Service.Build.Context)
		return nil
	}

	resolved, err := filepath.Abs(filepath.Join(c.ConfigDir, c.Service.Build.Context))
	if err != nil {
		return err
	}

	c.Service.Build.Context = resolved
	return nil
}
