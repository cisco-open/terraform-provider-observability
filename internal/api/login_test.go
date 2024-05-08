// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

//go:build unit

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/cisco-open/terraform-provider-observability/internal/api"
)

const (
	servicePrincipal = "service-principal"
	oauth            = "oauth"
	secretsFileName  = "sample_secrets_"
	payload          = `{"Client ID": "sample_client_id", "Secret": "sample_secret"}`
	tenant           = "sample_tenant"
	token            = "sample_token"
)

func TestServicePrincipalLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"access_token": "%s"}`, token)
	}))
	defer srv.Close()

	// Create a temporary JSON file with sample credentials
	tmpfile, err := createTempJSONFile(payload)
	if err != nil {
		t.Errorf("Failed during creation of temporary json file: %v", err)
	}
	defer os.Remove(tmpfile)

	// Create a new AppdClient with mock URL and credentials
	ac := &api.AppdClient{
		URL:        srv.URL,
		Tenant:     tenant,
		AuthMethod: servicePrincipal,
		SecretFile: tmpfile,
		APIClient:  srv.Client(),
	}

	// Call the Login method
	err = ac.Login()

	// Check for any errors
	if err != nil {
		t.Errorf("Login returned an error: %v", err)
	}

	// Check if the access token was properly set
	if ac.Token != token {
		t.Errorf("Login failed to set the access token")
	}
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
func _TestOauthLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/" + tenant + "/default/oauth2/authorize":
			// Simulate authorization code grant flow
			// Redirect user to callback URL with mock authorization code
			redirectURL := r.FormValue("redirect_uri") + "?code=mockAuthorizationCode&scope=A&state=" + r.FormValue("state")
			http.Redirect(w, r, redirectURL, http.StatusFound)
		case "/callback":
			// Simulate callback handler
			// Respond with success message
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Login successful. You can close this browser window."))
		case "/auth/" + tenant + "/default/oauth2/token":
			// Simulate token exchange request
			// Return mock tokens
			mockTokens := map[string]any{
				"access_token": token,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(mockTokens)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not found"))
		}
	}))
	defer srv.Close()

	// Create a new AppdClient with mock URL and credentials
	ac := &api.AppdClient{
		URL:        srv.URL,
		Tenant:     tenant,
		AuthMethod: oauth,
		APIClient:  srv.Client(),
	}

	// Call the Login method
	err := ac.Login()

	// Check for any errors
	if err != nil {
		t.Errorf("Login returned an error: %v", err)
	}

	// Check if the access token was properly set
	if ac.Token != token {
		t.Errorf("Login failed to set the access token")
	}
}

// Helper function to create a temporary JSON file
func createTempJSONFile(contents string) (string, error) {
	tmpfile, err := os.CreateTemp("", secretsFileName)
	if err != nil {
		log.Fatalf("Failed to create temporary file: %v", err)
		return "", fmt.Errorf("Failed to create temporary file: %w", err)
	}
	defer tmpfile.Close()

	if _, err := tmpfile.WriteString(contents); err != nil {
		return "", fmt.Errorf("Failed to write to temporary file: %w", err)
	}

	return tmpfile.Name(), nil
}
