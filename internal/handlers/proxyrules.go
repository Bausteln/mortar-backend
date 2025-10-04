package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ProxyRulesHandler struct {
	dynamicClient dynamic.Interface
}

func NewProxyRulesHandler(client dynamic.Interface) *ProxyRulesHandler {
	return &ProxyRulesHandler{
		dynamicClient: client,
	}
}

func (h *ProxyRulesHandler) GetProxyRules(w http.ResponseWriter, r *http.Request) {
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
	list, err := h.dynamicClient.Resource(proxyRuleGVR).Namespace("proxy-rules").List(context.Background(), metav1.ListOptions{})
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

func (h *ProxyRulesHandler) GetProxyRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule name from path: /api/proxyrules/{name}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid path format. Expected: /api/proxyrules/{name}", http.StatusBadRequest)
		return
	}
	name := parts[2]

	if name == "" {
		http.Error(w, "Rule name is required", http.StatusBadRequest)
		return
	}

	// Define the GroupVersionResource for proxyrules
	proxyRuleGVR := schema.GroupVersionResource{
		Group:    "bausteln.io",
		Version:  "v1",
		Resource: "proxyrules",
	}

	// Get specific proxyrule from proxy-rules namespace
	rule, err := h.dynamicClient.Resource(proxyRuleGVR).Namespace("proxy-rules").Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching proxyrule: %v", err), http.StatusNotFound)
		return
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}
