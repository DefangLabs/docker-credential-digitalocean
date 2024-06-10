package credhelper

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDigitalOceanHelper(t *testing.T) {
	credHelper := DigitalOceanCredHelper{
		ReadWrite:     true,
		ExpirySeconds: 3600,
		token:         "blah",
	}

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
	defer func() {
		apiEndpoint = "api.digitalocean.com"
		client = http.DefaultClient
	}()
	apiEndpoint, _ = strings.CutPrefix(server.URL, "https://")
	client = server.Client()

	user, pass, err := credHelper.Get("registry.digitalocean.com/defanglabs")
	if err != nil {
		t.Fatalf("failed to get credentials: %v", err)
	}
	if user != "b7d03a6947b217efb6f3ec3bd3504582" {
		t.Errorf("unexpected user: %q", user)
	}
	if pass != "b7d03a6947b217efb6f3ec3bd3504582\n" {
		t.Errorf("unexpected pass: %q", pass)
	}
}

func TestDigitalOceanHelperReal(t *testing.T) {
	credHelper := NewDigitalOceanCredHelper()
	user, pass, err := credHelper.Get("registry.digitalocean.com/defanglabs")
	if err != nil {
		t.Fatalf("failed to get credentials: %v", err)
	}
	if user == "" {
		t.Errorf("unexpected user: %q", user)
	}
	if pass == "" {
		t.Errorf("unexpected pass: %q", pass)
	}
}
