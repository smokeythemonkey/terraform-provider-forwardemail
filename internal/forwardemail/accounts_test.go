// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestClient_GetAccount(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     *Account
	}{
		{
			name: "no data",
		},
		{
			name: "ok",
			response: `{
  				"plan": "enhanced_protection",
  				"email": "tony@stark.com",
  				"full_email": "tony@stark.com",
  				"display_name": "tony@stark.com",
  				"last_locale": "en",
  				"address_country": "None",
  				"id": "59ad551ae6fb4a4c53427ca38079f029",
  				"object": "user",
  				"locale": "en",
  				"created_at": "2023-09-21T20:14:27.964Z",
  				"updated_at": "2023-10-07T17:47:54.595Z",
  				"address_html": ""
			}`,
			want: &Account{
				Plan:           "enhanced_protection",
				Email:          "tony@stark.com",
				FullEmail:      "tony@stark.com",
				DisplayName:    "tony@stark.com",
				LastLocale:     "en",
				AddressCountry: "None",
				ID:             "59ad551ae6fb4a4c53427ca38079f029",
				Object:         "user",
				Locale:         "en",
				CreatedAt:      parseTime(t, "2023-09-21T20:14:27.964Z"),
				UpdatedAt:      parseTime(t, "2023-10-07T17:47:54.595Z"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, tt.response)
			}))
			defer svr.Close()

			c := newTestClient(t, WithAPIURL(svr.URL))

			got, _ := c.GetAccount(context.Background())
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("values are not the same %s", diff)
			}
		})
	}
}

// fakeTB implements the small part of testing.TB that parseTime needs, so
// tests can prove parseTime fails loudly instead of returning zero time.
type fakeTB struct {
	fatalCalled bool
	msg         string
}

func (f *fakeTB) Helper() {}
func (f *fakeTB) Fatalf(format string, args ...any) {
	f.fatalCalled = true
	f.msg = fmt.Sprintf(format, args...)
}

// parseTB is the minimal interface parseTime needs, so tests can substitute
// a fake and prove parseTime fails loudly on a bad timestamp.
type parseTB interface {
	Helper()
	Fatalf(format string, args ...any)
}

// A malformed fixture timestamp must fail the test loudly, not silently
// parse to the zero time and produce confusing diffs.
func TestParseTimeFailsLoudlyOnBadInput(t *testing.T) {
	fake := &fakeTB{}
	got := parseTime(fake, "not-a-timestamp")
	if !fake.fatalCalled {
		t.Error("parseTime with an invalid timestamp must call Fatalf, not return zero time silently")
	}
	if !got.IsZero() {
		t.Errorf("parseTime returned %v, want zero time after Fatalf", got)
	}
}

func parseTime(tb parseTB, str string) time.Time {
	tb.Helper()
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		tb.Fatalf("invalid fixture timestamp %q: %v", str, err)
	}

	return t
}
