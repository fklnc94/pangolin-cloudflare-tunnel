package traefik

import "testing"

func TestRouter_HasTLSEnabled(t *testing.T) {
	tests := []struct {
		name     string
		router   Router
		expected bool
	}{
		{
			name: "TLS with CertResolver and Options",
			router: Router{
				TLS: struct {
					CertResolver string   `json:"certResolver"`
					Options      string   `json:"options"`
					Domains      []string `json:"domains,omitempty"`
				}{
					CertResolver: "letsencrypt",
					Options:      "default",
				},
			},
			expected: true,
		},
		{
			name: "TLS with CertResolver and Domains",
			router: Router{
				TLS: struct {
					CertResolver string   `json:"certResolver"`
					Options      string   `json:"options"`
					Domains      []string `json:"domains,omitempty"`
				}{
					CertResolver: "letsencrypt",
					Domains:      []string{"example.com"},
				},
			},
			expected: true,
		},
		{
			name: "TLS with only CertResolver",
			router: Router{
				TLS: struct {
					CertResolver string   `json:"certResolver"`
					Options      string   `json:"options"`
					Domains      []string `json:"domains,omitempty"`
				}{
					CertResolver: "letsencrypt",
				},
			},
			expected: false,
		},
		{
			name:     "No TLS",
			router:   Router{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.router.HasTLSEnabled()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRouter_HasMatchingEntrypoint(t *testing.T) {
	tests := []struct {
		name               string
		routerEntrypoints  []string
		allowedEntrypoints []string
		expected           bool
	}{
		{
			name:               "Match found",
			routerEntrypoints:  []string{"web", "websecure"},
			allowedEntrypoints: []string{"web"},
			expected:           true,
		},
		{
			name:               "No match",
			routerEntrypoints:  []string{"internal"},
			allowedEntrypoints: []string{"web", "websecure"},
			expected:           false,
		},
		{
			name:               "Empty allowed list",
			routerEntrypoints:  []string{"web"},
			allowedEntrypoints: []string{},
			expected:           true,
		},
		{
			name:               "Multiple matches",
			routerEntrypoints:  []string{"web", "websecure"},
			allowedEntrypoints: []string{"web", "websecure"},
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := Router{EntryPoints: tt.routerEntrypoints}
			result := router.HasMatchingEntrypoint(tt.allowedEntrypoints)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRouter_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Enabled", "enabled", true},
		{"Disabled", "disabled", false},
		{"Unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := Router{Status: tt.status}
			result := router.IsEnabled()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
