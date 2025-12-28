package traefik

import (
	"github.com/go-resty/resty/v2"
)

// Client wraps the Traefik API client
type Client struct {
	resty *resty.Client
}

// NewClient creates a new Traefik API client
func NewClient(apiEndpoint string) *Client {
	return &Client{
		resty: resty.New().SetBaseURL(apiEndpoint),
	}
}

// GetRouters fetches all HTTP routers from the Traefik API
func (c *Client) GetRouters() ([]Router, error) {
	var routers []Router
	_, err := c.resty.R().
		EnableTrace().
		SetResult(&routers).
		Get("/api/http/routers")

	if err != nil {
		return nil, err
	}

	return routers, nil
}
