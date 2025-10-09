package validation

import (
	"fmt"
	"io"
	"net/http"
)

const (
	// MaxRequestBodySize is the maximum allowed request body size (1MB)
	MaxRequestBodySize = 1 * 1024 * 1024 // 1MB
)

// ValidateJSONRequest validates that the request has appropriate JSON content type and size
func ValidateJSONRequest(w http.ResponseWriter, r *http.Request) error {
	// Check Content-Type header for POST/PUT/PATCH requests
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			return &ValidationError{
				Field:   "Content-Type",
				Message: "Content-Type header is required",
			}
		}

		// Check if Content-Type is application/json (allow charset parameter)
		if contentType != "application/json" && contentType != "application/json; charset=utf-8" {
			return &ValidationError{
				Field:   "Content-Type",
				Message: fmt.Sprintf("Content-Type must be 'application/json', got '%s'", contentType),
			}
		}
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

	return nil
}

// ValidateRequestBody validates that the request body is not empty and not too large
func ValidateRequestBody(body []byte) error {
	if len(body) == 0 {
		return &ValidationError{
			Field:   "body",
			Message: "request body is required",
		}
	}

	if len(body) > MaxRequestBodySize {
		return &ValidationError{
			Field:   "body",
			Message: fmt.Sprintf("request body size exceeds maximum of %d bytes", MaxRequestBodySize),
		}
	}

	return nil
}

// HandleValidationError sends an appropriate error response for validation errors
func HandleValidationError(w http.ResponseWriter, err error) {
	if validationErr, ok := err.(*ValidationError); ok {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	if validationErrs, ok := err.(ValidationErrors); ok {
		if len(validationErrs) > 0 {
			http.Error(w, validationErrs.Error(), http.StatusBadRequest)
			return
		}
	}

	// Check for MaxBytesReader error
	if err == io.ErrUnexpectedEOF || err.Error() == "http: request body too large" {
		http.Error(w, fmt.Sprintf("request body too large (max %d bytes)", MaxRequestBodySize), http.StatusRequestEntityTooLarge)
		return
	}

	// Generic error
	http.Error(w, fmt.Sprintf("validation error: %v", err), http.StatusBadRequest)
}
