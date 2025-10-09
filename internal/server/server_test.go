package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend/internal/testutil"
)

// TestE2E_ProxyRulesWorkflow tests a complete workflow of proxy rule operations
func TestE2E_ProxyRulesWorkflow(t *testing.T) {
	// Create test server
	fakeClient := testutil.NewFakeDynamicClient()
	srv := New("8080", fakeClient)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/proxyrules" && r.Method == http.MethodGet:
			srv.proxyRulesHandler.GetProxyRules(w, r)
		case r.URL.Path == "/api/proxyrules" && r.Method == http.MethodPost:
			srv.proxyRulesHandler.CreateProxyRule(w, r)
		case r.URL.Path == "/health":
			srv.handleHealth(w, r)
		default:
			srv.handleProxyRules(w, r)
		}
	}))
	defer server.Close()

	// Test 1: Health check
	t.Run("health check", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("failed to get health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test 2: List proxy rules (should be empty)
	t.Run("list empty proxy rules", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/proxyrules")
		if err != nil {
			t.Fatalf("failed to list proxy rules: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		items, ok := result["items"].([]interface{})
		if !ok {
			t.Fatal("expected items array in response")
		}

		if len(items) != 0 {
			t.Errorf("expected 0 items, got %d", len(items))
		}
	})

	// Test 3: Create a proxy rule
	var createdName string
	t.Run("create proxy rule", func(t *testing.T) {
		rule := map[string]interface{}{
			"apiVersion": "bausteln.io/v1",
			"kind":       "Proxyrule",
			"metadata": map[string]interface{}{
				"name": "test-e2e-rule",
			},
			"spec": map[string]interface{}{
				"domain":      "e2e-test.example.com",
				"destination": "10.0.0.100",
				"port":        8080,
				"tls":         true,
			},
		}

		bodyBytes, _ := json.Marshal(rule)
		resp, err := http.Post(server.URL+"/api/proxyrules", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("failed to create proxy rule: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}

		var created map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		metadata, ok := created["metadata"].(map[string]interface{})
		if !ok {
			t.Fatal("expected metadata in response")
		}

		createdName, ok = metadata["name"].(string)
		if !ok || createdName == "" {
			t.Fatal("expected name in metadata")
		}
	})

	// Test 4: List proxy rules (should have 1)
	t.Run("list proxy rules with one item", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/proxyrules")
		if err != nil {
			t.Fatalf("failed to list proxy rules: %v", err)
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		items, _ := result["items"].([]interface{})
		if len(items) != 1 {
			t.Errorf("expected 1 item, got %d", len(items))
		}
	})

	// Test 5: Get specific proxy rule
	t.Run("get specific proxy rule", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/proxyrules/" + createdName)
		if err != nil {
			t.Fatalf("failed to get proxy rule: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var rule map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&rule)

		metadata, _ := rule["metadata"].(map[string]interface{})
		name, _ := metadata["name"].(string)
		if name != createdName {
			t.Errorf("expected name %s, got %s", createdName, name)
		}
	})

	// Test 6: Update proxy rule
	t.Run("update proxy rule", func(t *testing.T) {
		update := map[string]interface{}{
			"spec": map[string]interface{}{
				"domain":      "updated-e2e-test.example.com",
				"destination": "10.0.0.101",
				"port":        8081,
			},
		}

		bodyBytes, _ := json.Marshal(update)
		req, _ := http.NewRequest(http.MethodPut, server.URL+"/api/proxyrules/"+createdName, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to update proxy rule: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// Test 7: Verify update
	t.Run("verify proxy rule was updated", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/proxyrules/" + createdName)
		if err != nil {
			t.Fatalf("failed to get proxy rule: %v", err)
		}
		defer resp.Body.Close()

		var rule map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&rule)

		spec, _ := rule["spec"].(map[string]interface{})
		domain, _ := spec["domain"].(string)
		if domain != "updated-e2e-test.example.com" {
			t.Errorf("expected domain 'updated-e2e-test.example.com', got %s", domain)
		}
	})

	// Test 8: Try to create duplicate domain
	t.Run("reject duplicate domain", func(t *testing.T) {
		rule := map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "duplicate-rule",
			},
			"spec": map[string]interface{}{
				"domain":      "updated-e2e-test.example.com", // same domain as existing rule
				"destination": "10.0.0.102",
			},
		}

		bodyBytes, _ := json.Marshal(rule)
		resp, err := http.Post(server.URL+"/api/proxyrules", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("failed to create proxy rule: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("expected status 409 (conflict), got %d", resp.StatusCode)
		}
	})

	// Test 9: Delete proxy rule
	t.Run("delete proxy rule", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, server.URL+"/api/proxyrules/"+createdName, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to delete proxy rule: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", resp.StatusCode)
		}
	})

	// Test 10: Verify deletion
	t.Run("verify proxy rule was deleted", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/proxyrules/" + createdName)
		if err != nil {
			t.Fatalf("failed to get proxy rule: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
}

// TestE2E_ValidationErrors tests various validation error scenarios
func TestE2E_ValidationErrors(t *testing.T) {
	fakeClient := testutil.NewFakeDynamicClient()
	srv := New("8080", fakeClient)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.handleProxyRules(w, r)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		rule           map[string]interface{}
		expectedStatus int
		errorContains  string
	}{
		{
			name: "missing domain",
			rule: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "missing-domain",
				},
				"spec": map[string]interface{}{
					"destination": "10.0.0.50",
				},
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "domain",
		},
		{
			name: "invalid IP",
			rule: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "invalid-ip",
				},
				"spec": map[string]interface{}{
					"domain":      "test.example.com",
					"destination": "300.400.500.600",
				},
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "invalid",
		},
		{
			name: "invalid port",
			rule: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "invalid-port",
				},
				"spec": map[string]interface{}{
					"domain":      "test.example.com",
					"destination": "10.0.0.50",
					"port":        99999,
				},
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "port",
		},
		{
			name: "invalid name",
			rule: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "Invalid_Name",
				},
				"spec": map[string]interface{}{
					"domain":      "test.example.com",
					"destination": "10.0.0.50",
				},
			},
			expectedStatus: http.StatusBadRequest,
			errorContains:  "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.rule)
			resp, err := http.Post(server.URL+"/api/proxyrules", "application/json", bytes.NewReader(bodyBytes))
			if err != nil {
				t.Fatalf("failed to create proxy rule: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var body bytes.Buffer
			body.ReadFrom(resp.Body)
			if tt.errorContains != "" && !bytes.Contains(body.Bytes(), []byte(tt.errorContains)) {
				t.Errorf("expected error containing %q, got %q", tt.errorContains, body.String())
			}
		})
	}
}

// TestE2E_ContentTypeValidation tests content-type validation
func TestE2E_ContentTypeValidation(t *testing.T) {
	fakeClient := testutil.NewFakeDynamicClient()
	srv := New("8080", fakeClient)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.handleProxyRules(w, r)
	}))
	defer server.Close()

	rule := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "test",
		},
		"spec": map[string]interface{}{
			"domain":      "test.example.com",
			"destination": "10.0.0.50",
		},
	}

	bodyBytes, _ := json.Marshal(rule)

	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "valid content type",
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing content type",
			contentType:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid content type",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/proxyrules", bytes.NewReader(bodyBytes))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed to create proxy rule: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body := new(bytes.Buffer)
				body.ReadFrom(resp.Body)
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, body.String())
			}
		})
	}
}

// Helper to setup a test server with routes
func setupTestServer(fakeClient *testutil.FakeDynamicClient) *httptest.Server {
	srv := New("8080", fakeClient)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/health":
			srv.handleHealth(w, r)
		case r.URL.Path == "/api/ingresses":
			srv.handleIngresses(w, r)
		default:
			srv.handleProxyRules(w, r)
		}
	}))
}
