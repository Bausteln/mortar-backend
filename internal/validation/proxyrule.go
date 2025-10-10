package validation

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ProxyRuleSpec represents the expected structure of a ProxyRule spec
type ProxyRuleSpec struct {
	Domain       string
	Destination  string
	Destinations []string
	Port         int
	TLS          bool
	Annotations  map[string]string
}

const (
	// maxNameLength is the maximum length for Kubernetes resource names
	maxNameLength = 253
	// maxDomainLength is the maximum length for a domain name
	maxDomainLength = 253
	// minPort and maxPort define valid port range
	minPort = 1
	maxPort = 65535
)

var (
	// dnsNameRegex validates DNS names (RFC 1123)
	dnsNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	// k8sNameRegex validates Kubernetes resource names (RFC 1123 subdomain)
	k8sNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	// ipv4Pattern matches strings that look like IPv4 addresses (digits and dots)
	ipv4Pattern = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)
)

// ValidateProxyRuleCreate validates a ProxyRule object for creation
func ValidateProxyRuleCreate(obj *unstructured.Unstructured) ValidationErrors {
	var errors ValidationErrors

	// Validate metadata
	errors = append(errors, validateMetadata(obj)...)

	// Validate spec
	errors = append(errors, validateSpec(obj)...)

	return errors
}

// ValidateProxyRuleUpdate validates a ProxyRule object for update
func ValidateProxyRuleUpdate(obj *unstructured.Unstructured) ValidationErrors {
	var errors ValidationErrors

	// Validate spec (metadata name cannot be changed in updates)
	errors = append(errors, validateSpec(obj)...)

	return errors
}

// validateMetadata validates the metadata section
func validateMetadata(obj *unstructured.Unstructured) ValidationErrors {
	var errors ValidationErrors

	// Validate name
	name := obj.GetName()
	if name == "" {
		errors = append(errors, ValidationError{
			Field:   "metadata.name",
			Message: "name is required",
		})
	} else {
		if len(name) > maxNameLength {
			errors = append(errors, ValidationError{
				Field:   "metadata.name",
				Message: fmt.Sprintf("name must not exceed %d characters", maxNameLength),
			})
		}
		if !k8sNameRegex.MatchString(name) {
			errors = append(errors, ValidationError{
				Field:   "metadata.name",
				Message: "name must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character",
			})
		}
	}

	return errors
}

// validateSpec validates the spec section
func validateSpec(obj *unstructured.Unstructured) ValidationErrors {
	var errors ValidationErrors

	spec, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil {
		errors = append(errors, ValidationError{
			Field:   "spec",
			Message: fmt.Sprintf("invalid spec structure: %v", err),
		})
		return errors
	}
	if !found {
		errors = append(errors, ValidationError{
			Field:   "spec",
			Message: "spec is required",
		})
		return errors
	}

	// Validate domain (required)
	domain, found, err := unstructured.NestedString(spec, "domain")
	if err != nil {
		errors = append(errors, ValidationError{
			Field:   "spec.domain",
			Message: fmt.Sprintf("invalid domain type: %v", err),
		})
	} else if !found || domain == "" {
		errors = append(errors, ValidationError{
			Field:   "spec.domain",
			Message: "domain is required",
		})
	} else {
		errors = append(errors, validateDomain(domain)...)
	}

	// Validate destination/destinations (at least one is required)
	destination, destFound, destErr := unstructured.NestedString(spec, "destination")
	destinations, destsFound, destsErr := unstructured.NestedStringSlice(spec, "destinations")

	// Check if at least one is provided
	if (!destFound || destination == "") && (!destsFound || len(destinations) == 0) {
		errors = append(errors, ValidationError{
			Field:   "spec.destination/destinations",
			Message: "either destination or destinations is required",
		})
	}

	// Validate single destination if provided
	if destErr != nil {
		errors = append(errors, ValidationError{
			Field:   "spec.destination",
			Message: fmt.Sprintf("invalid destination type: %v", destErr),
		})
	} else if destFound && destination != "" {
		errors = append(errors, validateDestination(destination)...)
	}

	// Validate destinations array if provided
	if destsErr != nil {
		errors = append(errors, ValidationError{
			Field:   "spec.destinations",
			Message: fmt.Sprintf("invalid destinations type: %v", destsErr),
		})
	} else if destsFound && len(destinations) > 0 {
		for i, dest := range destinations {
			if dest == "" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("spec.destinations[%d]", i),
					Message: "destination cannot be empty",
				})
			} else {
				// Validate each destination and prefix field name with index
				destErrors := validateDestination(dest)
				for _, e := range destErrors {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("spec.destinations[%d]", i),
						Message: e.Message,
					})
				}
			}
		}
	}

	// Validate port (optional)
	if portVal, found := spec["port"]; found {
		port, ok := portVal.(int64)
		if !ok {
			// Try to convert from float64 (common in JSON unmarshaling)
			if portFloat, ok := portVal.(float64); ok {
				port = int64(portFloat)
			} else {
				errors = append(errors, ValidationError{
					Field:   "spec.port",
					Message: "port must be an integer",
				})
			}
		}
		if ok || port != 0 {
			errors = append(errors, validatePort(int(port))...)
		}
	}

	// Validate TLS (optional)
	if tlsVal, found := spec["tls"]; found {
		if _, ok := tlsVal.(bool); !ok {
			errors = append(errors, ValidationError{
				Field:   "spec.tls",
				Message: "tls must be a boolean",
			})
		}
	}

	// Validate annotations (optional)
	if annotationsVal, found := spec["annotations"]; found {
		annotations, ok := annotationsVal.(map[string]interface{})
		if !ok {
			errors = append(errors, ValidationError{
				Field:   "spec.annotations",
				Message: "annotations must be a map of strings",
			})
		} else {
			for key, value := range annotations {
				if _, ok := value.(string); !ok {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("spec.annotations.%s", key),
						Message: "annotation value must be a string",
					})
				}
			}
		}
	}

	return errors
}

