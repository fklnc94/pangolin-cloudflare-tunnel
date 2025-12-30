// Package cloudflare provides Cloudflare API client and tunnel management.
package cloudflare

import (
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

// Client wraps the Cloudflare API client
type Client struct {
	api *cloudflare.API
}

// NewClient creates a new Cloudflare API client
func NewClient(apiToken string) (*Client, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudflare API client: %w", err)
	}

	return &Client{
		api: api,
	}, nil
}

// API returns the underlying Cloudflare API client
func (c *Client) API() *cloudflare.API {
	return c.api
}
