package appdetect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProjectType string

const (
	ProjectUnknown ProjectType = "unknown"
	ProjectDocker  ProjectType = "docker"
	ProjectNode    ProjectType = "node"
	ProjectGo      ProjectType = "go"
	ProjectPython  ProjectType = "python"
	ProjectJava    ProjectType = "java"
)

type DetectionResult struct {
	AppPath           string
	Name              string
	HasDockerfile     bool
	DockerfilePath    string
	ProjectTypes      []ProjectType
	ProjectIndicators []string
}

type Options struct {
	AppPath string
}

type indicator struct {
	filename    string
	projectType ProjectType
}

var supportedIndicators = []indicator{
	{filename: "Dockerfile", projectType: ProjectDocker},
	{filename: "package.json", projectType: ProjectNode},
	{filename: "go.mod", projectType: ProjectGo},
	{filename: "requirements.txt", projectType: ProjectPython},
	{filename: "pyproject.toml", projectType: ProjectPython},
	{filename: "pom.xml", projectType: ProjectJava},
	{filename: "build.gradle", projectType: ProjectJava},
}

func Detect(options Options) (DetectionResult, error) {
	if strings.TrimSpace(options.AppPath) == "" {
		return DetectionResult{}, fmt.Errorf("app path is required")
	}

	appPath, err := filepath.Abs(options.AppPath)
	if err != nil {
		return DetectionResult{}, fmt.Errorf("resolve app path: %w", err)
	}
	appPath = filepath.Clean(appPath)

	info, err := os.Stat(appPath)
	if err != nil {
		return DetectionResult{}, fmt.Errorf("inspect app path %q: %w", options.AppPath, err)
	}
	if !info.IsDir() {
		return DetectionResult{}, fmt.Errorf("app path is not a directory: %s", options.AppPath)
	}

	result := DetectionResult{
		AppPath: appPath,
		Name:    SanitizeServiceName(filepath.Base(appPath)),
	}
	seenTypes := make(map[ProjectType]bool)

	for _, candidate := range supportedIndicators {
		path := filepath.Join(appPath, candidate.filename)
		fileInfo, statErr := os.Stat(path)
		if statErr != nil || fileInfo.IsDir() {
			continue
		}

		result.ProjectIndicators = append(result.ProjectIndicators, candidate.filename)
		if !seenTypes[candidate.projectType] {
			result.ProjectTypes = append(result.ProjectTypes, candidate.projectType)
			seenTypes[candidate.projectType] = true
		}
		if candidate.projectType == ProjectDocker {
			result.HasDockerfile = true
			result.DockerfilePath = path
		}
	}

	if len(result.ProjectTypes) == 0 {
		result.ProjectTypes = []ProjectType{ProjectUnknown}
	}

	return result, nil
}

func SanitizeServiceName(value string) string {
	var builder strings.Builder
	lastWasSeparator := false

	for _, character := range strings.ToLower(strings.TrimSpace(value)) {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			builder.WriteRune(character)
			lastWasSeparator = false
			continue
		}
		if builder.Len() > 0 && !lastWasSeparator {
			builder.WriteByte('-')
			lastWasSeparator = true
		}
	}

	name := strings.Trim(builder.String(), "-")
	if name == "" {
		return "app"
	}
	return name
}