// validateDomain validates a domain name
func validateDomain(domain string) ValidationErrors {
	var errors ValidationErrors

	if len(domain) > maxDomainLength {
		errors = append(errors, ValidationError{
			Field:   "spec.domain",
			Message: fmt.Sprintf("domain must not exceed %d characters", maxDomainLength),
		})
	}

	// Check if it's a valid DNS name
	if !dnsNameRegex.MatchString(strings.ToLower(domain)) {
		errors = append(errors, ValidationError{
			Field:   "spec.domain",
			Message: "domain must be a valid DNS name (lowercase alphanumeric characters, '-', and '.' only)",
		})
	}

	// Check for leading/trailing dots
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		errors = append(errors, ValidationError{
			Field:   "spec.domain",
			Message: "domain must not start or end with a dot",
		})
	}

	// Check for consecutive dots
	if strings.Contains(domain, "..") {
		errors = append(errors, ValidationError{
			Field:   "spec.domain",
			Message: "domain must not contain consecutive dots",
		})
	}

	return errors
}

// validateDestination validates a destination (IP address or DNS name)
func validateDestination(destination string) ValidationErrors {
	var errors ValidationErrors

	// Check if it looks like an IPv4 address
	if ipv4Pattern.MatchString(destination) {
		// If it matches the IPv4 pattern, it must be a valid IP
		if net.ParseIP(destination) == nil {
			errors = append(errors, ValidationError{
				Field:   "spec.destination",
				Message: "destination appears to be an IPv4 address but is invalid (octets must be 0-255)",
			})
		}
		return errors
	}

	// Check if it's a valid IPv6 address
	if net.ParseIP(destination) != nil {
		return errors // Valid IPv6 address
	}

	// Otherwise, validate as DNS name
	if !dnsNameRegex.MatchString(strings.ToLower(destination)) {
		errors = append(errors, ValidationError{
			Field:   "spec.destination",
			Message: "destination must be a valid IP address or DNS name",
		})
	}

	// Check for leading/trailing dots
	if strings.HasPrefix(destination, ".") || strings.HasSuffix(destination, ".") {
		errors = append(errors, ValidationError{
			Field:   "spec.destination",
			Message: "destination must not start or end with a dot",
		})
	}

	// Check for consecutive dots
	if strings.Contains(destination, "..") {
		errors = append(errors, ValidationError{
			Field:   "spec.destination",
			Message: "destination must not contain consecutive dots",
		})
	}

	return errors
}

// validatePort validates a port number
func validatePort(port int) ValidationErrors {
	var errors ValidationErrors

	if port < minPort || port > maxPort {
		errors = append(errors, ValidationError{
			Field:   "spec.port",
			Message: fmt.Sprintf("port must be between %d and %d", minPort, maxPort),
		})
	}

	return errors
}
