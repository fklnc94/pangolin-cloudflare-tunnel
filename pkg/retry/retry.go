// Package retry provides utilities for retrying operations with exponential backoff.
package retry

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Operation is a function that can be retried
type Operation func() error

// Do attempts an operation multiple times with exponential backoff
func Do(maxRetries int, operation Operation) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			backoffDuration := time.Duration(1<<i) * time.Second
			log.WithError(err).Warnf("Operation failed, retrying in %v (%d/%d)...",
				backoffDuration, i+1, maxRetries)
			time.Sleep(backoffDuration)
		}
	}
	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}
