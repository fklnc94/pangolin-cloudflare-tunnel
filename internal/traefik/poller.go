package traefik

import (
	"context"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

// Response represents the response from polling Traefik routers
type Response struct {
	Routers []Router
	Err     error
}

// Poll periodically polls the Traefik API for router configurations.
// It returns a channel that emits router configurations whenever polled.
// The polling interval is specified with a small random jitter added to avoid thundering herd.
func Poll(ctx context.Context, client *Client, interval time.Duration) <-chan Response {
	ch := make(chan Response)

	go func() {
		defer close(ch)

		jitterSource := rand.New(rand.NewSource(time.Now().UnixNano()))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				routers, err := client.GetRouters()

				response := Response{
					Routers: routers,
					Err:     err,
				}

				if err != nil {
					log.WithError(err).Error("Failed to poll Traefik routers")
				}

				ch <- response

				// Add jitter to avoid thundering herd
				jitter := time.Duration(jitterSource.Int63n(int64(interval) / 2))
				select {
				case <-ctx.Done():
					return
				case <-time.After(jitter):
					// This just adds a random delay
				}
			}
		}
	}()

	return ch
}
