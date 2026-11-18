// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrMissingAPIKey is returned when no API key is provided.
	ErrMissingAPIKey = errors.New("no API key provided")
	// ErrNilContext is returned when a nil context is passed to an API method.
	ErrNilContext = errors.New("context cannot be nil")
	// ErrEmptyMethod is returned when an empty HTTP method is passed to newRequest.
	ErrEmptyMethod = errors.New("HTTP method cannot be empty")
	// ErrEmptyPath is returned when an empty request path is passed to newRequest.
	ErrEmptyPath = errors.New("request path cannot be empty")
	// ErrNilHTTPClient is returned when the client has a nil HTTPClient.
	ErrNilHTTPClient = errors.New("HTTP client cannot be nil")
	// ErrResponseTooLarge is returned when an API response body exceeds the client's read cap.
	ErrResponseTooLarge = errors.New("API response too large")

	// ErrEmptyDomain is returned when a domain parameter is empty.
	ErrEmptyDomain = errors.New("domain cannot be empty")
	// ErrEmptyAlias is returned when an alias parameter is empty.
	ErrEmptyAlias = errors.New("alias cannot be empty")
	// ErrEmptyDomainName is returned when a domain name parameter is empty.
	ErrEmptyDomainName = errors.New("domain name cannot be empty")
	// ErrEmptyEmail is returned when an email parameter is empty.
	ErrEmptyEmail = errors.New("email cannot be empty")
	// ErrEmptyGroup is returned when a group parameter is empty.
	ErrEmptyGroup = errors.New("group cannot be empty")

	// ErrMissingEncryptionInput is returned when required input for encryption is missing.
	ErrMissingEncryptionInput = errors.New("missing required input for encryption")
	// ErrUnsupportedJSONValue is returned when a JSON value cannot be decoded into the expected type.
	ErrUnsupportedJSONValue = errors.New("unsupported JSON value")
)

// APIError represents an error response from the Forward Email API.
type APIError struct {
	StatusCode int
	Body       []byte
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("forwardemail: API error %d: %s", e.StatusCode, http.StatusText(e.StatusCode))
	if len(e.Body) > 0 {
		msg += ": " + string(e.Body)
	}
	return msg
}

// IsStatus reports whether err is (or wraps) an *APIError with the given
// HTTP status code, matched through the error chain rather than the
// formatted message text.
func IsStatus(err error, code int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == code
}
