// Copyright 2024 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/apex/log"
)

type credentialsStruct struct {
	ClientID string `json:"Client ID"`
	Secret   string `json:"Secret"`
}

func (ac *AppdClient) servicePrincipalLogin() error {
	// read credentials file
	file := ac.SecretFile
	credentials, err := readJSONCredentials(file)
	if err != nil {
		return fmt.Errorf("failed to read credentials file %q: %w", file, err)
	}

	return servicePrincipalLogin(ac, credentials)
}

func servicePrincipalLogin(ac *AppdClient, credentials *credentialsStruct) error {
	// create a HTTP request
	uri, err := url.Parse(ac.URL)
	if err != nil {
		log.Fatalf("Failed to parse the url provided in context. URL: %s, err: %s", ac.URL, err)
	}
	uri.Path = "auth/" + ac.Tenant + "/default/oauth2/token"

	//nolint:noctx // To be removed in the future
	req, err := http.NewRequest("POST", uri.String(), strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return fmt.Errorf("failed to create a request for %q: %w", uri.String(), err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	req.SetBasicAuth(credentials.ClientID, credentials.Secret)

	// execute request
	client := ac.APIClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request auth (%q): %w", uri.String(), err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Login failed, status %q; details to follow", resp.Status)
	}

	// read body (success or error)
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed reading login response from %q: %w", uri.String(), err)
	}

	// update context with token
	var token appTokens
	err = json.Unmarshal(respBytes, &token)
	if err != nil {
		log.Errorf("failed to parse token: %v", err.Error())
		return err
	}
	log.Info("Login returned a valid token")
	ac.Token = token.AccessToken

	return nil
}

func readJSONCredentials(file string) (*credentialsStruct, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open the credentials file %q: %w", file, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read the credentials file %q: %w", file, err)
	}

	var credentials credentialsStruct
	if err = json.Unmarshal(data, &credentials); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file %q: %w", file, err)
	}

	return &credentials, nil
}
