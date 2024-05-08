// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cisco-open/terraform-provider-observability/internal/api"
)

var (
	mockObjectStore = make(map[string]string)
)

const (
	sampleObjectType = "sample_object_type"
	sampleLayerType  = "TENANT"
	sampleLayerID    = "sample_tenant"
	sampleObjectID   = "test"
)

//nolint:gocyclo // To be removed in the future
func TestCRUDObject(t *testing.T) {
	// Create a new test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		layerID := r.Header.Get("layer-id")
		layerType := r.Header.Get("layer-type")
		key := sampleObjectType + "/" + layerID + "/" + layerType + "/" + sampleObjectID

		switch r.Method {
		case http.MethodPost:
			// CreateObject logic
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mockObjectStore[key] = string(payload)
			w.WriteHeader(http.StatusOK)

		case http.MethodGet:
			// GetObject logic
			// Check if object is present
			if _, ok := mockObjectStore[key]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(mockObjectStore[key]))
		case http.MethodPut:
			// UpdateObject logic
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Check if object is present for update
			if _, ok := mockObjectStore[key]; !ok {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mockObjectStore[key] = string(payload)
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			// DeleteObject logic
			delete(mockObjectStore, key)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	// Create a new AppdClient with the mock server's URL
	ac := &api.AppdClient{
		URL:       srv.URL,
		APIClient: srv.Client(),
		Token:     token,
	}

	// Run subtest for CreateObject
	t.Run("CreateObject", func(t *testing.T) {
		// Call the CreateObject method

		err := ac.CreateObject(sampleObjectType, sampleLayerID, sampleLayerType,
			[]byte(`{"cloudType": "AWS", "connectionName": "just-terraform-testing", "region": "us-east-2"}`))

		// Check for any errors
		if err != nil {
			t.Errorf("CreateObject returned an unexpected error: %v", err)
		}
	})

	// Run subtest GetObject
	t.Run("GetObjectBeforeUpdate", func(t *testing.T) {
		// Call the UpdateObject method
		response, err := ac.GetObject(sampleObjectType, sampleObjectID, sampleLayerID, sampleLayerType)

		// Check for any errors
		if err != nil {
			t.Errorf("UpdateObject returned an unexpected error: %v", err)
		}

		// Check if object matches the expected payload
		expectedPayload := `{"cloudType": "AWS", "connectionName": "just-terraform-testing", "region": "us-east-2"}`

		if string(response) != expectedPayload {
			t.Errorf("Got object: %s but expected %s", payload, expectedPayload)
		}
	})

	// Run subtest for UpdateObject
	t.Run("UpdateObject", func(t *testing.T) {
		// Call the UpdateObject method
		err := ac.UpdateObject(sampleObjectType, sampleObjectID, sampleLayerID, sampleLayerType,
			[]byte(`{"cloudType": "GCP", "connectionName": "just-terraform-testing", "region": "us-west-2"}`))

		// Check for any errors
		if err != nil {
			t.Errorf("UpdateObject returned an unexpected error: %v", err)
		}

		// Check if object matches the expected payload
		expectedPayload := `{"cloudType": "GCP", "connectionName": "just-terraform-testing", "region": "us-west-2"}`
		key := sampleObjectType + "/" + sampleLayerID + "/" + sampleLayerType + "/" + sampleObjectID
		payload := mockObjectStore[key]
		if payload != expectedPayload {
			t.Errorf("Got object: %s but expected %s", payload, expectedPayload)
		}
	})

	// Run subtest GetObject
	t.Run("GetObjectAfterUpdate", func(t *testing.T) {
		// Call the UpdateObject method
		response, err := ac.GetObject(sampleObjectType, sampleObjectID, sampleLayerID, sampleLayerType)

		// Check for any errors
		if err != nil {
			t.Errorf("UpdateObject returned an unexpected error: %v", err)
		}

		// Check if object matches the expected payload
		expectedPayload := `{"cloudType": "GCP", "connectionName": "just-terraform-testing", "region": "us-west-2"}`

		if string(response) != expectedPayload {
			t.Errorf("Got object: %s but expected %s", payload, expectedPayload)
		}
	})

	// Run subtest DeleteObject
	t.Run("DeleteObject", func(t *testing.T) {
		// Call the UpdateObject method
		err := ac.DeleteObject(sampleObjectType, sampleObjectID, sampleLayerID, sampleLayerType)

		// Check for any errors
		if err != nil {
			t.Errorf("DeleteObject returned an unexpected error: %v", err)
		}

		// Check if object is missing after deletion
		key := sampleObjectType + "/" + sampleLayerID + "/" + sampleLayerType + "/" + sampleObjectID

		if obj, ok := mockObjectStore[key]; ok {
			t.Errorf("Object: %s still present but it should be missing", obj)
		}
	})
}
