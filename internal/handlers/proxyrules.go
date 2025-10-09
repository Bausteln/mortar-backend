package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend/internal/validation"
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

	// Validate request (content-type, body size)
	if err := validation.ValidateJSONRequest(w, r); err != nil {
		validation.HandleValidationError(w, err)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		validation.HandleValidationError(w, err)
		return
	}
	defer r.Body.Close()

	// Validate request body
	if err := validation.ValidateRequestBody(body); err != nil {
		validation.HandleValidationError(w, err)
		return
	}

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

	// Set namespace if not provided
	if unstructuredObj.GetNamespace() == "" {
		unstructuredObj.SetNamespace(proxyRulesNamespace)
	}

	// Validate ProxyRule
	if validationErrs := validation.ValidateProxyRuleCreate(unstructuredObj); len(validationErrs) > 0 {
		validation.HandleValidationError(w, validationErrs)
		return
	}

	// Check for duplicate name
	existingByName, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Get(context.Background(), unstructuredObj.GetName(), metav1.GetOptions{})
	if err == nil && existingByName != nil {
		http.Error(w, fmt.Sprintf("Proxy rule with name '%s' already exists", unstructuredObj.GetName()), http.StatusConflict)
		return
	}

	// Check for duplicate domain
	if err := h.checkDuplicateDomain(unstructuredObj, ""); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
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

	// Validate request (content-type, body size)
	if err := validation.ValidateJSONRequest(w, r); err != nil {
		validation.HandleValidationError(w, err)
		return
	}

	// Fetch the existing resource to get resourceVersion
	existing, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching existing proxyrule: %v", err), http.StatusNotFound)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		validation.HandleValidationError(w, err)
		return
	}
	defer r.Body.Close()

	// Validate request body
	if err := validation.ValidateRequestBody(body); err != nil {
		validation.HandleValidationError(w, err)
		return
	}

	// Parse JSON into map
	var updates map[string]interface{}
	if err := json.Unmarshal(body, &updates); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Update the spec field from the request
	if spec, ok := updates["spec"]; ok {
		existing.Object["spec"] = spec
	}

	// Update metadata labels and annotations if provided
	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		existingMetadata := existing.Object["metadata"].(map[string]interface{})

		if labels, ok := metadata["labels"]; ok {
			existingMetadata["labels"] = labels
		}
		if annotations, ok := metadata["annotations"]; ok {
			existingMetadata["annotations"] = annotations
		}
	}

	// Validate updated ProxyRule
	if validationErrs := validation.ValidateProxyRuleUpdate(existing); len(validationErrs) > 0 {
		validation.HandleValidationError(w, validationErrs)
		return
	}

	// Check for duplicate domain (excluding the current rule)
	if err := h.checkDuplicateDomain(existing, name); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Update the resource
	result, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).Update(context.Background(), existing, metav1.UpdateOptions{})
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

// checkDuplicateDomain checks if another proxy rule already uses the same domain
// excludeName is used during updates to exclude the rule being updated from the check
func (h *ProxyRulesHandler) checkDuplicateDomain(obj *unstructured.Unstructured, excludeName string) error {
	// Get the domain from the spec
	domain, found, err := unstructured.NestedString(obj.Object, "spec", "domain")
	if err != nil || !found || domain == "" {
		return nil // No domain to check
	}

	// List all proxy rules
	list, err := h.dynamicClient.Resource(h.getGVR()).Namespace(proxyRulesNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error checking for duplicate domain: %v", err)
	}

	// Check each rule for matching domain
	for _, item := range list.Items {
		// Skip the rule we're updating (if any)
		if excludeName != "" && item.GetName() == excludeName {
			continue
		}

		existingDomain, found, err := unstructured.NestedString(item.Object, "spec", "domain")
		if err != nil || !found {
			continue
		}

		if existingDomain == domain {
			return fmt.Errorf("proxy rule with domain '%s' already exists (used by rule '%s')", domain, item.GetName())
		}
	}

	return nil
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
