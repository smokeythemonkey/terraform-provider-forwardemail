// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_EncryptRecord(t *testing.T) {
	const input = "forward-email-site-verification=abc123"

	response := `{
		"encrypted": "forward-email-site-verification=$2a$10$abc123"
	}`
	want := &EncryptionResponse{
		Encrypted: "forward-email-site-verification=$2a$10$abc123",
	}

	var gotMethod, gotPath, gotBody string
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, err := io.ReadAll(r.Body)
		if err != nil {
			gotBody = "read error: " + err.Error()
		} else {
			gotBody = string(body)
		}
		fmt.Fprint(w, response)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	got, err := c.EncryptRecord(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/v1/encrypt" {
		t.Errorf("path = %s, want /v1/encrypt", gotPath)
	}
	if wantBody := fmt.Sprintf(`{"input":%q}`, input); gotBody != wantBody {
		t.Errorf("request body = %q, want %q", gotBody, wantBody)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("values are not the same %s", diff)
	}
}

func TestClient_EncryptRecordMissingInput(t *testing.T) {
	c := newTestClient(t)

	_, err := c.EncryptRecord(context.Background(), "")
	if !errors.Is(err, ErrMissingEncryptionInput) {
		t.Fatalf("expected error %v, got %v", ErrMissingEncryptionInput, err)
	}
}

// Body must be built with json.Marshal so quotes in the input are escaped.
func TestClient_EncryptRecordEscapesJSONInput(t *testing.T) {
	const input = `she said "hello"`

	var gotBody, decoded string
	var payloadErr error
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			gotBody = "read error: " + err.Error()
		} else {
			gotBody = string(body)
		}
		var payload struct {
			Input string `json:"input"`
		}
		payloadErr = json.Unmarshal(body, &payload)
		decoded = payload.Input
		fmt.Fprint(w, `{"encrypted": "ok"}`)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	if _, err := c.EncryptRecord(context.Background(), input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payloadErr != nil {
		t.Errorf("request body is not valid JSON: %v (raw: %s)", payloadErr, gotBody)
	}
	if decoded != input {
		t.Errorf("server decoded input %q, want %q", decoded, input)
	}
}

// API errors must stay wrapped so errors.As can inspect the status code.
func TestClient_EncryptRecordAPIErrorChain(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	_, err := c.EncryptRecord(context.Background(), "input")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error must wrap *APIError so errors.As can inspect the status, got %v", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, apiErr.StatusCode)
	}
}

// Parse failures must wrap the encoding/json cause for diagnosability.
func TestClient_EncryptRecordParseErrorChain(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{not json`)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	_, err := c.EncryptRecord(context.Background(), "input")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var synErr *json.SyntaxError
	if !errors.As(err, &synErr) {
		t.Fatalf("error must wrap the encoding/json cause so the failure is diagnosable, got %v", err)
	}
}
