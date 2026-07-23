package kubernetes

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestAdapterStartAndCleanupUseOwnedNamespace(t *testing.T) {
	runner := &recordingKubectlRunner{}
	adapter := newWithKubectlRunner(runner)
	cfg := testManifestConfig()
	cfg.Runtime = config.RuntimeConfig{
		Type:    "kubernetes",
		Context: "kind-local",
	}
	cfg.Settings.NamespacePrefix = "test-prefix"

	env, err := adapter.Prepare(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	workDir := env.WorkDir

	if err := adapter.Start(context.Background(), env, &cfg); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := adapter.Cleanup(context.Background(), env, &cfg); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	if _, err := os.Stat(workDir); !os.IsNotExist(err) {
		t.Fatalf("work directory still exists after cleanup: %s", workDir)
	}

	wantCommands := [][]string{
		{
			"--context",
			"kind-local",
			"create",
			"namespace",
			env.Name,
		},
		{
			"--context",
			"kind-local",
			"label",
			"namespace",
			env.Name,
			"app.kubernetes.io/managed-by=predeploy-guard",
			"app.kubernetes.io/part-of=predeploy-guard",
			"predeploy.guard/run=" + env.Name,
			"--overwrite",
		},
		{
			"--context",
			"kind-local",
			"apply",
			"--namespace",
			env.Name,
			"-f",
			workDir + string(os.PathSeparator) + manifestFileName,
		},
		{
			"--context",
			"kind-local",
			"delete",
			"namespace",
			env.Name,
			"--ignore-not-found=true",
			"--wait=true",
			"--timeout=60s",
		},
	}
	if !reflect.DeepEqual(runner.commands, wantCommands) {
		t.Fatalf("kubectl commands = %#v, want %#v", runner.commands, wantCommands)
	}
}

func TestAdapterCleanupDoesNotDeleteNamespaceWhenCreationFails(t *testing.T) {
	runner := &recordingKubectlRunner{failCommandContaining: "create namespace"}
	adapter := newWithKubectlRunner(runner)
	cfg := testManifestConfig()
	cfg.Runtime.Type = "kubernetes"
	cfg.Settings.NamespacePrefix = "test-prefix"

	env, err := adapter.Prepare(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	if err := adapter.Start(context.Background(), env, &cfg); err == nil {
		t.Fatal("Start error = nil, want namespace creation failure")
	}
	if err := adapter.Cleanup(context.Background(), env, &cfg); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	for _, command := range runner.commands {
		if len(command) > 1 && command[0] == "delete" && command[1] == "namespace" {
			t.Fatalf("cleanup attempted namespace deletion after creation failure: %#v", command)
		}
	}
}

type recordingKubectlRunner struct {
	commands              [][]string
	failCommandContaining string
}

func (r *recordingKubectlRunner) Run(
	_ context.Context,
	args ...string,
) (string, error) {
	copiedArgs := append([]string(nil), args...)
	r.commands = append(r.commands, copiedArgs)

	if r.failCommandContaining != "" &&
		strings.Contains(strings.Join(args, " "), r.failCommandContaining) {
		return "simulated failure", os.ErrNotExist
	}

	return "", nil
}

func (r *recordingKubectlRunner) Start(
	_ context.Context,
	_ ...string,
) (runningProcess, error) {
	return nil, os.ErrInvalid
}
