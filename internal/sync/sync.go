// Package sync provides synchronization logic between Traefik and Cloudflare.
package sync

import (
	"context"
	"fmt"
	"reflect"

	"hhftechnology/pangolin-cloudflare-tunnel/internal/cloudflare"
	"hhftechnology/pangolin-cloudflare-tunnel/internal/config"
	"hhftechnology/pangolin-cloudflare-tunnel/internal/traefik"

	log "github.com/sirupsen/logrus"
)

// Service orchestrates the synchronization between Traefik and Cloudflare
type Service struct {
	config           *config.Config
	traefikClient    *traefik.Client
	cloudflareClient *cloudflare.Client
	cache            []traefik.Router
}

// NewService creates a new sync service
func NewService(cfg *config.Config, traefikClient *traefik.Client, cloudflareClient *cloudflare.Client) *Service {
	return &Service{
		config:           cfg,
		traefikClient:    traefikClient,
		cloudflareClient: cloudflareClient,
	}
}

// Run starts the main synchronization loop
func (s *Service) Run(ctx context.Context) error {
	pollCh := traefik.Poll(ctx, s.traefikClient, s.config.PollInterval)

	for {
		select {
		case <-ctx.Done():
			return nil
		case poll, ok := <-pollCh:
			if !ok {
				return fmt.Errorf("poll channel closed unexpectedly")
			}

			if poll.Err != nil {
				log.Errorf("Error polling Traefik routers: %v", poll.Err)
				continue
			}

			// Skip if no changes to traefik routers
			if reflect.DeepEqual(s.cache, poll.Routers) {
				continue
			}

			log.Info("Changes detected in Traefik routers")

			// Update the cache
			s.cache = poll.Routers

			if err := s.processRouterChanges(ctx, poll.Routers); err != nil {
				log.Errorf("Failed to process router changes: %v", err)
				// Continue the loop instead of failing completely
			}
		}
	}
}

// processRouterChanges processes Traefik router changes and updates Cloudflare tunnel configuration
func (s *Service) processRouterChanges(ctx context.Context, routers []traefik.Router) error {
	// Build ingress rules from routers
	ingressResult, err := cloudflare.BuildIngressRules(routers, s.config)
	if err != nil {
		return fmt.Errorf("failed to build ingress rules: %w", err)
	}

	// Update tunnel configuration
	if err := cloudflare.SyncTunnelConfig(
		ctx,
		s.cloudflareClient,
		s.config.CloudflareAccountID,
		s.config.CloudflareTunnelID,
		ingressResult.Rules,
	); err != nil {
		return fmt.Errorf("failed to sync tunnel configuration: %w", err)
	}

	// Update DNS records
	if err := cloudflare.SyncDNSRecords(
		ctx,
		s.cloudflareClient,
		s.config.CloudflareTunnelID,
		ingressResult.Domains,
	); err != nil {
		return fmt.Errorf("failed to sync DNS records: %w", err)
	}

	// Cleanup stale DNS records if enabled
	if s.config.EnableDNSCleanup {
		log.Debug("Running DNS cleanup")
		if err := cloudflare.CleanupDNSRecords(
			ctx,
			s.cloudflareClient,
			s.config.CloudflareTunnelID,
			ingressResult.Domains,
		); err != nil {
			// Don't fail the whole sync if cleanup fails
			log.WithError(err).Warn("Failed to cleanup DNS records")
		}
	}

	return nil
}
