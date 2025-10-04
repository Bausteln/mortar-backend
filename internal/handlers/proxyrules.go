package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	proxyRulesNamespace = "proxy-rules"
)

type ProxyRulesHandler struct {
	dynamicClient dynamic.Interface
}

func NewProxyRulesHandler(client dynamic.Interface) *ProxyRulesHandler {
	return &ProxyRulesHandler{
		dynamicClient: client,
	}
}

func (h *ProxyRulesHandler) getGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "bausteln.io",
		Version:  "v1",
		Resource: "proxyrules",
	}
}

func (h *ProxyRulesHandler) GetProxyRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get proxyrules from proxy-rules namespace
	list, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).List(context.Background(), metav1.ListOptions{})
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

	// Get specific proxyrule from proxy-rules namespace
	rule, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Get(context.Background(), name, metav1.GetOptions{})
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

func (h *ProxyRulesHandler) CreateProxyRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON into unstructured object
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Create unstructured object
	unstructuredObj := &unstructured.Unstructured{
		Object: obj,
	}

	// Set apiVersion and kind if not provided
	if unstructuredObj.GetAPIVersion() == "" {
		unstructuredObj.SetAPIVersion("bausteln.io/v1")
	}
	if unstructuredObj.GetKind() == "" {
		unstructuredObj.SetKind("Proxyrule")
	}

	// Create the resource
	result, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating proxyrule: %v", err), http.StatusInternalServerError)
		return
	}

	// Return created resource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *ProxyRulesHandler) UpdateProxyRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule name from path
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

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON into unstructured object
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Create unstructured object
	unstructuredObj := &unstructured.Unstructured{
		Object: obj,
	}

	// Ensure name matches
	if unstructuredObj.GetName() != "" && unstructuredObj.GetName() != name {
		http.Error(w, "Resource name in body does not match URL path", http.StatusBadRequest)
		return
	}
	unstructuredObj.SetName(name)

	// Set apiVersion and kind if not provided
	if unstructuredObj.GetAPIVersion() == "" {
		unstructuredObj.SetAPIVersion("bausteln.io/v1")
	}
	if unstructuredObj.GetKind() == "" {
		unstructuredObj.SetKind("Proxyrule")
	}

	// Update the resource
	result, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating proxyrule: %v", err), http.StatusInternalServerError)
		return
	}

	// Return updated resource
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *ProxyRulesHandler) DeleteProxyRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule name from path
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

	// Delete the resource
	err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting proxyrule: %v", err), http.StatusNotFound)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}
