package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		wantError bool
	}{
		{
			name:      "valid name",
			inputName: "test-rule",
			wantError: false,
		},
		{
			name:      "valid name with numbers",
			inputName: "test-rule-123",
			wantError: false,
		},
		{
			name:      "empty name",
			inputName: "",
			wantError: true,
		},
		{
			name:      "uppercase not allowed",
			inputName: "Test-Rule",
			wantError: true,
		},
		{
			name:      "starts with hyphen",
			inputName: "-test-rule",
			wantError: true,
		},
		{
			name:      "ends with hyphen",
			inputName: "test-rule-",
			wantError: true,
		},
		{
			name:      "special characters",
			inputName: "test_rule",
			wantError: true,
		},
		{
			name:      "too long",
			inputName: string(make([]byte, 254)),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": tt.inputName,
					},
				},
			}
			errors := validateMetadata(obj)
			hasError := len(errors) > 0
			if hasError != tt.wantError {
				t.Errorf("validateMetadata() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		wantError bool
	}{
		{
			name:      "valid domain",
			domain:    "example.com",
			wantError: false,
		},
		{
			name:      "valid subdomain",
			domain:    "api.example.com",
			wantError: false,
		},
		{
			name:      "valid with hyphens",
			domain:    "my-api.example-site.com",
			wantError: false,
		},
		{
			name:      "empty domain",
			domain:    "",
			wantError: true,
		},
		{
			name:      "uppercase allowed (converted to lowercase)",
			domain:    "Example.Com",
			wantError: false,
		},
		{
			name:      "starts with dot",
			domain:    ".example.com",
			wantError: true,
		},
		{
			name:      "ends with dot",
			domain:    "example.com.",
			wantError: true,
		},
		{
			name:      "consecutive dots",
			domain:    "example..com",
			wantError: true,
		},
		{
			name:      "too long",
			domain:    string(make([]byte, 254)) + ".com",
			wantError: true,
		},
		{
			name:      "special characters",
			domain:    "example_test.com",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateDomain(tt.domain)
			hasError := len(errors) > 0
			if hasError != tt.wantError {
				t.Errorf("validateDomain() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateDestination(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		wantError   bool
	}{
		{
			name:        "valid IPv4",
			destination: "10.0.0.50",
			wantError:   false,
		},
		{
			name:        "valid IPv4 with max octets",
			destination: "255.255.255.255",
			wantError:   false,
		},
		{
			name:        "valid IPv4 with zeros",
			destination: "0.0.0.0",
			wantError:   false,
		},
		{
			name:        "invalid IPv4 - octet too large",
			destination: "10.300.500.400",
			wantError:   true,
		},
		{
			name:        "invalid IPv4 - octet 256",
			destination: "192.168.1.256",
			wantError:   true,
		},
		{
			name:        "valid DNS name",
			destination: "backend.example.com",
			wantError:   false,
		},
		{
			name:        "valid DNS with hyphens",
			destination: "my-backend.example.com",
			wantError:   false,
		},
		{
			name:        "empty destination",
			destination: "",
			wantError:   true,
		},
		{
			name:        "uppercase in DNS allowed (converted to lowercase)",
			destination: "Backend.Example.Com",
			wantError:   false,
		},
		{
			name:        "starts with dot",
			destination: ".example.com",
			wantError:   true,
		},
		{
			name:        "ends with dot",
			destination: "example.com.",
			wantError:   true,
		},
		{
			name:        "consecutive dots",
			destination: "example..com",
			wantError:   true,
		},
		{
			name:        "valid IPv6",
			destination: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantError:   false,
		},
		{
			name:        "valid IPv6 short form",
			destination: "::1",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateDestination(tt.destination)
			hasError := len(errors) > 0
			if hasError != tt.wantError {
				t.Errorf("validateDestination(%s) error = %v, wantError %v", tt.destination, errors, tt.wantError)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		wantError bool
	}{
		{
			name:      "valid port 80",
			port:      80,
			wantError: false,
		},
		{
			name:      "valid port 3000",
			port:      3000,
			wantError: false,
		},
		{
			name:      "valid port 65535",
			port:      65535,
			wantError: false,
		},
		{
			name:      "valid port 1",
			port:      1,
			wantError: false,
		},
		{
			name:      "port 0",
			port:      0,
			wantError: true,
		},
		{
			name:      "negative port",
			port:      -1,
			wantError: true,
		},
		{
			name:      "port too large",
			port:      65536,
			wantError: true,
		},
		{
			name:      "port way too large",
			port:      100000,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validatePort(tt.port)
			hasError := len(errors) > 0
			if hasError != tt.wantError {
				t.Errorf("validatePort(%d) error = %v, wantError %v", tt.port, errors, tt.wantError)
			}
		})
	}
}

func TestValidateProxyRuleCreate(t *testing.T) {
	tests := []struct {
		name      string
		obj       *unstructured.Unstructured
		wantError bool
	}{
		{
			name: "valid proxy rule",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-rule",
					},
					"spec": map[string]interface{}{
						"domain":      "example.com",
						"destination": "10.0.0.50",
						"port":        int64(3000),
						"tls":         true,
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing name",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{},
					"spec": map[string]interface{}{
						"domain":      "example.com",
						"destination": "10.0.0.50",
					},
				},
			},
			wantError: true,
		},
		{
			name: "missing domain",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-rule",
					},
					"spec": map[string]interface{}{
						"destination": "10.0.0.50",
					},
				},
			},
			wantError: true,
		},
		{
			name: "missing destination",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-rule",
					},
					"spec": map[string]interface{}{
						"domain": "example.com",
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid port",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-rule",
					},
					"spec": map[string]interface{}{
						"domain":      "example.com",
						"destination": "10.0.0.50",
						"port":        int64(70000),
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid IP destination",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-rule",
					},
					"spec": map[string]interface{}{
						"domain":      "example.com",
						"destination": "10.300.500.400",
					},
				},
			},
			wantError: true,
		},
		{
			name: "optional port not provided",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-rule",
					},
					"spec": map[string]interface{}{
						"domain":      "example.com",
						"destination": "10.0.0.50",
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateProxyRuleCreate(tt.obj)
			hasError := len(errors) > 0
			if hasError != tt.wantError {
				t.Errorf("ValidateProxyRuleCreate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}

func TestValidateProxyRuleUpdate(t *testing.T) {
	tests := []struct {
		name      string
		obj       *unstructured.Unstructured
		wantError bool
	}{
		{
			name: "valid update",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "existing-rule",
					},
					"spec": map[string]interface{}{
						"domain":      "updated.example.com",
						"destination": "10.0.0.60",
						"port":        int64(3001),
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid domain update - consecutive dots",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "existing-rule",
					},
					"spec": map[string]interface{}{
						"domain":      "invalid..com",
						"destination": "10.0.0.50",
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateProxyRuleUpdate(tt.obj)
			hasError := len(errors) > 0
			if hasError != tt.wantError {
				t.Errorf("ValidateProxyRuleUpdate() error = %v, wantError %v", errors, tt.wantError)
			}
		})
	}
}
