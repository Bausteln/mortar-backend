package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var dynamicClient dynamic.Interface

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

	// Create Kubernetes dynamic client
	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes dynamic client: %v", err)
	}

	// Set up HTTP routes
	http.HandleFunc("/api/proxyrules", getProxyRules)

	// Start server
	port := "8080"
	fmt.Printf("Starting API server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func getProxyRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define the GroupVersionResource for proxyrules
	proxyRuleGVR := schema.GroupVersionResource{
		Group:    "bausteln.io",
		Version:  "v1",
		Resource: "proxyrules",
	}

	// Get proxyrules from proxy-rules namespace
	list, err := dynamicClient.Resource(proxyRuleGVR).Namespace("proxy-rules").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching proxyrules: %v", err), http.StatusInternalServerError)
		return
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}
