// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"fmt"
	"io"
	"net/http"
)

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
