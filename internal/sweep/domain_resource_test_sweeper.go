package sweep

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// SweepDomains will delete all domains starting with "tf-acc".
func SweepDomains(name string) *resource.Sweeper {
	return &resource.Sweeper{
		Name: name,
		F: func(_ string) error {
			client, err := ConfigureSweeperClient(name)
			if err != nil {
				// No API key means there is nothing to sweep, not a
				// failure: keyless runs (e.g. fork PRs in CI) skip.
				if errors.Is(err, errNoAPIKey) {
					return nil
				}
				return fmt.Errorf("error configuring Forward Email client: %w", err)
			}
			if client == nil {
				return nil
			}

			// Get all domains
			domains, err := client.GetDomains(context.Background())
			if err != nil {
				return fmt.Errorf("error getting domains: %w", err)
			}

			prefix := "tfacc"
			for _, domain := range domains {
				if strings.HasPrefix(domain.Name, prefix) {
					log.Printf("[INFO] Deleting domain: %s", domain.Name)
					err := client.DeleteDomain(context.Background(), domain.Name)
					if err != nil {
						log.Printf("[ERROR] Failed to delete domain %s: %s", domain.Name, err)
					}
				}
			}

			return nil
		},
	}
}
