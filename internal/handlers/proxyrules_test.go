package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.bausteln.ch/net-core/reverse-proxy/mortar-backend/internal/testutil"
)

func TestProxyRulesHandler_CreateProxyRule(t *testing.T) {
	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid proxy rule",
			body: map[string]interface{}{
				"apiVersion": "bausteln.io/v1",
				"kind":       "Proxyrule",
				"metadata": map[string]interface{}{
					"name": "test-rule",
				},
				"spec": map[string]interface{}{
					"domain":      "example.com",
					"destination": "10.0.0.50",
					"port":        3000,
					"tls":         true,
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing domain",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-rule",
				},
				"spec": map[string]interface{}{
					"destination": "10.0.0.50",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "domain is required",
		},
		{
			name: "missing destination",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-rule",
				},
				"spec": map[string]interface{}{
					"domain": "example.com",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "destination is required",
		},
		{
			name: "invalid port",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-rule",
				},
				"spec": map[string]interface{}{
					"domain":      "example.com",
					"destination": "10.0.0.50",
					"port":        70000,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "port must be between",
		},
		{
			name: "invalid IP destination",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test-rule",
				},
				"spec": map[string]interface{}{
					"domain":      "example.com",
					"destination": "10.300.500.400",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid",
		},
		{
			name: "duplicate name",
			body: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "existing-rule",
				},
				"spec": map[string]interface{}{
					"domain":      "new.example.com",
					"destination": "10.0.0.60",
				},
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			fakeClient := testutil.NewFakeDynamicClient()

			// Seed with existing rule for duplicate test
			if tt.name == "duplicate name" {
				fakeClient.SeedProxyRule("existing-rule", "proxy-rules", "existing.example.com", "10.0.0.50", 3000)
			}

			handler := NewProxyRulesHandler(fakeClient)

			// Create request
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/proxyrules", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			handler.CreateProxyRule(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				body := w.Body.String()
				if body == "" || len(body) < len(tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, body)
				}
			}
		})
	}
}

func TestProxyRulesHandler_GetProxyRules(t *testing.T) {
	fakeClient := testutil.NewFakeDynamicClient()

	// Seed with test data
	fakeClient.SeedProxyRule("rule1", "proxy-rules", "example1.com", "10.0.0.50", 3000)
	fakeClient.SeedProxyRule("rule2", "proxy-rules", "example2.com", "10.0.0.51", 3001)

	handler := NewProxyRulesHandler(fakeClient)

	req := httptest.NewRequest(http.MethodGet, "/api/proxyrules", nil)
	w := httptest.NewRecorder()

	handler.GetProxyRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	items, ok := result["items"].([]interface{})
	if !ok {
		t.Fatal("expected items array in response")
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestProxyRulesHandler_GetProxyRule(t *testing.T) {
	fakeClient := testutil.NewFakeDynamicClient()
	fakeClient.SeedProxyRule("test-rule", "proxy-rules", "example.com", "10.0.0.50", 3000)

	handler := NewProxyRulesHandler(fakeClient)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "existing rule",
			path:           "/api/proxyrules/test-rule",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent rule",
			path:           "/api/proxyrules/non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.GetProxyRule(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProxyRulesHandler_UpdateProxyRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleName       string
		body           map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "valid update",
			ruleName: "test-rule",
			body: map[string]interface{}{
				"spec": map[string]interface{}{
					"domain":      "updated.example.com",
					"destination": "10.0.0.60",
					"port":        3001,
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "invalid domain",
			ruleName: "test-rule",
			body: map[string]interface{}{
				"spec": map[string]interface{}{
					"domain":      "invalid..com",
					"destination": "10.0.0.50",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "consecutive dots",
		},
		{
			name:     "non-existent rule",
			ruleName: "non-existent",
			body: map[string]interface{}{
				"spec": map[string]interface{}{
					"domain":      "example.com",
					"destination": "10.0.0.50",
				},
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := testutil.NewFakeDynamicClient()
			fakeClient.SeedProxyRule("test-rule", "proxy-rules", "example.com", "10.0.0.50", 3000)

			handler := NewProxyRulesHandler(fakeClient)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/api/proxyrules/"+tt.ruleName, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.UpdateProxyRule(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				body := w.Body.String()
				if body == "" || len(body) < len(tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, body)
				}
			}
		})
	}
}

func TestProxyRulesHandler_DeleteProxyRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleName       string
		expectedStatus int
	}{
		{
			name:           "delete existing rule",
			ruleName:       "test-rule",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent rule",
			ruleName:       "non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := testutil.NewFakeDynamicClient()
			fakeClient.SeedProxyRule("test-rule", "proxy-rules", "example.com", "10.0.0.50", 3000)

			handler := NewProxyRulesHandler(fakeClient)

			req := httptest.NewRequest(http.MethodDelete, "/api/proxyrules/"+tt.ruleName, nil)
			w := httptest.NewRecorder()

			handler.DeleteProxyRule(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestProxyRulesHandler_DuplicateDomain(t *testing.T) {
	fakeClient := testutil.NewFakeDynamicClient()
	fakeClient.SeedProxyRule("rule1", "proxy-rules", "example.com", "10.0.0.50", 3000)

	handler := NewProxyRulesHandler(fakeClient)

	// Try to create another rule with the same domain
	body := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "rule2",
		},
		"spec": map[string]interface{}{
			"domain":      "example.com",
			"destination": "10.0.0.60",
		},
	}

	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/proxyrules", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateProxyRule(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}

	responseBody := w.Body.String()
	if responseBody == "" {
		t.Error("expected error message about duplicate domain")
	}
}
