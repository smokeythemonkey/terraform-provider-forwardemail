// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		opts      []Option
		wantURL   string
		wantError error
	}{
		{
			name:      "empty api key returns error",
			apiKey:    "",
			wantError: ErrMissingAPIKey,
		},
		{
			name:    "with api key uses default URL",
			apiKey:  "test-key",
			wantURL: "https://api.forwardemail.net",
		},
		{
			name:    "with custom URL",
			apiKey:  "test-key",
			opts:    []Option{WithAPIURL("https://google.com")},
			wantURL: "https://google.com",
		},
		{
			name:      "nil http client returns error",
			apiKey:    "test-key",
			opts:      []Option{WithHTTPClient(nil)},
			wantError: ErrNilHTTPClient,
		},
		{
			name:    "with custom http client",
			apiKey:  "test-key",
			opts:    []Option{WithHTTPClient(&http.Client{Timeout: 10 * time.Second})},
			wantURL: "https://api.forwardemail.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.apiKey, tt.opts...)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.apiURL != tt.wantURL {
				t.Fatalf("expected apiURL %q, got %q", tt.wantURL, got.apiURL)
			}
			if got.apiKey != tt.apiKey {
				t.Fatalf("expected apiKey %q, got %q", tt.apiKey, got.apiKey)
			}
			if got.httpClient == nil {
				t.Fatal("expected non-nil httpClient")
			}
		})
	}
}
