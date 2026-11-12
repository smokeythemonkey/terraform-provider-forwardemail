// Package sweep provides utility functions for configuring and running
// test sweepers for Forward Email resources in Terraform.
package sweep

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/smokeythemonkey/terraform-provider-forwardemail/internal/forwardemail"
)

var errNoAPIKey = errors.New("failed to find FORWARDEMAIL_API_KEY")

// ConfigureSweeperClient initializes a Forward Email client using environment variables
// It returns the client and an error if initialization fails.
func ConfigureSweeperClient(name string) (*forwardemail.Client, error) {
	apiKey := os.Getenv("FORWARDEMAIL_API_KEY")
	if apiKey == "" {
		log.Printf("[INFO] Skipping %s sweeper - FORWARDEMAIL_API_KEY not set", name)
		return nil, errNoAPIKey
	}

	// Initialize the client
	client, err := forwardemail.NewClient(
		apiKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Forward Email client: %w", err)
	}

	return client, nil
}
