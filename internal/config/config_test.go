package config

import (
	"os"
	"testing"
)

func TestParseZoneConfigs_MultiZone(t *testing.T) {
	os.Setenv("CLOUDFLARE_ZONE_IDS", "zone1,zone2,zone3")
	os.Setenv("DOMAIN_NAMES", "example.com,test.com,demo.com")
	defer os.Unsetenv("CLOUDFLARE_ZONE_IDS")
	defer os.Unsetenv("DOMAIN_NAMES")

	zones, err := parseZoneConfigs()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(zones) != 3 {
		t.Errorf("Expected 3 zones, got %d", len(zones))
	}

	expected := []ZoneConfig{
		{ZoneID: "zone1", DomainName: "example.com"},
		{ZoneID: "zone2", DomainName: "test.com"},
		{ZoneID: "zone3", DomainName: "demo.com"},
	}

	for i, zone := range zones {
		if zone.ZoneID != expected[i].ZoneID {
			t.Errorf("Zone %d: expected ZoneID %s, got %s", i, expected[i].ZoneID, zone.ZoneID)
		}
		if zone.DomainName != expected[i].DomainName {
			t.Errorf("Zone %d: expected DomainName %s, got %s", i, expected[i].DomainName, zone.DomainName)
		}
	}
}

func TestParseZoneConfigs_SingleZone(t *testing.T) {
	os.Setenv("CLOUDFLARE_ZONE_ID", "single-zone")
	os.Setenv("DOMAIN_NAME", "example.com")
	defer os.Unsetenv("CLOUDFLARE_ZONE_ID")
	defer os.Unsetenv("DOMAIN_NAME")

	zones, err := parseZoneConfigs()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(zones) != 1 {
		t.Errorf("Expected 1 zone, got %d", len(zones))
	}

	if zones[0].ZoneID != "single-zone" {
		t.Errorf("Expected ZoneID 'single-zone', got %s", zones[0].ZoneID)
	}

	if zones[0].DomainName != "example.com" {
		t.Errorf("Expected DomainName 'example.com', got %s", zones[0].DomainName)
	}
}

func TestParseZoneConfigs_MismatchedCounts(t *testing.T) {
	os.Setenv("CLOUDFLARE_ZONE_IDS", "zone1,zone2")
	os.Setenv("DOMAIN_NAMES", "example.com")
	defer os.Unsetenv("CLOUDFLARE_ZONE_IDS")
	defer os.Unsetenv("DOMAIN_NAMES")

	_, err := parseZoneConfigs()
	if err == nil {
		t.Error("Expected error for mismatched zone IDs and domain names")
	}
}

func TestGetZoneForDomain(t *testing.T) {
	cfg := &Config{
		CloudflareZones: []ZoneConfig{
			{ZoneID: "zone1", DomainName: "example.com"},
			{ZoneID: "zone2", DomainName: "test.com"},
		},
	}

	tests := []struct {
		domain   string
		expected string
		hasError bool
	}{
		{"example.com", "zone1", false},
		{"sub.example.com", "zone1", false},
		{"test.com", "zone2", false},
		{"api.test.com", "zone2", false},
		{"unknown.com", "", true},
	}

	for _, tt := range tests {
		zone, err := cfg.GetZoneForDomain(tt.domain)
		if tt.hasError {
			if err == nil {
				t.Errorf("Expected error for domain %s", tt.domain)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for domain %s: %v", tt.domain, err)
			}
			if zone.ZoneID != tt.expected {
				t.Errorf("Domain %s: expected zone %s, got %s", tt.domain, tt.expected, zone.ZoneID)
			}
		}
	}
}

func TestShouldIgnoreDomain(t *testing.T) {
	os.Setenv("IGNORE_PATTERNS", "^jellyfin\\.,^media\\.")
	defer os.Unsetenv("IGNORE_PATTERNS")

	patterns, err := parseIgnorePatterns()
	if err != nil {
		t.Fatalf("Failed to parse ignore patterns: %v", err)
	}

	cfg := &Config{
		IgnorePatterns: patterns,
	}

	tests := []struct {
		domain   string
		expected bool
	}{
		{"jellyfin.example.com", true},
		{"media.example.com", true},
		{"api.example.com", false},
		{"example.com", false},
	}

	for _, tt := range tests {
		result := cfg.ShouldIgnoreDomain(tt.domain)
		if result != tt.expected {
			t.Errorf("Domain %s: expected %v, got %v", tt.domain, tt.expected, result)
		}
	}
}
