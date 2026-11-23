// Copyright Forward Email 2026
// SPDX-License-Identifier: MIT

package forwardemail

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_CreateDomainInvite(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		domain   string
		email    string
		group    string
		response string
		want     *Invite
	}{
		{
			name:    "no data",
			wantErr: true,
			domain:  "stark.com",
			email:   "tony@stark.com",
			group:   "admin",
		},
		{
			name:   "ok",
			domain: "stark.com",
			email:  "tony@stark.com",
			group:  "admin",
			response: `{
				"email": "tony@stark.com",
				"group": "admin",
				"id": "12345",
				"object": "invite",
				"created_at": "2024-12-10T18:14:31.378Z",
				"updated_at": "2024-12-10T18:14:31.378Z"
			}`,
			want: &Invite{
				Email:     "tony@stark.com",
				Group:     "admin",
				ID:        "12345",
				Object:    "invite",
				CreatedAt: parseTime(t, "2024-12-10T18:14:31.378Z"),
				UpdatedAt: parseTime(t, "2024-12-10T18:14:31.378Z"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				fmt.Fprint(w, tt.response)
			}))
			defer svr.Close()

			c := newTestClient(t, WithAPIURL(svr.URL))

			got, err := c.CreateDomainInvite(context.Background(), tt.domain, tt.email, tt.group)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected parse error for empty body, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != http.MethodPost {
				t.Errorf("method = %s, want POST", gotMethod)
			}
			wantPath := fmt.Sprintf("/v1/domains/%s/invites", tt.domain)
			if gotPath != wantPath {
				t.Errorf("path = %s, want %s", gotPath, wantPath)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("values are not the same %s", diff)
			}
		})
	}
}

func TestClient_CreateDomainInviteTrimsBeforeSending(t *testing.T) {
	var gotPath, gotEmail, gotGroup string
	var formErr error

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		formErr = r.ParseForm()
		gotEmail = r.PostForm.Get("email")
		gotGroup = r.PostForm.Get("group")
		fmt.Fprint(w, `{"email":"tony@stark.com","group":"admin"}`)
	}))
	defer svr.Close()

	c := newTestClient(t, WithAPIURL(svr.URL))

	if _, err := c.CreateDomainInvite(context.Background(), "  stark.com  ", "  tony@stark.com  ", "  admin  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if formErr != nil {
		t.Errorf("failed to parse form: %v", formErr)
	}
	if want := "/v1/domains/stark.com/invites"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if want := "tony@stark.com"; gotEmail != want {
		t.Errorf("email sent to API = %q, want %q", gotEmail, want)
	}
	if want := "admin"; gotGroup != want {
		t.Errorf("group sent to API = %q, want %q", gotGroup, want)
	}
}
