// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	CloudflareToken        string
	CloudflareAccountID    string
	CloudflareTunnelID     string
	CloudflareZones        []ZoneConfig
	TraefikAPIEndpoint     string
	TraefikEntrypoints     []string
	TraefikServiceEndpoint string
	SkipTLSRoutes          bool
	PollInterval           time.Duration
	IgnorePatterns         []*regexp.Regexp
	EnableDNSCleanup       bool
}

// ZoneConfig represents a Cloudflare zone configuration
type ZoneConfig struct {
	ZoneID     string
	DomainName string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Parse TraefikEntrypoints from comma-separated list
	entrypointsStr := os.Getenv("TRAEFIK_ENTRYPOINTS")
	var entrypoints []string
	if entrypointsStr != "" {
		for _, ep := range strings.Split(entrypointsStr, ",") {
			entrypoints = append(entrypoints, strings.TrimSpace(ep))
		}
	} else {
		// Backward compatibility for TRAEFIK_ENTRYPOINT
		singleEntrypoint := os.Getenv("TRAEFIK_ENTRYPOINT")
		if singleEntrypoint != "" {
			entrypoints = append(entrypoints, singleEntrypoint)
		}
	}

	// Parse zone configurations (supports both single and multiple zones)
	zones, err := parseZoneConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse zone configurations: %w", err)
	}

	// Parse SkipTLSRoutes (default to true for backward compatibility)
	skipTLSRoutes := true
	skipTLSStr := os.Getenv("SKIP_TLS_ROUTES")
	if skipTLSStr != "" {
		var err error
		skipTLSRoutes, err = strconv.ParseBool(skipTLSStr)
		if err != nil {
			log.Warnf("Invalid SKIP_TLS_ROUTES value: %s, defaulting to true", skipTLSStr)
		}
	}

	// Parse poll interval
	pollInterval := 10 * time.Second
	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	if pollIntervalStr != "" {
		var err error
		pollInterval, err = time.ParseDuration(pollIntervalStr)
		if err != nil {
			log.Warnf("Invalid POLL_INTERVAL value: %s, defaulting to 10s", pollIntervalStr)
			pollInterval = 10 * time.Second
		}
	}

	// Parse ignore patterns for resource exclusion
	ignorePatterns, err := parseIgnorePatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to parse ignore patterns: %w", err)
	}

	// Parse DNS cleanup option (default to true for new behavior)
	enableDNSCleanup := true
	cleanupStr := os.Getenv("ENABLE_DNS_CLEANUP")
	if cleanupStr != "" {
		var err error
		enableDNSCleanup, err = strconv.ParseBool(cleanupStr)
		if err != nil {
			log.Warnf("Invalid ENABLE_DNS_CLEANUP value: %s, defaulting to true", cleanupStr)
		}
	}

	config := &Config{
		CloudflareToken:        os.Getenv("CLOUDFLARE_API_TOKEN"),
		CloudflareAccountID:    os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		CloudflareTunnelID:     os.Getenv("CLOUDFLARE_TUNNEL_ID"),
		CloudflareZones:        zones,
		TraefikAPIEndpoint:     os.Getenv("TRAEFIK_API_ENDPOINT"),
		TraefikEntrypoints:     entrypoints,
		TraefikServiceEndpoint: os.Getenv("TRAEFIK_SERVICE_ENDPOINT"),
		SkipTLSRoutes:          skipTLSRoutes,
		PollInterval:           pollInterval,
		IgnorePatterns:         ignorePatterns,
		EnableDNSCleanup:       enableDNSCleanup,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// parseZoneConfigs parses zone configurations from environment variables
// Supports both legacy single zone and new multi-zone configurations
func parseZoneConfigs() ([]ZoneConfig, error) {
	zones := []ZoneConfig{}

	// Try new multi-zone configuration first
	zoneIDsStr := os.Getenv("CLOUDFLARE_ZONE_IDS")
	domainNamesStr := os.Getenv("DOMAIN_NAMES")

	if zoneIDsStr != "" && domainNamesStr != "" {
		// Multi-zone configuration
		zoneIDs := strings.Split(zoneIDsStr, ",")
		domainNames := strings.Split(domainNamesStr, ",")

		if len(zoneIDs) != len(domainNames) {
			return nil, fmt.Errorf("CLOUDFLARE_ZONE_IDS and DOMAIN_NAMES must have the same number of comma-separated values")
		}

		for i := range zoneIDs {
			zones = append(zones, ZoneConfig{
				ZoneID:     strings.TrimSpace(zoneIDs[i]),
				DomainName: strings.TrimSpace(domainNames[i]),
			})
		}

		log.Infof("Loaded %d zone configurations", len(zones))
		return zones, nil
	}

	// Fallback to legacy single zone configuration
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	domainName := os.Getenv("DOMAIN_NAME")

	if zoneID != "" {
		zones = append(zones, ZoneConfig{
			ZoneID:     zoneID,
			DomainName: domainName, // May be empty for backward compatibility
		})
		log.Info("Loaded single zone configuration (legacy mode)")
		return zones, nil
	}

	return nil, fmt.Errorf("no zone configuration found: set either CLOUDFLARE_ZONE_ID or CLOUDFLARE_ZONE_IDS+DOMAIN_NAMES")
}

// parseIgnorePatterns parses resource ignore patterns from environment variables
func parseIgnorePatterns() ([]*regexp.Regexp, error) {
	patternsStr := os.Getenv("IGNORE_PATTERNS")
	if patternsStr == "" {
		return nil, nil
	}

	patterns := []*regexp.Regexp{}
	for _, patternStr := range strings.Split(patternsStr, ",") {
		patternStr = strings.TrimSpace(patternStr)
		if patternStr == "" {
			continue
		}

		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore pattern %q: %w", patternStr, err)
		}
		patterns = append(patterns, pattern)
	}

	log.Infof("Loaded %d ignore patterns", len(patterns))
	return patterns, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	missing := []string{}

	if c.CloudflareToken == "" {
		missing = append(missing, "CLOUDFLARE_API_TOKEN")
	}
	if c.CloudflareAccountID == "" {
		missing = append(missing, "CLOUDFLARE_ACCOUNT_ID")
	}
	if c.CloudflareTunnelID == "" {
		missing = append(missing, "CLOUDFLARE_TUNNEL_ID")
	}
	if len(c.CloudflareZones) == 0 {
		missing = append(missing, "CLOUDFLARE_ZONE_ID or CLOUDFLARE_ZONE_IDS+DOMAIN_NAMES")
	}
	if c.TraefikAPIEndpoint == "" {
		missing = append(missing, "TRAEFIK_API_ENDPOINT")
	}
	if len(c.TraefikEntrypoints) == 0 {
		missing = append(missing, "TRAEFIK_ENTRYPOINTS or TRAEFIK_ENTRYPOINT")
	}
	if c.TraefikServiceEndpoint == "" {
		missing = append(missing, "TRAEFIK_SERVICE_ENDPOINT")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	// Validate zone configurations
	for i, zone := range c.CloudflareZones {
		if zone.ZoneID == "" {
			return fmt.Errorf("zone %d has empty zone ID", i)
		}
	}

	return nil
}

// GetZoneForDomain returns the zone configuration for a given domain
func (c *Config) GetZoneForDomain(domain string) (*ZoneConfig, error) {
	// If we have domain names configured, match by domain
	for i := range c.CloudflareZones {
		zone := &c.CloudflareZones[i]
		if zone.DomainName != "" {
			// Check if domain ends with the zone's domain name
			if domain == zone.DomainName || strings.HasSuffix(domain, "."+zone.DomainName) {
				return zone, nil
			}
		}
	}

	// Fallback to first zone if no domain names configured (legacy behavior)
	if len(c.CloudflareZones) == 1 && c.CloudflareZones[0].DomainName == "" {
		return &c.CloudflareZones[0], nil
	}

	// If we have multiple zones but no match, return error
	return nil, fmt.Errorf("no zone configuration found for domain %q", domain)
}

// ShouldIgnoreDomain checks if a domain should be ignored based on configured patterns
func (c *Config) ShouldIgnoreDomain(domain string) bool {
	for _, pattern := range c.IgnorePatterns {
		if pattern.MatchString(domain) {
			log.WithFields(log.Fields{
				"domain":  domain,
				"pattern": pattern.String(),
			}).Debug("Domain matches ignore pattern")
			return true
		}
	}
	return false
}
