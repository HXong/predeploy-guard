package factory

import (
	"strings"
	"testing"

	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
)

func TestNewAdapterDefaultsToDockerCompose(t *testing.T) {
	adapter, err := NewAdapter("")
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}

	if adapter.Type() != predeployruntime.TypeDockerCompose {
		t.Fatalf("adapter type = %q, want %q", adapter.Type(), predeployruntime.TypeDockerCompose)
	}
}

func TestNewAdapterReturnsDockerCompose(t *testing.T) {
	adapter, err := NewAdapter(string(predeployruntime.TypeDockerCompose))
	if err != nil {
		t.Fatalf("NewAdapter: %v", err)
	}

	if adapter.Type() != predeployruntime.TypeDockerCompose {
		t.Fatalf("adapter type = %q, want %q", adapter.Type(), predeployruntime.TypeDockerCompose)
	}
}

func TestNewAdapterRejectsUnsupportedRuntime(t *testing.T) {
	adapter, err := NewAdapter("kubernetes")
	if err == nil {
		t.Fatal("NewAdapter error = nil, want unsupported runtime error")
	}
	if adapter != nil {
		t.Fatalf("adapter = %#v, want nil", adapter)
	}

	wantParts := []string{
		`unsupported runtime.type "kubernetes"`,
		`supported runtimes: docker-compose`,
	}
	for _, want := range wantParts {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("NewAdapter error = %q, want it to contain %q", err, want)
		}
	}
}
