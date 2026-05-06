package runner

import (
	"fmt"

	"github.com/HXong/predeploy-guard/internal/sandbox"
)

func collectLogsSafely(sb *sandbox.ComposeSandbox) string {
	fmt.Println("Collecting container logs...")

	logs, err := sb.Logs()
	if err != nil {
		return fmt.Sprintf("Failed to collect logs: %v\n\nPartial output:\n%s", err, logs)
	}

	return logs
}
