// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

//go:build unit

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cisco-open/terraform-provider-observability/internal/api"
)

const testType = "sample_type"
const expectedResponse = `{"type": "sample_type"}`

func TestGetType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"type": "sample_type"}`))
	}))
	defer srv.Close()

	// Create a new AppdClient with mock URL and credentials with a token
	ac := &api.AppdClient{
		URL:       srv.URL,
		Tenant:    tenant,
		APIClient: srv.Client(),
		Token:     token,
	}

	// Call the GetType method
	response, err := ac.GetType(testType)

	// Check for any errors
	if err != nil {
		t.Errorf("GetType returned an error: %v", err)
	}

	// Check if the response matches the expected JSON payload
	if string(response) != expectedResponse {
		t.Errorf("GetType returned incorrect response: got %s, want %s", response, expectedResponse)
	}
}
