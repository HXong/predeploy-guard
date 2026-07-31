package doctor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
)

type Status string

const (
	StatusPass Status = "PASS"
	StatusWarn Status = "WARN"
	StatusFail Status = "FAIL"
)

const commandTimeout = 10 * time.Second

const (
	CategoryLocalEnvironment = "Local Environment"
	CategoryDocker           = "Docker"
	CategoryKubernetes       = "Kubernetes"
	CategoryApplication      = "Application"
)

type CheckResult struct {
	Category string
	Name     string
	Status   Status
	Message  string
	Details  string
}

type Report struct {
	Results []CheckResult
}

type Options struct {
	ConfigPath string
	AppPath    string
	WorkingDir string
}

type CommandRunner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type execCommandRunner struct{}

func (execCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (execCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

type requirements struct {
	dockerCLI     bool
	dockerDaemon  bool
	dockerCompose bool
	kubectl       bool
	cluster       bool
}

func Run(ctx context.Context, options Options) Report {
	return RunWithRunner(ctx, options, execCommandRunner{})
}

func RunWithRunner(ctx context.Context, options Options, commandRunner CommandRunner) Report {
	report := Report{}
	workingDir := options.WorkingDir
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			report.add(CategoryLocalEnvironment, "working-directory", StatusFail,
				"Current directory unavailable", err.Error())
			return report
		}
	}

	workingDir, err := filepath.Abs(workingDir)
	if err != nil {
		report.add(CategoryLocalEnvironment, "working-directory", StatusFail,
			"Current directory unavailable", err.Error())
		return report
	}

	report.checkWritable(workingDir)
	report.checkGitRepository(workingDir)

	loadedConfig := report.checkConfig(options.ConfigPath, workingDir)
	required := configRequirements(loadedConfig)
	report.checkDocker(ctx, commandRunner, required)
	report.checkKubernetes(ctx, commandRunner, required)

	if options.AppPath != "" {
		report.checkAppPath(options.AppPath, workingDir)
	}

	return report
}

func (r *Report) Counts() (passed int, warned int, failed int) {
	for _, result := range r.Results {
		switch result.Status {
		case StatusPass:
			passed++
		case StatusWarn:
			warned++
		case StatusFail:
			failed++
		}
	}
	return passed, warned, failed
}

func (r *Report) HasFailures() bool {
	_, _, failed := r.Counts()
	return failed > 0
}

func (r *Report) add(category string, name string, status Status, message string, details string) {
	r.Results = append(r.Results, CheckResult{
		Category: category,
		Name:     name,
		Status:   status,
		Message:  message,
		Details:  details,
	})
}

func (r *Report) checkWritable(workingDir string) {
	file, err := os.CreateTemp(workingDir, ".predeploy-doctor-*")
	if err != nil {
		r.add(CategoryLocalEnvironment, "current-directory-writable", StatusFail,
			"Current directory is not writable", err.Error())
		return
	}

	name := file.Name()
	closeErr := file.Close()
	removeErr := os.Remove(name)
	if closeErr != nil || removeErr != nil {
		r.add(CategoryLocalEnvironment, "current-directory-writable", StatusFail,
			"Current directory write check could not clean up its temporary file",
			errors.Join(closeErr, removeErr).Error())
		return
	}

	r.add(CategoryLocalEnvironment, "current-directory-writable", StatusPass,
		"Current directory is writable", "")
}

