package credhelper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDigitalOceanHelperNew(t *testing.T) {
	credHelper := NewDigitalOceanCredentialHelper(
		WithToken("dop_v1_12345678"),
		WithExpiry(42),
		WithReadWrite(),
	)
	if credHelper.token != "dop_v1_12345678" {
		t.Errorf("unexpected token: %q", credHelper.token)
	}
	if credHelper.ExpirySeconds != 42 {
		t.Errorf("unexpected ExpirySeconds: %d", credHelper.ExpirySeconds)
	}
	if !credHelper.ReadWrite {
		t.Errorf("unexpected ReadWrite: %t", credHelper.ReadWrite)
	}
}

func TestDigitalOceanHelperGet(t *testing.T) {
	credHelper := DigitalOceanCredentialHelper{
		ReadWrite:     true,
		ExpirySeconds: 3600,
		token:         "blah",
	}

	t.Run("unsupported registry", func(t *testing.T) {
		u, p, err := credHelper.Get("not-digitalocean.com")
		if err == nil || u != "" || p != "" {
			t.Errorf("expected error, got user=%q pass=%q", u, p)
		}
	})

	t.Run("happy path", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/registry/docker-credentials" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Header.Get("Authorization") != "Bearer blah" {
				t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
			}
			if r.URL.Query().Get("read_write") != "true" {
				t.Errorf("unexpected read_write: %s", r.URL.Query().Get("read_write"))
			}
			if r.URL.Query().Get("expiry_seconds") != "3600" {
				t.Errorf("unexpected expiry_seconds: %s", r.URL.Query().Get("expiry_seconds"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
  "auths": {
    "registry.digitalocean.com": {
      "auth": "YjdkMDNhNjk0N2IyMTdlZmI2ZjNlYzNiZDM1MDQ1ODI6YjdkMDNhNjk0N2IyMTdlZmI2ZjNlYzNiZDM1MDQ1ODIK"
    }
  }
}`))
		}))
		defer server.Close()

		// Override the client to use the test server
		t.Cleanup(func() {
			apiEndpoint = "api.digitalocean.com"
			client = http.DefaultClient
		})
		apiEndpoint, _ = strings.CutPrefix(server.URL, "https://")
		client = server.Client()

		user, pass, err := credHelper.Get("registry.digitalocean.com")
		if err != nil {
			t.Fatalf("failed to get credentials: %v", err)
		}
		if user != "b7d03a6947b217efb6f3ec3bd3504582" {
			t.Errorf("unexpected user: %q", user)
		}
		if pass != "b7d03a6947b217efb6f3ec3bd3504582\n" {
			t.Errorf("unexpected pass: %q", pass)
		}
	})

	t.Run("real api", func(t *testing.T) {
		credHelper := NewDigitalOceanCredentialHelper(WithExpiry(1)) // 1 second expiry to avoid polluting the dashboard
		if credHelper.token == "" {
			t.Skip("skipping integration test because DIGITALOCEAN_TOKEN is not set")
		}

		user, pass, err := credHelper.Get("registry.digitalocean.com")
		if err != nil {
			t.Fatalf("failed to get credentials: %v", err)
		}
		if user == "" {
			t.Errorf("unexpected user: %q", user)
		}
		if pass == "" {
			t.Errorf("unexpected pass: %q", pass)
		}
	})
}
