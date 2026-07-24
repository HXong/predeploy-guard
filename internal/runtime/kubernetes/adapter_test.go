package kubernetes

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

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
	if err := adapter.Cleanup(context.Background(), env, &cfg, nil); err != nil {
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
	if err := adapter.Cleanup(context.Background(), env, &cfg, nil); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}

	for _, command := range runner.commands {
		if len(command) > 1 && command[0] == "delete" && command[1] == "namespace" {
			t.Fatalf("cleanup attempted namespace deletion after creation failure: %#v", command)
		}
	}
}

func TestAdapterCleanupDeletesNamespaceWhenPortForwardStopFailsOrHangs(t *testing.T) {
	tests := []struct {
		name    string
		process *controlledStopProcess
		timeout time.Duration
	}{
		{
			name: "stop fails",
			process: &controlledStopProcess{
				stopErr: os.ErrPermission,
			},
			timeout: 100 * time.Millisecond,
		},
		{
			name: "stop hangs",
			process: &controlledStopProcess{
				release: make(chan struct{}),
			},
			timeout: 10 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := &recordingKubectlRunner{}
			adapter := newWithKubectlRunner(runner)
			adapter.portForwardStopTimeout = test.timeout

			cfg := testManifestConfig()
			cfg.Runtime.Type = "kubernetes"
			cfg.Settings.NamespacePrefix = "test-prefix"

			env, err := adapter.Prepare(context.Background(), &cfg)
			if err != nil {
				t.Fatalf("Prepare: %v", err)
			}
			workDir := env.WorkDir

			if err := adapter.Start(context.Background(), env, &cfg); err != nil {
				t.Fatalf("Start: %v", err)
			}
			adapter.portForward = test.process

			var progress []string
			cleanupErr := adapter.Cleanup(
				context.Background(),
				env,
				&cfg,
				func(message string) {
					progress = append(progress, message)
				},
			)

			if test.process.release != nil {
				close(test.process.release)
			}
			if cleanupErr == nil {
				t.Fatal("Cleanup error = nil, want port-forward stop error")
			}
			if !strings.Contains(cleanupErr.Error(), "stop kubectl port-forward") {
				t.Fatalf("Cleanup error = %q, want port-forward stop context", cleanupErr)
			}
			if !containsCommandSequence(runner.commands, "delete", "namespace", env.Name) {
				t.Fatalf("namespace deletion was not attempted: %#v", runner.commands)
			}
			if _, err := os.Stat(workDir); !os.IsNotExist(err) {
				t.Fatalf("work directory still exists after cleanup: %s", workDir)
			}

			wantProgress := []string{
				"Stopping kubectl port-forward...",
				"Deleting Kubernetes namespace " + env.Name + "...",
				"Removing runtime sandbox files...",
			}
			if !reflect.DeepEqual(progress, wantProgress) {
				t.Fatalf("cleanup progress = %#v, want %#v", progress, wantProgress)
			}
		})
	}
}

type recordingKubectlRunner struct {
	commands              [][]string
	failCommandContaining string
}

type controlledStopProcess struct {
	stopErr error
	release chan struct{}
}

func (p *controlledStopProcess) Stop() error {
	if p.release != nil {
		<-p.release
	}
	return p.stopErr
}

func (p *controlledStopProcess) Done() <-chan error {
	return nil
}

func (p *controlledStopProcess) Output() string {
	return ""
}

func containsCommandSequence(commands [][]string, sequence ...string) bool {
	for _, command := range commands {
		for index := 0; index+len(sequence) <= len(command); index++ {
			if reflect.DeepEqual(command[index:index+len(sequence)], sequence) {
				return true
			}
		}
	}

	return false
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
