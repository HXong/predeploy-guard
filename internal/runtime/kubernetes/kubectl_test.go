package kubernetes

import (
	"reflect"
	"testing"
)

func TestKubectlArgsWithoutContext(t *testing.T) {
	got := kubectlArgs("", "get", "pods", "--namespace", "test")
	want := []string{"get", "pods", "--namespace", "test"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("kubectlArgs = %#v, want %#v", got, want)
	}
}

func TestKubectlArgsWithContext(t *testing.T) {
	got := kubectlArgs("kind-local", "get", "pods", "--namespace", "test")
	want := []string{
		"--context",
		"kind-local",
		"get",
		"pods",
		"--namespace",
		"test",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("kubectlArgs = %#v, want %#v", got, want)
	}
}
