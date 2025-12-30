// Package errors provides custom error types for the application.
package errors

import "fmt"

// ConfigError represents a configuration-related error
type ConfigError struct {
	Field   string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error for %s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("config error for %s: %s", e.Field, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new configuration error
func NewConfigError(field, message string, err error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// TraefikError represents a Traefik API-related error
type TraefikError struct {
	Operation string
	Err       error
}

func (e *TraefikError) Error() string {
	return fmt.Sprintf("traefik %s error: %v", e.Operation, e.Err)
}

func (e *TraefikError) Unwrap() error {
	return e.Err
}

// NewTraefikError creates a new Traefik error
func NewTraefikError(operation string, err error) *TraefikError {
	return &TraefikError{
		Operation: operation,
		Err:       err,
	}
}

// CloudflareError represents a Cloudflare API-related error
type CloudflareError struct {
	Operation string
	Resource  string
	Err       error
}

func (e *CloudflareError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("cloudflare %s error for %s: %v", e.Operation, e.Resource, e.Err)
	}
	return fmt.Sprintf("cloudflare %s error: %v", e.Operation, e.Err)
}

func (e *CloudflareError) Unwrap() error {
	return e.Err
}

// NewCloudflareError creates a new Cloudflare error
func NewCloudflareError(operation, resource string, err error) *CloudflareError {
	return &CloudflareError{
		Operation: operation,
		Resource:  resource,
		Err:       err,
	}
}

// SyncError represents a synchronization error
type SyncError struct {
	Phase string
	Err   error
}

func (e *SyncError) Error() string {
	return fmt.Sprintf("sync error during %s: %v", e.Phase, e.Err)
}

func (e *SyncError) Unwrap() error {
	return e.Err
}

// NewSyncError creates a new synchronization error
func NewSyncError(phase string, err error) *SyncError {
	return &SyncError{
		Phase: phase,
		Err:   err,
	}
}
