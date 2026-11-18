// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

var errConnectionRefused = errors.New("connection refused")

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
