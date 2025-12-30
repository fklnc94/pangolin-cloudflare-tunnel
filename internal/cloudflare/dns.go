package cloudflare

import (
	"context"
	"fmt"

	"hhftechnology/pangolin-cloudflare-tunnel/pkg/retry"

	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

// SyncDNSRecords updates DNS records for all domains in the ingress rules
// domainToZone maps each domain to its zone ID
func SyncDNSRecords(ctx context.Context, client *Client, tunnelID string, domainToZone map[string]string) error {
	tunnelDomain := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)

	for domain, zoneID := range domainToZone {
		if err := ensureDNSRecord(ctx, client, zoneID, domain, tunnelDomain); err != nil {
			log.WithError(err).Errorf("Failed to ensure DNS record for %s", domain)
			// Continue processing other domains
		}
	}

	return nil
}

// CleanupDNSRecords removes DNS records that are no longer in the active domains list
func CleanupDNSRecords(ctx context.Context, client *Client, tunnelID string, activeDomains map[string]string) error {
	tunnelDomain := fmt.Sprintf("%s.cfargotunnel.com", tunnelID)

	// Group domains by zone for efficient cleanup
	zoneToActiveDomains := make(map[string]map[string]bool)
	for domain, zoneID := range activeDomains {
		if zoneToActiveDomains[zoneID] == nil {
			zoneToActiveDomains[zoneID] = make(map[string]bool)
		}
		zoneToActiveDomains[zoneID][domain] = true
	}

	// Process each zone
	for zoneID, activeDomainsInZone := range zoneToActiveDomains {
		if err := cleanupZoneDNSRecords(ctx, client, zoneID, tunnelDomain, activeDomainsInZone); err != nil {
			log.WithError(err).Errorf("Failed to cleanup DNS records for zone %s", zoneID)
			// Continue processing other zones
		}
	}

	return nil
}

// cleanupZoneDNSRecords removes DNS records in a zone that point to the tunnel but are not in active domains
func cleanupZoneDNSRecords(ctx context.Context, client *Client, zoneID, tunnelDomain string, activeDomains map[string]bool) error {
	zoneIdentifier := cloudflare.ZoneIdentifier(zoneID)

	// List all DNS records in the zone
	allRecords, _, err := client.api.ListDNSRecords(ctx, zoneIdentifier, cloudflare.ListDNSRecordsParams{
		Type: "CNAME",
	})
	if err != nil {
		return fmt.Errorf("failed to list DNS records for zone %s: %w", zoneID, err)
	}

	// Find records that point to our tunnel but are not in active domains
	for _, record := range allRecords {
		if record.Content == tunnelDomain && !activeDomains[record.Name] {
			log.WithFields(log.Fields{
				"domain": record.Name,
				"zone":   zoneID,
			}).Info("Removing stale DNS record")

			if err := deleteDNSRecord(ctx, client, zoneIdentifier, record.ID); err != nil {
				log.WithError(err).Errorf("Failed to delete DNS record %s", record.Name)
				// Continue with other records
			}
		}
	}

	return nil
}

// ensureDNSRecord ensures that a DNS record exists and is correctly configured
func ensureDNSRecord(ctx context.Context, client *Client, zoneID, domain, tunnelDomain string) error {
	return retry.Do(3, func() error {
		zoneIdentifier := cloudflare.ZoneIdentifier(zoneID)

		// Create record template
		proxied := true
		record := cloudflare.DNSRecord{
			Type:    "CNAME",
			Name:    domain,
			Content: tunnelDomain,
			TTL:     1,
			Proxied: &proxied,
		}

		// Check if record already exists
		existingRecords, _, err := client.api.ListDNSRecords(ctx, zoneIdentifier, cloudflare.ListDNSRecordsParams{Name: domain})
		if err != nil {
			return fmt.Errorf("error checking DNS records for %s: %w", domain, err)
		}

		// Create new record if it doesn't exist
		if len(existingRecords) == 0 {
			_, err := client.api.CreateDNSRecord(ctx, zoneIdentifier, cloudflare.CreateDNSRecordParams{
				Name:    record.Name,
				Type:    record.Type,
				Content: record.Content,
				TTL:     record.TTL,
				Proxied: record.Proxied,
			})
			if err != nil {
				return fmt.Errorf("failed to create DNS record for %s: %w", domain, err)
			}
			log.WithField("domain", domain).Info("DNS record created successfully")
			return nil
		}

		// Update record if content doesn't match
		if existingRecords[0].Content != tunnelDomain {
			_, err = client.api.UpdateDNSRecord(ctx, zoneIdentifier, cloudflare.UpdateDNSRecordParams{
				ID:      existingRecords[0].ID,
				Name:    record.Name,
				Type:    record.Type,
				Content: record.Content,
				TTL:     record.TTL,
				Proxied: record.Proxied,
			})
			if err != nil {
				return fmt.Errorf("failed to update DNS record for %s: %w", domain, err)
			}
			log.WithField("domain", domain).Info("DNS record updated successfully")
		}

		return nil
	})
}

// deleteDNSRecord deletes a DNS record by ID
func deleteDNSRecord(ctx context.Context, client *Client, zoneIdentifier *cloudflare.ResourceContainer, recordID string) error {
	return retry.Do(3, func() error {
		err := client.api.DeleteDNSRecord(ctx, zoneIdentifier, recordID)
		if err != nil {
			return fmt.Errorf("failed to delete DNS record: %w", err)
		}
		return nil
	})
}
