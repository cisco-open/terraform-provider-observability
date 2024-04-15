// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateObject is a method used to POST the knowledge store object
// based on the fullyQualifiedTypeName with the payload set in the body
// layerID which will be the tenant and layerType (TENANT/SOLUTION/...)
func (ac *AppdClient) CreateObject(fullyQualifiedTypeName, layerID, layerType string, body map[string]interface{}) error {
	url := ac.URL + objectAPIPath + fullyQualifiedTypeName

	bodyPayload := make(map[string]interface{})
	result, err := json.Marshal(bodyPayload)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(result)
	req, err := http.NewRequest(http.MethodPost, url, bodyReader) //nolint:noctx // To be removed in the future
	if err != nil {
		return fmt.Errorf("failed to create a request for %q: %w", url, err)
	}

	// Add headers
	contentType := "application/json"
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("Authorization", "Bearer "+ac.Token)

	req.Header.Add("layer-id", layerID)
	req.Header.Add("layer-type", layerType)

	// Do request
	resp, err := ac.APIClient.Do(req)
	if err != nil {
		return fmt.Errorf("%v request to %q failed: %w", http.MethodPost, req.URL.String(), err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("failed to POST request to %q (status %v): %s", req.URL.String(), resp.StatusCode, resp.Status)
	}

	return nil
}

// UpdateObject is a method used to PUT the knowledge store object
// based on the fullyQualifiedTypeName with the payload set in the body
// layerID which will be the tenant and layerType (TENANT/SOLUTION/...)
func (ac *AppdClient) UpdateObject(fullyQualifiedTypeName, layerID, layerType string, body map[string]interface{}) error {
	url := ac.URL + objectAPIPath + fullyQualifiedTypeName

	bodyPayload := make(map[string]interface{})
	result, err := json.Marshal(bodyPayload)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(result)
	req, err := http.NewRequest(http.MethodPut, url, bodyReader) //nolint:noctx // To be removed in the future
	if err != nil {
		return fmt.Errorf("failed to create a request for %q: %w", url, err)
	}

	// Add headers
	contentType := "application/json"
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("Authorization", "Bearer "+ac.Token)

	req.Header.Add("layer-id", layerID)
	req.Header.Add("layer-type", layerType)

	// Do request
	resp, err := ac.APIClient.Do(req)
	if err != nil {
		return fmt.Errorf("%v request to %q failed: %w", http.MethodPut, req.URL.String(), err)
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("failed to POST request to %q (status %v): %s", req.URL.String(), resp.StatusCode, resp.Status)
	}

	return nil
}

// GetObject is a method used to GET the knowledge store object
// based on the fullyQualifiedTypeName and objectID
// layerID which will be the tenant and layerType (TENANT/SOLUTION/...)
// If objectID ie an empty string this will result in a list of objects being returned
func (ac *AppdClient) GetObject(fullyQualifiedTypeName, objectID, layerID, layerType string) ([]byte, error) {
	var url string
	if objectID == "" {
		url = ac.URL + objectAPIPath + fullyQualifiedTypeName
	} else {
		url = ac.URL + objectAPIPath + fullyQualifiedTypeName + "/" + objectID
	}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody) //nolint:noctx // To be removed in the future
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for %q: %w", url, err)
	}

	// Add headers
	contentType := "application/json"

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("Authorization", "Bearer "+ac.Token)

	req.Header.Add("layer-id", layerID)
	req.Header.Add("layer-type", layerType)

	// Do request
	resp, err := ac.APIClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%v request to %q failed: %w", http.MethodGet, req.URL.String(), err)
	}

	// Unmarshal the contents
	var respBytes []byte
	defer resp.Body.Close()

	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response to %v to %q (status %v): %w", http.MethodGet, req.URL.String(), resp.StatusCode, err)
	}

	return respBytes, nil
}

// DeleteObject is a method used to DELETE the knowledge store object
// based on the fullyQualifiedTypeName and objectID
// layerID which will be the tenant and layerType (TENANT/SOLUTION/...)
func (ac *AppdClient) DeleteObject(fullyQualifiedTypeName, objectID, layerID, layerType string) error {
	url := ac.URL + objectAPIPath + fullyQualifiedTypeName + "/" + objectID

	req, err := http.NewRequest(http.MethodDelete, url, http.NoBody) //nolint:noctx // To be removed in the future
	if err != nil {
		return fmt.Errorf("failed to delete a request for %q: %w", url, err)
	}

	// Add headers
	contentType := "application/json"

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add("Authorization", "Bearer "+ac.Token)

	req.Header.Add("layer-id", layerID)
	req.Header.Add("layer-type", layerType)

	// Do request
	resp, err := ac.APIClient.Do(req)
	if err != nil {
		return fmt.Errorf("%v request to %q failed: %w", http.MethodDelete, req.URL.String(), err)
	}

	resp.Body.Close()
	return nil
}
