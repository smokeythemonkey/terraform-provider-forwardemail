// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient builds a client against the test server and fails the test
// cleanly if the constructor errors, instead of returning a nil client that
// panics at the first call site.
func newTestClient(t *testing.T, opts ...Option) *Client {
	t.Helper()
	c, err := NewClient("test-key", opts...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// A nil client returned alongside a swallowed constructor error must panic
// at first use rather than fail the test with a clear message.
func TestNewTestClientFailsCleanlyOnError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when using a nil client from a swallowed constructor error")
		}
	}()
	c, _ := NewClient("", WithAPIURL("http://unused"))
	_, _ = c.GetAccount(context.Background())
}

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

// Base URL must be normalized so a trailing slash doesn't double the path separator.
func TestNewClientAPIURLTrailingSlash(t *testing.T) {
	var gotPath string

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		fmt.Fprint(w, `{"object": "account"}`)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL+"/"))

	if _, err := c.GetAccount(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const want = "/v1/account"
	if gotPath != want {
		t.Errorf("request path with trailing-slash base URL = %q, want %q", gotPath, want)
	}
}

// The sentinel must be returned directly, not wrapped in an empty fmt.Errorf layer.
func TestNewClientMissingAPIKeyErrorIdentity(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
	if errors.Unwrap(err) != nil {
		t.Errorf(`NewClient("") must return ErrMissingAPIKey directly, got wrapped: %v`, err)
	}
}

// A response body exceeding the cap must fail closed with ErrResponseTooLarge.
func TestClient_ResponseTooLarge(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(bytes.Repeat([]byte("a"), maxResponseBytes+1)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	_, err := c.GetDomain(context.Background(), "stark.com")
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

// A success body exactly at the cap boundary must still be returned.
func TestClient_ResponseAtCapSucceeds(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(bytes.Repeat([]byte("a"), maxResponseBytes)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	req, err := c.newRequest(context.Background(), http.MethodGet, "/v1/account", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		t.Fatalf("body exactly at cap must be accepted, got %v", err)
	}
	if len(body) != maxResponseBytes {
		t.Errorf("body len = %d, want %d", len(body), maxResponseBytes)
	}
}

// An oversized error body must be truncated to maxErrorBodyBytes, not retained in full.
func TestClient_ErrorBodyTruncated(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(bytes.Repeat([]byte("e"), maxErrorBodyBytes*2)); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	_, err := c.GetDomain(context.Background(), "stark.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr := &APIError{}
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if len(apiErr.Body) != maxErrorBodyBytes {
		t.Errorf("error body len = %d, want %d", len(apiErr.Body), maxErrorBodyBytes)
	}
}
