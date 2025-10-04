package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// NewDynamicClient creates a new Kubernetes dynamic client
func NewDynamicClient() (dynamic.Interface, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}
