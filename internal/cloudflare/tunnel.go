package cloudflare

import (
	"context"
	"fmt"

	"hhftechnology/pangolin-cloudflare-tunnel/pkg/retry"

	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

// SyncTunnelConfig updates the Cloudflare tunnel configuration with new ingress rules
func SyncTunnelConfig(ctx context.Context, client *Client, accountID, tunnelID string, ingress []cloudflare.UnvalidatedIngressRule) error {
	return retry.Do(3, func() error {
		// Get current tunnel config
		accountRC := cloudflare.AccountIdentifier(accountID)
		tunnelConfig, err := client.api.GetTunnelConfiguration(ctx, accountRC, tunnelID)
		if err != nil {
			return fmt.Errorf("failed to get current tunnel configuration: %w", err)
		}

		// Update config with new ingress rules
		tunnelConfig.Config.Ingress = ingress
		_, err = client.api.UpdateTunnelConfiguration(ctx, accountRC, cloudflare.TunnelConfigurationParams{
			TunnelID: tunnelID,
			Config:   tunnelConfig.Config,
		})
		if err != nil {
			return fmt.Errorf("failed to update tunnel configuration: %w", err)
		}

		log.Info("Tunnel configuration updated successfully")
		return nil
	})
}
