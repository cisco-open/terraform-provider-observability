package api

import (
	"fmt"
	"io"
	"net/http"
)

// GetType is a method used to GET the type based on the fullyQualifiedTypeName
func (ac *AppdClient) GetType(fullyQualifiedTypeName string) ([]byte, error) {
	url := ac.URL + typeAPIPath + fullyQualifiedTypeName

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create a request for %q: %w", url, err)
	}

	// Add headers
	contentType := "application/json"

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)

	if ac.AuthMethod == authMethodOAuth {
		req.Header.Add("Authorization", "Bearer "+ac.Token)
	}

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
