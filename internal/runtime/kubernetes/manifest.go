package kubernetes

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/HXong/predeploy-guard/internal/config"
	"gopkg.in/yaml.v3"
)

type objectMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels"`
}

type deploymentManifest struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   objectMetadata         `yaml:"metadata"`
	Spec       deploymentManifestSpec `yaml:"spec"`
}

type deploymentManifestSpec struct {
	Replicas int32         `yaml:"replicas"`
	Selector labelSelector `yaml:"selector"`
	Template podTemplate   `yaml:"template"`
}

type labelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

type podTemplate struct {
	Metadata podMetadata `yaml:"metadata"`
	Spec     podSpec     `yaml:"spec"`
}

type podMetadata struct {
	Labels map[string]string `yaml:"labels"`
}

type podSpec struct {
	Containers []containerManifest `yaml:"containers"`
}

type containerManifest struct {
	Name           string             `yaml:"name"`
	Image          string             `yaml:"image"`
	Ports          []containerPort    `yaml:"ports,omitempty"`
	Env            []environmentValue `yaml:"env,omitempty"`
	ReadinessProbe *readinessProbe    `yaml:"readinessProbe,omitempty"`
}

type containerPort struct {
	ContainerPort int32 `yaml:"containerPort"`
}

type environmentValue struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type readinessProbe struct {
	Exec           *execAction    `yaml:"exec,omitempty"`
	HTTPGet        *httpGetAction `yaml:"httpGet,omitempty"`
	PeriodSeconds  int            `yaml:"periodSeconds,omitempty"`
	TimeoutSeconds int            `yaml:"timeoutSeconds,omitempty"`
}

type execAction struct {
	Command []string `yaml:"command"`
}

type httpGetAction struct {
	Path string `yaml:"path"`
	Port int32  `yaml:"port"`
}

type serviceManifest struct {
	APIVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   objectMetadata      `yaml:"metadata"`
	Spec       serviceManifestSpec `yaml:"spec"`
}

type serviceManifestSpec struct {
	Selector map[string]string `yaml:"selector"`
	Ports    []servicePort     `yaml:"ports"`
}

type servicePort struct {
	Port       int32 `yaml:"port"`
	TargetPort int32 `yaml:"targetPort"`
}

func generateManifests(cfg *config.Config, namespace string) ([]byte, error) {
	serviceName := sanitizeDNS1123(cfg.Service.Name)
	usedDeploymentNames := map[string]string{
		serviceName: "service",
	}

	documents := make([]interface{}, 0, 2+len(cfg.Dependencies)*2)
	for _, dependencyName := range sortedDependencyNames(cfg.Dependencies) {
		resourceName := sanitizeDNS1123(dependencyName)
		if owner, exists := usedDeploymentNames[resourceName]; exists {
			return nil, fmt.Errorf(
				"Kubernetes resource name %q for dependency %q conflicts with %s",
				resourceName,
				dependencyName,
				owner,
			)
		}
		usedDeploymentNames[resourceName] = fmt.Sprintf("dependency %q", dependencyName)

		dependency := cfg.Dependencies[dependencyName]
		documents = append(documents, dependencyDeployment(namespace, resourceName, dependency))
		if dependency.Port > 0 {
			documents = append(documents, newServiceManifest(
				namespace,
				resourceName,
				int32(dependency.Port),
			))
		}
	}

	documents = append(documents, serviceDeployment(namespace, serviceName, cfg.Service))
	documents = append(documents, newServiceManifest(namespace, serviceName, cfg.Service.Port))

	var output bytes.Buffer
	for index, document := range documents {
		if index > 0 {
			output.WriteString("---\n")
		}

		data, err := yaml.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("marshal Kubernetes manifest: %w", err)
		}
		output.Write(data)
	}

	return output.Bytes(), nil
}

func dependencyDeployment(
	namespace string,
	name string,
	dependency config.DependencyConfig,
) deploymentManifest {
	container := containerManifest{
		Name:  name,
		Image: dependency.Image,
		Env:   sortedEnvironment(dependency.Env),
	}
	if dependency.Port > 0 {
		container.Ports = []containerPort{{ContainerPort: int32(dependency.Port)}}
	}

	if len(dependency.Readiness.Command) > 0 {
		container.ReadinessProbe = dependencyReadinessProbe(
			dependency.Readiness,
			dependency.Readiness.Command,
		)
	} else if dependency.Readiness.Shell != "" {
		container.ReadinessProbe = dependencyReadinessProbe(
			dependency.Readiness,
			[]string{"sh", "-c", dependency.Readiness.Shell},
		)
	}

	return newDeploymentManifest(namespace, name, container)
}

func serviceDeployment(
	namespace string,
	name string,
	service config.ServiceConfig,
) deploymentManifest {
	return newDeploymentManifest(namespace, name, containerManifest{
		Name:  name,
		Image: service.Image,
		Ports: []containerPort{{ContainerPort: service.Port}},
		Env:   sortedEnvironment(service.Env),
		ReadinessProbe: &readinessProbe{
			HTTPGet: &httpGetAction{
				Path: service.HealthPath,
				Port: service.Port,
			},
		},
	})
}

func dependencyReadinessProbe(
	readiness config.ReadinessConfig,
	command []string,
) *readinessProbe {
	return &readinessProbe{
		Exec:           &execAction{Command: command},
		PeriodSeconds:  readiness.IntervalSeconds,
		TimeoutSeconds: readiness.TimeoutSeconds,
	}
}

func newDeploymentManifest(
	namespace string,
	name string,
	container containerManifest,
) deploymentManifest {
	labels := resourceLabels(name, namespace)

	return deploymentManifest{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: objectMetadata{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: deploymentManifestSpec{
			Replicas: 1,
			Selector: labelSelector{
				MatchLabels: selectorLabels(name, namespace),
			},
			Template: podTemplate{
				Metadata: podMetadata{Labels: labels},
				Spec: podSpec{
					Containers: []containerManifest{container},
				},
			},
		},
	}
}

func newServiceManifest(namespace string, name string, port int32) serviceManifest {
	return serviceManifest{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata: objectMetadata{
			Name:      name,
			Namespace: namespace,
			Labels:    resourceLabels(name, namespace),
		},
		Spec: serviceManifestSpec{
			Selector: selectorLabels(name, namespace),
			Ports: []servicePort{{
				Port:       port,
				TargetPort: port,
			}},
		},
	}
}

func resourceLabels(name string, namespace string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "predeploy-guard",
		"app.kubernetes.io/name":       name,
		"app.kubernetes.io/part-of":    "predeploy-guard",
		"predeploy.guard/run":          namespace,
	}
}

func selectorLabels(name string, namespace string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": name,
		"predeploy.guard/run":    namespace,
	}
}

func sortedEnvironment(values map[string]string) []environmentValue {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]environmentValue, 0, len(keys))
	for _, key := range keys {
		result = append(result, environmentValue{
			Name:  key,
			Value: values[key],
		})
	}

	return result
}

func sortedDependencyNames(dependencies map[string]config.DependencyConfig) []string {
	names := make([]string, 0, len(dependencies))
	for name := range dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
