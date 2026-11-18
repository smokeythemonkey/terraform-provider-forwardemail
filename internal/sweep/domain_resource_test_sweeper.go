package sweep

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// SweepDomains deletes domains left behind by acceptance tests.
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

			for _, domain := range domains {
				if isSweepable(domain.Name) {
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

// isSweepable reports whether a domain name belongs to an acceptance test
// run and may be deleted: the tfacc/tf-acc prefixes used by older fixtures
// plus the fixed domain the current fixtures create.
func isSweepable(name string) bool {
	return strings.HasPrefix(name, "tfacc") ||
		strings.HasPrefix(name, "tf-acc") ||
		name == "forwardemailprovider.com"
}
