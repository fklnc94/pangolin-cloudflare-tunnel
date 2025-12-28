// Package traefik provides Traefik API client and router management.
package traefik

import (
	"fmt"

	"github.com/traefik/traefik/v3/pkg/muxer/http"
)

// Router represents a Traefik router configuration
type Router struct {
	EntryPoints []string `json:"entryPoints"`
	Service     string   `json:"service"`
	Rule        string   `json:"rule"`
	Status      string   `json:"status"`
	Using       []string `json:"using"`
	ServiceName string   `json:"name"`
	Provider    string   `json:"provider"`
	Middlewares []string `json:"middlewares,omitempty"`
	TLS         struct {
		CertResolver string   `json:"certResolver"`
		Options      string   `json:"options"`
		Domains      []string `json:"domains,omitempty"`
	} `json:"tls,omitempty"`
	Priority int `json:"priority,omitempty"`
}

// HasTLSEnabled determines if a router has TLS properly configured
func (r *Router) HasTLSEnabled() bool {
	return r.TLS.CertResolver != "" &&
		(r.TLS.Options != "" || len(r.TLS.Domains) > 0)
}

// HasMatchingEntrypoint checks if any of the router's entrypoints match the allowed list
func (r *Router) HasMatchingEntrypoint(allowedEntrypoints []string) bool {
	// If no allowed entrypoints specified, accept all
	if len(allowedEntrypoints) == 0 {
		return true
	}

	for _, routerEP := range r.EntryPoints {
		for _, allowedEP := range allowedEntrypoints {
			if routerEP == allowedEP {
				return true
			}
		}
	}
	return false
}

// ParseDomains extracts domains from the router's rule
func (r *Router) ParseDomains() ([]string, error) {
	domains, err := http.ParseDomains(r.Rule)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domains from rule %q: %w", r.Rule, err)
	}
	return domains, nil
}

// IsEnabled checks if the router is enabled
func (r *Router) IsEnabled() bool {
	return r.Status == "enabled"
}
