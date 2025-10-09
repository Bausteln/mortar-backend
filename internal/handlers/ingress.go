package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type IngressHandler struct {
	dynamicClient dynamic.Interface
}

func NewIngressHandler(client dynamic.Interface) *IngressHandler {
	return &IngressHandler{
		dynamicClient: client,
	}
}

func (h *IngressHandler) getIngressGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}
}

// GetIngresses returns all ingresses from all namespaces, excluding those that belong to proxy rules
func (h *IngressHandler) GetIngresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all ingresses from all namespaces
	list, err := h.dynamicClient.Resource(h.getIngressGVR()).Namespace("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching ingresses: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter out ingresses that belong to proxy rules
	filteredItems := []unstructured.Unstructured{}
	for _, item := range list.Items {
		if !h.belongsToProxyRule(item) {
			filteredItems = append(filteredItems, item)
		}
	}

	// Create filtered list
	filteredList := &unstructured.UnstructuredList{
		Object: list.Object,
		Items:  filteredItems,
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filteredList); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// belongsToProxyRule checks if an ingress belongs to a proxy rule
// by checking if it's in the proxy-rules namespace
func (h *IngressHandler) belongsToProxyRule(ingress unstructured.Unstructured) bool {
	// Ingresses created by proxy rules are in the proxy-rules namespace
	namespace := ingress.GetNamespace()
	return namespace == "proxy-rules"
}
