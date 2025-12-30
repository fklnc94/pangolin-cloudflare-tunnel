package cloudflare

import (
	"hhftechnology/pangolin-cloudflare-tunnel/internal/config"
	"hhftechnology/pangolin-cloudflare-tunnel/internal/traefik"

	"github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

// IngressResult contains the ingress rules and domains to process
type IngressResult struct {
	Rules   []cloudflare.UnvalidatedIngressRule
	Domains map[string]string // domain -> zone ID mapping
}

// BuildIngressRules creates Cloudflare ingress rules from Traefik routers
// Returns ingress rules and a map of domains to their zone IDs
func BuildIngressRules(routers []traefik.Router, cfg *config.Config) (*IngressResult, error) {
	ingress := []cloudflare.UnvalidatedIngressRule{}
	processedDomains := make(map[string]bool)
	domainToZone := make(map[string]string)

	for _, router := range routers {
		// Skip disabled routes
		if !router.IsEnabled() {
			continue
		}

		// Skip routes with TLS configured if SkipTLSRoutes is enabled
		if cfg.SkipTLSRoutes && router.HasTLSEnabled() {
			log.WithField("router", router.ServiceName).Debug("Skipping TLS-enabled router")
			continue
		}

		// Only use routes with one of the specified entrypoints
		if !router.HasMatchingEntrypoint(cfg.TraefikEntrypoints) {
			continue
		}

		// Parse domains from router rule
		routerDomains, err := router.ParseDomains()
		if err != nil {
			return nil, err
		}

		for _, domain := range routerDomains {
			// Check if domain should be ignored
			if cfg.ShouldIgnoreDomain(domain) {
				log.WithField("domain", domain).Info("Skipping ignored domain")
				continue
			}

			// Skip duplicate domains
			if processedDomains[domain] {
				log.WithField("domain", domain).Info("Skipping duplicate domain")
				continue
			}

			// Find the zone for this domain
			zone, err := cfg.GetZoneForDomain(domain)
			if err != nil {
				log.WithError(err).Warnf("Skipping domain %s: no matching zone", domain)
				continue
			}

			processedDomains[domain] = true
			domainToZone[domain] = zone.ZoneID

			log.WithFields(log.Fields{
				"domain":  domain,
				"zone":    zone.ZoneID,
				"service": cfg.TraefikServiceEndpoint,
			}).Info("Adding domain to tunnel configuration")

			// Create origin request config with TLS verification settings
			noTLSVerify := true
			originRequest := &cloudflare.OriginRequestConfig{
				HTTPHostHeader:   &domain,
				NoTLSVerify:      &noTLSVerify,
				OriginServerName: &domain,
			}

			ingress = append(ingress, cloudflare.UnvalidatedIngressRule{
				Hostname:      domain,
				Service:       cfg.TraefikServiceEndpoint,
				OriginRequest: originRequest,
			})
		}
	}

	// Add catch-all rule
	ingress = append(ingress, cloudflare.UnvalidatedIngressRule{
		Service: "http_status:404",
	})

	return &IngressResult{
		Rules:   ingress,
		Domains: domainToZone,
	}, nil
}
