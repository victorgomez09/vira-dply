package cloudprovider

import "fmt"

// CloudProviderError represents errors from cloud provider operations
type CloudProviderError struct {
	Provider  string `json:"provider"`
	Operation string `json:"operation"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

func (e *CloudProviderError) Error() string {
	return fmt.Sprintf("%s %s error [%s]: %s", e.Provider, e.Operation, e.Code, e.Message)
}

// NewCloudProviderError creates a new cloud provider error
func NewCloudProviderError(provider, operation, code, message string, retryable bool) *CloudProviderError {
	return &CloudProviderError{
		Provider:  provider,
		Operation: operation,
		Code:      code,
		Message:   message,
		Retryable: retryable,
	}
}

// Common error codes
const (
	ErrCodeInvalidCredentials     = "INVALID_CREDENTIALS"
	ErrCodeResourceNotFound       = "RESOURCE_NOT_FOUND"
	ErrCodeResourceAlreadyExists  = "RESOURCE_ALREADY_EXISTS"
	ErrCodeQuotaExceeded         = "QUOTA_EXCEEDED"
	ErrCodePermissionDenied      = "PERMISSION_DENIED"
	ErrCodeInvalidConfiguration  = "INVALID_CONFIGURATION"
	ErrCodeNetworkError          = "NETWORK_ERROR"
	ErrCodeServiceUnavailable    = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimited          = "RATE_LIMITED"
	ErrCodeInternalError        = "INTERNAL_ERROR"
)

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if cloudErr, ok := err.(*CloudProviderError); ok {
		return cloudErr.Retryable
	}
	return false
}

// Validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error `json:"errors"`
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors occurred: %d errors", len(e.Errors))
}

func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}