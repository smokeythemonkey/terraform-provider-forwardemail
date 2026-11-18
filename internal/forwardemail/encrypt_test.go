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

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/v1/encrypt" {
			t.Errorf("expected URL /v1/encrypt, got %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if string(body) != fmt.Sprintf(`{"input":%q}`, input) {
			t.Errorf("unexpected request body %q", string(body))
		}

		fmt.Fprint(w, response)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	got, err := c.EncryptRecord(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		var payload struct {
			Input string `json:"input"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("request body is not valid JSON: %v (raw: %s)", err, body)
		}
		if payload.Input != input {
			t.Errorf("server decoded input %q, want %q", payload.Input, input)
		}
		fmt.Fprint(w, `{"encrypted": "ok"}`)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	if _, err := c.EncryptRecord(context.Background(), input); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
