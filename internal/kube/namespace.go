package kube

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GenerateNamespace(prefix string, serviceName string) string {
	timestamp := time.Now().Unix()

	cleanServiceName := strings.ToLower(serviceName)
	cleanServiceName = strings.ReplaceAll(cleanServiceName, "_", "-")

	return fmt.Sprintf("%s-%s-%d", prefix, cleanServiceName, timestamp)
}

func CreateNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "predeploy-guard",
				"predeploy-guard":              "true",
			},
		},
	}

	_, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}

		return fmt.Errorf("Create namespace %s: %w", namespace, err)
	}

	return nil
}

func DeleteNamespace(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	err := clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("Delete namespace %s: %w", namespace, err)
	}

	return nil
}
