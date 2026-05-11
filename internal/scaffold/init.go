package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

const DefaultConfigFilename = "predeploy.yaml"

const DefaultPredeployYAML = `runtime:
  type: docker-compose

service:
  name: my-service
  image: my-service:predeploy
  build:
    context: .
    dockerfile: Dockerfile
  port: 8080
  healthPath: /health

checks:
  smoke:
    - name: health check
      method: GET
      path: /health
      expectedStatus: 200

performance:
  enabled: true
  vus: 10
  duration: 15s
  thresholds:
    maxP95LatencyMs: 300
    maxErrorRate: 0.01
  endpoints:
    - name: health load
      method: GET
      path: /health

settings:
  cleanup: true
  timeoutSeconds: 60
`

func WriteDefaultConfig(outputPath string, overwrite bool) error {
	if outputPath == "" {
		outputPath = DefaultConfigFilename
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	if !overwrite {
		if _, err := os.Stat(absPath); err == nil {
			return fmt.Errorf("config file already exists: %s; use --force to overwrite", absPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check existing config file: %w", err)
		}
	}

	if err := os.WriteFile(absPath, []byte(DefaultPredeployYAML), 0644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}
