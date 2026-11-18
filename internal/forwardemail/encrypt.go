// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

// Forward Email allows you to encrypt records even on the free plan at no cost. As highly requested in a Privacy Guides discussion and on our GitHub issues we've added this.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// EncryptionResponse represents the API Response for encrypting a TXT record.
type EncryptionResponse struct {
	Encrypted string `json:"encrypted"`
}

// EncryptRecord allows you to encrypt a TXT record value so that you can safely add it to your DNS without exposing the value in plaintext.
func (c *Client) EncryptRecord(ctx context.Context, input string) (*EncryptionResponse, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if input == "" {
		return nil, ErrMissingEncryptionInput
	}

	payload, err := json.Marshal(struct {
		Input string `json:"input"`
	}{Input: input})
	if err != nil {
		return nil, fmt.Errorf("forwardemail: failed to marshal encrypt request: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/v1/encrypt", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("forwardemail: failed to create encrypt request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("forwardemail: failed to encrypt record: %w", err)
	}

	var encryptionResponse EncryptionResponse
	if err := json.Unmarshal(res, &encryptionResponse); err != nil {
		return nil, fmt.Errorf("forwardemail: failed to parse encrypt response: %w", err)
	}

	return &encryptionResponse, nil
}