func (r *Report) checkGitRepository(workingDir string) {
	current := workingDir
	for {
		info, err := os.Stat(filepath.Join(current, ".git"))
		if err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			r.add(CategoryLocalEnvironment, "git-repository", StatusPass,
				"Git repository detected", "")
			return
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	r.add(CategoryLocalEnvironment, "git-repository", StatusWarn,
		"No Git repository detected in the current directory or its parents",
		"Git is optional, but a repository makes configuration and reports easier to track.")
}

func (r *Report) checkConfig(configPath string, workingDir string) *config.Config {
	supplied := configPath != ""
	if configPath == "" {
		configPath = "predeploy.yaml"
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(workingDir, configPath)
	}
	configPath = filepath.Clean(configPath)

	_, err := os.Stat(configPath)
	if err != nil {
		if !supplied && errors.Is(err, os.ErrNotExist) {
			r.add(CategoryLocalEnvironment, "config", StatusWarn,
				"No predeploy.yaml found in the current directory",
				"Pass --config to check a config at another path.")
			return nil
		}
		r.add(CategoryLocalEnvironment, "config", StatusFail,
			fmt.Sprintf("Config validation failed: %v", err), "")
		return nil
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		r.add(CategoryLocalEnvironment, "config", StatusFail,
			fmt.Sprintf("Config validation failed: %v", err), "")
		return nil
	}

	r.add(CategoryLocalEnvironment, "config", StatusPass,
		fmt.Sprintf("Configuration valid: %s", displayPath(configPath, workingDir)), "")
	r.add(CategoryLocalEnvironment, "configured-runtime", StatusPass,
		fmt.Sprintf("Configured runtime: %s", cfg.Runtime.Type), "")

	if cfg.Gateway.Ingress.Enabled {
		r.add(CategoryLocalEnvironment, "ingress-management", StatusPass,
			"Generated ingress is enabled",
			"PreDeploy Guard checks kubectl and cluster access but does not install or manage an ingress controller.")
	}
	if cfg.Performance.Enabled {
		r.add(CategoryLocalEnvironment, "performance-engine", StatusPass,
			"Dockerized performance checks are enabled",
			"PreDeploy Guard will require the Docker CLI and daemon for k6; doctor does not run k6.")
	}

	return cfg
}

func configRequirements(cfg *config.Config) requirements {
	if cfg == nil {
		return requirements{}
	}

	required := requirements{}
	switch cfg.Runtime.Type {
	case "docker-compose":
		required.dockerCLI = true
		required.dockerDaemon = true
		required.dockerCompose = true
	case "kubernetes":
		required.kubectl = true
		required.cluster = true
	}

	if cfg.Gateway.Ingress.Enabled {
		required.kubectl = true
		required.cluster = true
	}
	if cfg.Performance.Enabled {
		required.dockerCLI = true
		required.dockerDaemon = true
	}

	return required
}

func (r *Report) checkDocker(ctx context.Context, commandRunner CommandRunner, required requirements) {
	path, err := commandRunner.LookPath("docker")
	if err != nil {
		r.add(CategoryDocker, "docker-cli", requiredStatus(required.dockerCLI),
			requiredMessage("Docker CLI not found", required.dockerCLI),
			"Install Docker yourself if the selected workflow needs it; doctor never installs tools.")
		r.add(CategoryDocker, "docker-daemon", requiredStatus(required.dockerDaemon),
			requiredMessage("Docker daemon could not be checked because the Docker CLI is missing", required.dockerDaemon), "")
		r.add(CategoryDocker, "docker-compose", requiredStatus(required.dockerCompose),
			requiredMessage("Docker Compose could not be checked because the Docker CLI is missing", required.dockerCompose), "")
		return
	}

	r.add(CategoryDocker, "docker-cli", StatusPass,
		fmt.Sprintf("Docker CLI found: %s", path), "")

	output, err := runCommand(ctx, commandRunner, "docker", "info")
	if err != nil {
		r.add(CategoryDocker, "docker-daemon", requiredStatus(required.dockerDaemon),
			requiredMessage("Docker daemon is not reachable", required.dockerDaemon), commandDetails(output, err))
	} else {
		r.add(CategoryDocker, "docker-daemon", StatusPass,
			"Docker daemon is reachable", "")
	}

	output, err = runCommand(ctx, commandRunner, "docker", "compose", "version")
	if err != nil {
		r.add(CategoryDocker, "docker-compose", requiredStatus(required.dockerCompose),
			requiredMessage("Docker Compose is unavailable", required.dockerCompose), commandDetails(output, err))
	} else {
		r.add(CategoryDocker, "docker-compose", StatusPass,
			"Docker Compose is available", "")
	}
}

func (r *Report) checkKubernetes(ctx context.Context, commandRunner CommandRunner, required requirements) {
	path, err := commandRunner.LookPath("kubectl")
	if err != nil {
		r.add(CategoryKubernetes, "kubectl", requiredStatus(required.kubectl),
			requiredMessage("kubectl not found", required.kubectl),
			"Install kubectl yourself if the selected workflow needs it; doctor never installs tools.")
		r.add(CategoryKubernetes, "kubernetes-context", requiredStatus(required.cluster),
			requiredMessage("Kubernetes context could not be checked because kubectl is missing", required.cluster), "")
		r.add(CategoryKubernetes, "kubernetes-cluster", requiredStatus(required.cluster),
			requiredMessage("Kubernetes cluster reachability could not be checked because kubectl is missing", required.cluster), "")
	} else {
		r.add(CategoryKubernetes, "kubectl", StatusPass,
			fmt.Sprintf("kubectl found: %s", path), "")

		output, runErr := runCommand(ctx, commandRunner, "kubectl", "config", "current-context")
		contextName := strings.TrimSpace(output)
		if runErr != nil || contextName == "" {
			r.add(CategoryKubernetes, "kubernetes-context", requiredStatus(required.cluster),
				requiredMessage("No Kubernetes context is available", required.cluster), commandDetails(output, runErr))
		} else {
			r.add(CategoryKubernetes, "kubernetes-context", StatusPass,
				fmt.Sprintf("Kubernetes context available: %s", firstLine(contextName)), "")
		}

		output, runErr = runCommand(ctx, commandRunner, "kubectl", "cluster-info")
		if runErr != nil {
			r.add(CategoryKubernetes, "kubernetes-cluster", requiredStatus(required.cluster),
				requiredMessage("Kubernetes cluster is not reachable", required.cluster), commandDetails(output, runErr))
		} else {
			r.add(CategoryKubernetes, "kubernetes-cluster", StatusPass,
				"Kubernetes cluster is reachable", "")
		}
	}

	path, err = commandRunner.LookPath("minikube")
	if err != nil {
		r.add(CategoryKubernetes, "minikube", StatusWarn,
			"Minikube not found; it is optional when another kubeconfig context is used",
			"PreDeploy Guard supports existing kubeconfig contexts and does not install or start Minikube.")
		return
	}
	r.add(CategoryKubernetes, "minikube", StatusPass,
		fmt.Sprintf("Minikube found: %s", path),
		"Minikube is optional; any accessible kubeconfig context can be used.")
}

func (r *Report) checkAppPath(appPath string, workingDir string) {
	resolvedPath := appPath
	if !filepath.IsAbs(resolvedPath) {
		resolvedPath = filepath.Join(workingDir, resolvedPath)
	}
	resolvedPath = filepath.Clean(resolvedPath)

	info, err := os.Stat(resolvedPath)
	if err != nil {
		r.add(CategoryApplication, "app-path", StatusFail,
			fmt.Sprintf("App path does not exist: %s", appPath), err.Error())
		return
	}
	if !info.IsDir() {
		r.add(CategoryApplication, "app-path", StatusFail,
			fmt.Sprintf("App path is not a directory: %s", appPath), "")
		return
	}

	r.add(CategoryApplication, "app-path", StatusPass,
		fmt.Sprintf("App path exists: %s", appPath), "")

	indicatorNames := []string{
		"Dockerfile",
		"package.json",
		"go.mod",
		"requirements.txt",
		"pyproject.toml",
		"pom.xml",
		"build.gradle",
	}
	found := make([]string, 0, len(indicatorNames))
	for _, name := range indicatorNames {
		if fileExists(filepath.Join(resolvedPath, name)) {
			found = append(found, name)
		}
	}

	if len(found) == 0 {
		r.add(CategoryApplication, "project-indicators", StatusWarn,
			"No supported project indicators found",
			"Doctor only checks common top-level files and does not perform deep framework detection.")
	} else {
		r.add(CategoryApplication, "project-indicators", StatusPass,
			fmt.Sprintf("Project indicators found: %s", strings.Join(found, ", ")), "")
	}

	if fileExists(filepath.Join(resolvedPath, "Dockerfile")) {
		r.add(CategoryApplication, "dockerfile", StatusPass,
			"Dockerfile found", "")
	} else {
		r.add(CategoryApplication, "dockerfile", StatusWarn,
			"Dockerfile not found",
			"PreDeploy Guard can still reference an existing image, but build-context integration needs a Dockerfile.")
	}
}

func requiredStatus(required bool) Status {
	if required {
		return StatusFail
	}
	return StatusWarn
}

func runCommand(ctx context.Context, commandRunner CommandRunner, name string, args ...string) (string, error) {
	commandCtx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	return commandRunner.Run(commandCtx, name, args...)
}

func requiredMessage(message string, required bool) string {
	if required {
		return message + "; required by the loaded configuration"
	}
	return message
}

func commandDetails(output string, err error) string {
	output = firstLine(output)
	switch {
	case output != "" && err != nil:
		return fmt.Sprintf("%s (%v)", output, err)
	case output != "":
		return output
	case err != nil:
		return err.Error()
	default:
		return ""
	}
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.IndexAny(value, "\r\n"); index >= 0 {
		return value[:index]
	}
	return value
}

func displayPath(path string, workingDir string) string {
	relative, err := filepath.Rel(workingDir, path)
	if err == nil && relative != "." && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return relative
	}
	return path
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
