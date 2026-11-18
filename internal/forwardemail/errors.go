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
	ErrMissingAPIKey = errors.New("forwardemail: no api key provided")
	// ErrNilContext is returned when a nil context is passed to an API method.
	ErrNilContext = errors.New("forwardemail: context cannot be nil")
	// ErrEmptyMethod is returned when an empty HTTP method is passed to newRequest.
	ErrEmptyMethod = errors.New("forwardemail: http method cannot be empty")
	// ErrEmptyPath is returned when an empty request path is passed to newRequest.
	ErrEmptyPath = errors.New("forwardemail: request path cannot be empty")
	// ErrNilHTTPClient is returned when the client has a nil HTTPClient.
	ErrNilHTTPClient = errors.New("forwardemail: http client cannot be nil")
	// ErrResponseTooLarge is returned when an API response body exceeds the client's read cap.
	ErrResponseTooLarge = errors.New("forwardemail: api response too large")

	// ErrEmptyDomain is returned when a domain parameter is empty.
	ErrEmptyDomain = errors.New("forwardemail: domain cannot be empty")
	// ErrEmptyAlias is returned when an alias parameter is empty.
	ErrEmptyAlias = errors.New("forwardemail: alias cannot be empty")
	// ErrEmptyDomainName is returned when a domain name parameter is empty.
	ErrEmptyDomainName = errors.New("forwardemail: domain name cannot be empty")
	// ErrEmptyEmail is returned when an email parameter is empty.
	ErrEmptyEmail = errors.New("forwardemail: email cannot be empty")
	// ErrEmptyGroup is returned when a group parameter is empty.
	ErrEmptyGroup = errors.New("forwardemail: group cannot be empty")

	// ErrMissingEncryptionInput is returned when required input for encryption is missing.
	ErrMissingEncryptionInput = errors.New("forwardemail: missing required input for encryption")
	// ErrUnsupportedJSONValue is returned when a JSON value cannot be decoded into the expected type.
	ErrUnsupportedJSONValue = errors.New("forwardemail: unsupported json value")
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
