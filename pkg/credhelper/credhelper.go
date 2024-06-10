package credhelper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
)

const (
	doRegistry = "registry.digitalocean.com"
)

var (
	apiEndpoint = "api.digitalocean.com"
	client      = http.DefaultClient
)

type DigitalOceanCredHelper struct {
	// The duration in seconds that the returned registry credentials will be valid. If not set or 0, the credentials will not expire.
	ExpirySeconds int
	// By default, the registry credentials allow for read-only access. Set this query parameter to true to obtain read-write credentials.
	ReadWrite bool

	token string
}

func NewDigitalOceanCredHelper() DigitalOceanCredHelper {
	return DigitalOceanCredHelper{
		token: os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

func (d DigitalOceanCredHelper) Get(serverURL string) (string, string, error) {
	serverUrl, err := url.Parse("https://" + serverURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse registry URL: %w", err)
	}
	if serverUrl.Hostname() != doRegistry {
		return "", "", fmt.Errorf("not a Digital Ocean registry: %s", serverUrl.Hostname())
	}

	query := url.Values{}
	if d.ExpirySeconds > 0 {
		query.Set("expiry_seconds", strconv.Itoa(d.ExpirySeconds))
	}
	if d.ReadWrite {
		query.Set("read_write", "true")
	}

	api := url.URL{
		Scheme:   "https",
		Host:     apiEndpoint,
		Path:     "/v2/registry/docker-credentials",
		RawQuery: query.Encode(),
	}
	req, err := http.NewRequest("GET", api.String(), nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.token)
	res, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to get credentials from API: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to get credentials from API: %s", res.Status)
	}
	var creds dockerCredentialsResponse
	if err := json.NewDecoder(res.Body).Decode(&creds); err != nil {
		return "", "", fmt.Errorf("failed to decode credentials response: %w", err)
	}

	registry := serverUrl.Hostname()
	auth := creds.Auths[registry].Auth
	if len(auth) == 0 {
		return "", "", fmt.Errorf("no credentials for registry %q", registry)
	}
	colon := slices.Index(auth, ':')
	if colon == -1 {
		return "", "", fmt.Errorf("invalid credentials")
	}
	user := string(auth[:colon])
	pass := string(auth[colon+1:])
	return user, pass, nil
}

type dockerCredentialsResponse struct {
	Auths map[string]struct {
		Auth []byte `json:"auth"`
	} `json:"auths"`
}
