package forwardemail

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestClient_GetSieveScript(t *testing.T) {
	type request struct {
		domain string
		alias  string
		script string
	}

	tests := []struct {
		name     string
		wantErr  bool
		req      request
		response string
		want     *SieveScript
	}{
		{
			name:    "no script exists",
			wantErr: true,
			req: request{
				domain: "stark.com",
				alias:  "tony",
				script: "move_to_jarvis",
			},
		},
		{
			name: "sieve script exists and returned",
			req: request{
				domain: "stark.com",
				alias:  "tony",
				script: "move_to_jarvis",
			},
			response: `{
				"id": "move_to_jarvis",
				"name": "move_to_jarvis",
				"content": "still need to figure out format",
				"alias": "tony",
				"is_active": true,
				"priority": 0,
				"created_at": "2026-11-27T08:25:29.327Z",
				"updated_at": "2026-11-27T08:25:29.327Z",
  				"object": "sieve_script"
			}`,
			want: &SieveScript{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(writer, tt.response)
			}))
			defer server.Close()

			client := newTestClient(t, WithAPIURL(server.URL))

			got, err := client.GetSieveScript(context.Background(), tt.req.domain, tt.req.alias)
			if tt.wanrErr {
				if err == nil {
					t.Fatal("expected parse error for empty body, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("values are not the same: %s", diff)
			}
		})
	}
}
