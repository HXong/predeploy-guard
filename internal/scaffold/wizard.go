package scaffold

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/HXong/predeploy-guard/internal/appdetect"
)

type WizardOptions struct {
	Initial InitOptions
	Input   io.Reader
	Output  io.Writer
}

type WizardResult struct {
	Options   InitOptions
	Cancelled bool
}

type initWizard struct {
	reader *bufio.Reader
	output io.Writer
}

func RunInitWizard(options WizardOptions) (WizardResult, error) {
	if options.Input == nil {
		return WizardResult{}, fmt.Errorf("wizard input is required")
	}
	if options.Output == nil {
		return WizardResult{}, fmt.Errorf("wizard output is required")
	}

	wizard := initWizard{
		reader: bufio.NewReader(options.Input),
		output: options.Output,
	}
	selected := options.Initial
	selected.Guided = true

	fmt.Fprintln(wizard.output, "PreDeploy Guard guided init")
	fmt.Fprintln(wizard.output)

	appPath, err := wizard.promptAppPath(selected.AppPath)
	if err != nil {
		return WizardResult{}, err
	}
	selected.AppPath = appPath

	var detection appdetect.DetectionResult
	if strings.TrimSpace(appPath) != "" {
		detection, err = appdetect.Detect(appdetect.Options{AppPath: appPath})
		if err != nil {
			return WizardResult{}, err
		}
		printDetectionSummary(wizard.output, appPath, detection)
	}

	serviceDefault := "my-service"
	if detection.Name != "" {
		serviceDefault = detection.Name
	}
	if strings.TrimSpace(selected.ServiceName) != "" {
		serviceDefault = appdetect.SanitizeServiceName(selected.ServiceName)
	}
	serviceName, err := wizard.promptDefault("Service name", serviceDefault)
	if err != nil {
		return WizardResult{}, err
	}
	selected.ServiceName = appdetect.SanitizeServiceName(serviceName)

	runtimeDefault := strings.ToLower(strings.TrimSpace(selected.Runtime))
	if runtimeDefault == "" {
		runtimeDefault = "docker-compose"
	}
	selected.Runtime, err = wizard.promptRuntime(runtimeDefault)
	if err != nil {
		return WizardResult{}, err
	}

	imageDefault := strings.TrimSpace(selected.Image)
	if imageDefault == "" {
		if strings.TrimSpace(appPath) == "" {
			imageDefault = "my-service:predeploy"
		} else {
			imageDefault = fmt.Sprintf("predeploy-%s:local", selected.ServiceName)
		}
	}
	selected.Image, err = wizard.promptDefault("Image", imageDefault)
	if err != nil {
		return WizardResult{}, err
	}

	portDefault := selected.Port
	if portDefault == 0 {
		portDefault = 8080
	}
	selected.Port, err = wizard.promptPort(portDefault)
	if err != nil {
		return WizardResult{}, err
	}

	healthDefault := strings.TrimSpace(selected.HealthPath)
	if healthDefault == "" {
		healthDefault = "/health"
	}
	selected.HealthPath, err = wizard.promptHealthPath(healthDefault)
	if err != nil {
		return WizardResult{}, err
	}

	if strings.TrimSpace(appPath) != "" {
		if detection.HasDockerfile {
			useBuild, promptErr := wizard.promptYesNo(
				"Use Dockerfile build context?",
				!selected.NoBuild,
			)
			if promptErr != nil {
				return WizardResult{}, promptErr
			}
			selected.NoBuild = !useBuild
		} else {
			selected.NoBuild = true
			fmt.Fprintln(wizard.output, "Dockerfile not found. Generated config will reference an image only.")
		}
	}

	selected.Dependencies, err = wizard.promptDependencies(selected.Dependencies)
	if err != nil {
		return WizardResult{}, err
	}

	outputDefault := strings.TrimSpace(selected.OutputPath)
	if outputDefault == "" {
		outputDefault = DefaultConfigFilename
	}
	selected.OutputPath, err = wizard.promptDefault("Output file", outputDefault)
	if err != nil {
		return WizardResult{}, err
	}

	printWizardPreview(wizard.output, selected, detection)
	confirmed, err := wizard.promptYesNo("Create this config?", true)
	if err != nil {
		return WizardResult{}, err
	}
	if !confirmed {
		return WizardResult{Options: selected, Cancelled: true}, nil
	}

	return WizardResult{Options: selected}, nil
}

func (w initWizard) promptAppPath(defaultValue string) (string, error) {
	prompt := "Application directory [leave blank to use generic starter config]: "
	if strings.TrimSpace(defaultValue) != "" {
		prompt = fmt.Sprintf("Application directory [%s]: ", defaultValue)
	}
	answer, err := w.readAnswer(prompt)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(answer) == "" {
		return defaultValue, nil
	}
	return strings.TrimSpace(answer), nil
}

func (w initWizard) promptDefault(label string, defaultValue string) (string, error) {
	answer, err := w.readAnswer(fmt.Sprintf("%s [%s]: ", label, defaultValue))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(answer) == "" {
		return defaultValue, nil
	}
	return strings.TrimSpace(answer), nil
}

