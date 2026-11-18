// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	forwardemailAPIURL = "https://api.forwardemail.net"

	// MaxResponseBytes caps how much of a response body we read into memory,
	// so a misbehaving API can't OOM the provider.
	maxResponseBytes = 32 << 20

	// MaxErrorBodyBytes caps how much of an error body we retain on APIError.
	maxErrorBodyBytes = 4 << 10
)

// Client is the main client for interacting with the Forward Email API.
type Client struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// Option configures a Client. They are produced by With* helpers.
type Option func(*Client)

// WithHTTPClient lets callers supply their own *http.Client (for custom
// timeouts, proxies, tracing, etc.)
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithAPIURL lets callers override the default Forward Email API base URL.
func WithAPIURL(u string) Option { return func(c *Client) { c.apiURL = u } }

// NewClient returns a new Forward Email API Client.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	c := &Client{
		apiKey:     apiKey,
		apiURL:     forwardemailAPIURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.apiURL = strings.TrimSuffix(c.apiURL, "/")

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// validate checks that the client is properly configured.
func (c *Client) validate() error {
	if c.httpClient == nil {
		return ErrNilHTTPClient
	}
	return nil
}

// validateRequest validates the request parameters and client state.
func (c *Client) validateRequest(ctx context.Context, method, path string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if method == "" {
		return ErrEmptyMethod
	}
	if path == "" {
		return ErrEmptyPath
	}
	return nil
}

// validateContext is the shared preflight every API method runs: reject a
// nil or already-cancelled context before doing any work.
func validateContext(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	return ctx.Err()
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if err := c.validateRequest(ctx, method, path); err != nil {
		return nil, err
	}

	if body == nil {
		body = http.NoBody
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL+path, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.apiKey, "")

	return req, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("%w (limit %d bytes)", ErrResponseTooLarge, maxResponseBytes)
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return body, nil
	}

	if len(body) > maxErrorBodyBytes {
		body = body[:maxErrorBodyBytes]
	}

	return nil, &APIError{StatusCode: res.StatusCode, Body: body}
}
