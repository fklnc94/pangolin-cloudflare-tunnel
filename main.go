// Package main provides a service that automatically syncs Traefik router
// configurations to Cloudflare tunnels, including DNS record management.
// It polls the Traefik API periodically and updates Cloudflare tunnel
// configurations whenever changes are detected in the router configuration.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"hhftechnology/pangolin-cloudflare-tunnel/internal/cloudflare"
	"hhftechnology/pangolin-cloudflare-tunnel/internal/config"
	"hhftechnology/pangolin-cloudflare-tunnel/internal/sync"
	"hhftechnology/pangolin-cloudflare-tunnel/internal/traefik"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Configure logging
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	// Set log level based on environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log configuration summary
	log.WithFields(log.Fields{
		"zones":            len(cfg.CloudflareZones),
		"entrypoints":      cfg.TraefikEntrypoints,
		"skip_tls":         cfg.SkipTLSRoutes,
		"poll_interval":    cfg.PollInterval,
		"ignore_patterns":  len(cfg.IgnorePatterns),
		"enable_cleanup":   cfg.EnableDNSCleanup,
	}).Info("Configuration loaded")

	// Set up Cloudflare client
	cloudflareClient, err := cloudflare.NewClient(cfg.CloudflareToken)
	if err != nil {
		log.Fatalf("Failed to setup Cloudflare client: %v", err)
	}

	// Set up Traefik client
	traefikClient := traefik.NewClient(cfg.TraefikAPIEndpoint)

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalCh
		log.Infof("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Create and start sync service
	syncService := sync.NewService(cfg, traefikClient, cloudflareClient)

	log.Info("Starting synchronization service")
	if err := syncService.Run(ctx); err != nil {
		log.Fatalf("Error in sync service: %v", err)
	}

	log.Info("Service stopped gracefully")
}