func (w initWizard) promptRuntime(defaultValue string) (string, error) {
	for {
		answer, err := w.promptDefault("Runtime", defaultValue)
		if err != nil {
			return "", err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "docker-compose" || answer == "kubernetes" {
			return answer, nil
		}
		fmt.Fprintln(w.output, "Please enter docker-compose or kubernetes.")
	}
}

func (w initWizard) promptPort(defaultValue int) (int, error) {
	for {
		answer, err := w.promptDefault("Port", strconv.Itoa(defaultValue))
		if err != nil {
			return 0, err
		}
		port, parseErr := strconv.Atoi(answer)
		if parseErr == nil && port >= 1 && port <= 65535 {
			return port, nil
		}
		fmt.Fprintln(w.output, "Please enter a port from 1 to 65535.")
	}
}

func (w initWizard) promptHealthPath(defaultValue string) (string, error) {
	for {
		answer, err := w.promptDefault("Health path", defaultValue)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(answer, "/") {
			return answer, nil
		}
		fmt.Fprintln(w.output, "Please enter a health path that starts with /.")
	}
}

func (w initWizard) promptDependencies(defaultValue string) (string, error) {
	canonicalDefault, err := normalizeDependencySelection(defaultValue)
	if err != nil {
		canonicalDefault = strings.TrimSpace(defaultValue)
	}
	displayDefault := canonicalDefault
	if displayDefault == "" {
		displayDefault = "none"
	}

	for {
		answer, readErr := w.promptDefault("Dependency presets", displayDefault)
		if readErr != nil {
			return "", readErr
		}
		canonical, validationErr := normalizeDependencySelection(answer)
		if validationErr == nil {
			return canonical, nil
		}
		fmt.Fprintln(w.output, "Please enter none, postgres, redis, or postgres,redis.")
	}
}

func (w initWizard) promptYesNo(label string, defaultYes bool) (bool, error) {
	suffix := "[Y/n]"
	if !defaultYes {
		suffix = "[y/N]"
	}
	for {
		answer, err := w.readAnswer(fmt.Sprintf("%s %s: ", label, suffix))
		if err != nil {
			return false, err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer == "" {
			return defaultYes, nil
		}
		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(w.output, "Please enter yes or no.")
		}
	}
}

func (w initWizard) readAnswer(prompt string) (string, error) {
	if _, err := fmt.Fprint(w.output, prompt); err != nil {
		return "", err
	}
	answer, err := w.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if errors.Is(err, io.EOF) && answer == "" {
		return "", fmt.Errorf("guided init input ended before setup was complete")
	}
	return strings.TrimRight(answer, "\r\n"), nil
}

func normalizeDependencySelection(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "none" {
		return "", nil
	}
	presets, err := GetDependencyPresets(value)
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(presets))
	for _, preset := range presets {
		names = append(names, preset.Name)
	}
	return strings.Join(names, ","), nil
}

func printDetectionSummary(
	output io.Writer,
	displayPath string,
	detection appdetect.DetectionResult,
) {
	indicators := strings.Join(detection.ProjectIndicators, ", ")
	if indicators == "" {
		indicators = "none"
	}
	dockerfile := "not found"
	if detection.HasDockerfile {
		dockerfile = "found"
	}

	fmt.Fprintln(output, "Detected application:")
	fmt.Fprintf(output, "  Path: %s\n", displayPath)
	fmt.Fprintf(output, "  Suggested service name: %s\n", detection.Name)
	fmt.Fprintf(output, "  Project indicators: %s\n", indicators)
	fmt.Fprintf(output, "  Dockerfile: %s\n\n", dockerfile)
}

func printWizardPreview(
	output io.Writer,
	options InitOptions,
	detection appdetect.DetectionResult,
) {
	buildContext := "not configured"
	if strings.TrimSpace(options.AppPath) == "" {
		if !options.NoBuild {
			buildContext = "."
		}
	} else if detection.HasDockerfile && !options.NoBuild {
		outputPath, err := filepath.Abs(options.OutputPath)
		if err == nil {
			buildContext = relativeBuildContext(filepath.Dir(outputPath), detection.AppPath)
		}
	}
	dependencies := options.Dependencies
	if dependencies == "" {
		dependencies = "none"
	}

	fmt.Fprintln(output, "\nGenerated config preview:")
	fmt.Fprintf(output, "  Output: %s\n", options.OutputPath)
	fmt.Fprintf(output, "  Runtime: %s\n", options.Runtime)
	fmt.Fprintf(output, "  Service: %s\n", options.ServiceName)
	fmt.Fprintf(output, "  Image: %s\n", options.Image)
	fmt.Fprintf(output, "  Build context: %s\n", buildContext)
	fmt.Fprintf(output, "  Port: %d\n", options.Port)
	fmt.Fprintf(output, "  Health path: %s\n", options.HealthPath)
	fmt.Fprintf(output, "  Dependencies: %s\n\n", dependencies)
}
