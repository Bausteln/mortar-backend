package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Build kubeconfig path
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	kubeconfig := filepath.Join(home, ".kube", "config")

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Test connection by getting server version
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("Error connecting to Kubernetes cluster: %v", err)
	}

	fmt.Printf("Successfully connected to Kubernetes cluster!\n")
	fmt.Printf("Server version: %s\n", version.String())

	// Example: List pods in default namespace
	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	fmt.Printf("\nPods in default namespace: %d\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("  - %s\n", pod.Name)
	}
}
