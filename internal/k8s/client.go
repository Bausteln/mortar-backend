package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewDynamicClient creates a new Kubernetes dynamic client
// It first tries to use in-cluster config (when running in a pod with ServiceAccount)
// If that fails, it falls back to using kubeconfig file (for local development)
func NewDynamicClient() (dynamic.Interface, error) {
	// Try in-cluster config first (for production deployment)
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file (for local development)
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfig := filepath.Join(home, ".kube", "config")

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return dynamic.NewForConfig(config)
}
