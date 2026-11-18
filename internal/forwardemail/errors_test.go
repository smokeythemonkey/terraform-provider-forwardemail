// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

var errConnectionRefused = errors.New("connection refused")

// Error() must surface the API's response body (already truncated to
// maxErrorBodyBytes by the client) so users see why a request failed,
// not just the status code.
func TestAPIErrorErrorIncludesBody(t *testing.T) {
	tests := []struct {
		name string
		err  *APIError
		want []string
	}{
		{
			name: "with body",
			err:  &APIError{StatusCode: http.StatusBadRequest, Body: []byte(`{"message":"invalid recipient"}`)},
			want: []string{"400", "Bad Request", `{"message":"invalid recipient"}`},
		},
		{
			name: "without body",
			err:  &APIError{StatusCode: http.StatusInternalServerError},
			want: []string{"500", "Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			for _, want := range tt.want {
				if !strings.Contains(msg, want) {
					t.Errorf("Error() = %q, want it to contain %q", msg, want)
				}
			}
		})
	}
}

// IsStatus must match an API status through the error chain, not by
// scanning the formatted error string.
func TestIsStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
		want bool
	}{
		{
			name: "wrapped APIError matches",
			err:  fmt.Errorf("failed to fetch account: %w", &APIError{StatusCode: http.StatusPaymentRequired, Body: []byte(`{"message":"payment required"}`)}),
			code: http.StatusPaymentRequired,
			want: true,
		},
		{
			name: "different status does not match",
			err:  &APIError{StatusCode: http.StatusBadRequest, Body: []byte(`{}`)},
			code: http.StatusPaymentRequired,
			want: false,
		},
		{
			name: "not found matches 404",
			err:  &APIError{StatusCode: http.StatusNotFound, Body: []byte(`{}`)},
			code: http.StatusNotFound,
			want: true,
		},
		{
			name: "non-API error",
			err:  errConnectionRefused,
			code: http.StatusPaymentRequired,
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			code: http.StatusPaymentRequired,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStatus(tt.err, tt.code); got != tt.want {
				t.Errorf("IsStatus(%v, %d) = %v, want %v", tt.err, tt.code, got, tt.want)
			}
		})
	}
}
