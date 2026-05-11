package scaffold

import (
	"fmt"
	"sort"
	"strings"
)

type DependencyPreset struct {
	Name string
	YAML string
	Env  map[string]string
}

func GetDependencyPresets(raw string) ([]DependencyPreset, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	names := strings.Split(raw, ",")
	presets := make([]DependencyPreset, 0, len(names))
	seen := make(map[string]bool)

	for _, name := range names {
		cleanName := strings.ToLower(strings.TrimSpace(name))
		if cleanName == "" {
			continue
		}

		if seen[cleanName] {
			continue
		}
		seen[cleanName] = true

		preset, err := getDependencyPreset(cleanName)
		if err != nil {
			return nil, err
		}

		presets = append(presets, preset)
	}

	sort.Slice(presets, func(i, j int) bool {
		return presets[i].Name < presets[j].Name
	})

	return presets, nil
}

func getDependencyPreset(name string) (DependencyPreset, error) {
	switch name {
	case "postgres", "postgresql":
		return DependencyPreset{
			Name: "postgres",
			YAML: `  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: testdb
    readiness:
      command: ["pg_isready", "-U", "test", "-d", "testdb"]
      intervalSeconds: 2
      timeoutSeconds: 30
`,
			Env: map[string]string{
				"DATABASE_URL": "postgres://test:test@postgres:5432/testdb",
			},
		}, nil

	case "redis":
		return DependencyPreset{
			Name: "redis",
			YAML: `  redis:
    image: redis:7
    port: 6379
    readiness:
      shell: "redis-cli ping | grep PONG"
      intervalSeconds: 2
      timeoutSeconds: 30
`,
			Env: map[string]string{
				"REDIS_URL": "redis://redis:6379",
			},
		}, nil

	default:
		return DependencyPreset{}, fmt.Errorf("unsupported dependency preset: %s", name)
	}
}
