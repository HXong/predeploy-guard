package kubernetes

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
)

type kubectlRunner interface {
	Run(ctx context.Context, args ...string) (string, error)
	Start(ctx context.Context, args ...string) (runningProcess, error)
}

type runningProcess interface {
	Stop() error
	Done() <-chan error
	Output() string
}

type execKubectlRunner struct{}

func (execKubectlRunner) Run(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, "kubectl", args...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func (execKubectlRunner) Start(ctx context.Context, args ...string) (runningProcess, error) {
	output := &lockedBuffer{}
	command := exec.CommandContext(ctx, "kubectl", args...)
	command.Stdout = output
	command.Stderr = output

	if err := command.Start(); err != nil {
		return nil, err
	}

	process := &execRunningProcess{
		command: command,
		done:    make(chan error, 1),
		output:  output,
	}
	go func() {
		process.done <- command.Wait()
		close(process.done)
	}()

	return process, nil
}

type execRunningProcess struct {
	command *exec.Cmd
	done    chan error
	output  *lockedBuffer
}

func (p *execRunningProcess) Stop() error {
	select {
	case <-p.done:
		return nil
	default:
	}

	if p.command.Process == nil {
		return nil
	}

	err := p.command.Process.Kill()
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	<-p.done
	return nil
}

func (p *execRunningProcess) Done() <-chan error {
	return p.done
}

func (p *execRunningProcess) Output() string {
	return p.output.String()
}

type lockedBuffer struct {
	mutex sync.Mutex
	data  bytes.Buffer
}

func (b *lockedBuffer) Write(value []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.data.Write(value)
}

func (b *lockedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.data.String()
}

func kubectlArgs(contextName string, args ...string) []string {
	result := make([]string, 0, len(args)+2)
	if contextName != "" {
		result = append(result, "--context", contextName)
	}

	return append(result, args...)
}
