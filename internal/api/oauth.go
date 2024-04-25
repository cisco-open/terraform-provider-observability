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
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// appTokens is what the AppD backend returns when it hands back the tokens (in exchange for the authorization code)
type appTokens struct {
	AccessToken  string `json:"access_token"` // aka JWT token to make requests
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"` // this is what we use to get a fresh JWT token
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"` // e.g., bearer
}

// authCodes is the authorization that the oauth2 method provides
type authCodes struct {
	Code  string
	Scope string
	State string
}

// oauthErrorPayload is the structure returned by the auth/token endpoints on 4xx errors
type oauthErrorPayload struct {
	Error      string `json:"error"` // short error id, e.g., `invalid_client`
	ErrorDesc  string `json:"error_description"`
	ErrorHing  string `json:"error_hint"`
	StatusCode int    `json:"status_code"`
}

func (ac *AppdClient) oauthLogin() error {
	log.Infof("Starting OAuth authentication flow")

	// try refresh token if present
	if ac.RefreshToken != "" {
		// refresh and return if successful
		err := oauthRefreshToken(ac)
		if err == nil {
			log.Infof("Access token refreshed successfully")
			return nil
		}
		return err
	}

	// generate code verifier
	code, err := GenerateCodeVerifier()
	if err != nil {
		return err
	}

	// generate state
	state, err := GenerateCodeVerifier()
	if err != nil {
		return err
	}

	// prepare OAuth2 config
	conf := &oauth2.Config{
		ClientID:    oauth2ClientID,
		RedirectURL: oauthRedirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:   oauthURIWithSuffix(ac, oauth2AuthURISuffix),
			TokenURL:  oauthURIWithSuffix(ac, oauth2TokenURISuffix),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{"openid", "introspect_tokens", "offline_access"},
	}
	authCodeURL := conf.AuthCodeURL(state,
		Method(SHA256Hash),
		Challenge(GenerateCodeChallenge(code)),
	)

	// open browser to perform login, collect auth with a localhost http server
	authCode, err := getAuthorizationCodes(authCodeURL)
	if err != nil {
		return fmt.Errorf("login failed to obtain the authorization code: %w", err)
	}

	// verify nonce, must match
	if state != authCode.State {
		return fmt.Errorf("login failed: received auth state doesn't match (a session replay" +
			" or similar attack is likely in progress; please log out of all sessions!)")
	}

	// exchange auth code for token
	// TODO return token
	token, err := exchangeCodeForToken(conf, ac.APIClient, code, authCode)
	if err != nil {
		return fmt.Errorf("failed to exchange auth code for a token: %v", err.Error())
	}

	ac.Token = token.AccessToken
	fmt.Printf("Token is: %s", ac.Token)

	// PROBLEM: where will we store the token....
	return nil
}

func exchangeCodeForToken(conf *oauth2.Config, client *http.Client, codeVerifier string, authCode *authCodes) (*appTokens, error) {
	log.Infof("Exchanging authorization codes for access token")

	// prepare urlencoded data body
	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("client_id", "default")
	values.Add("code_verifier", codeVerifier)
	values.Add("code", authCode.Code)
	values.Add("redirect_uri", oauthRedirectURI)
	bodyReader := bytes.NewReader([]byte(values.Encode()))

	// create a POST HTTP request
	req, err := http.NewRequest("POST", conf.Endpoint.TokenURL, bodyReader) //nolint:noctx // To be removed in the future
	if err != nil {
		return nil, fmt.Errorf("failed to create a request %q: %v", conf.Endpoint.TokenURL, err.Error())
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST request to %q failed: %v", req.RequestURI, err.Error())
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("request failed, status %q; more info to follow", resp.Status)
	}

	// collect response body (whether success or error)
	var respBytes []byte
	defer resp.Body.Close()
	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response to POST to %q: %v", req.RequestURI, err.Error())
	}

	// parse response body in case of error (special parsing logic, tolerate non-JSON responses)
	if resp.StatusCode/100 != 2 {
		var errobj oauthErrorPayload

		// try to unmarshal JSON
		err := json.Unmarshal(respBytes, &errobj)
		if err != nil {
			// process as a string instead, ignore parsing error
			return nil, fmt.Errorf("error response: `%v`", bytes.NewBuffer(respBytes).String())
		}
		return nil, fmt.Errorf("error response: %+v", errobj)
	}

	// parse tokens
	var tokenObject appTokens
	if err := json.Unmarshal(respBytes, &tokenObject); err != nil {
		return nil, fmt.Errorf("failed to JSON parse the response as a token object: %v", err.Error())
	}

	return &tokenObject, nil
}

func oauthRefreshToken(cfg *AppdClient) error {
	log.Infof("Trying to get a new access token using the refresh token")

	// prepare urlencoded data body
	values := url.Values{}
	values.Add("client_id", oauth2ClientID)
	values.Add("redirect_uri", oauthRedirectURI)
	values.Add("grant_type", "refresh_token")
	values.Add("refresh_token", cfg.RefreshToken)
	bodyReader := bytes.NewReader([]byte(values.Encode()))

	// create a POST HTTP request
	tokenURI := oauthURIWithSuffix(cfg, oauth2TokenURISuffix)
	req, err := http.NewRequest("POST", tokenURI, bodyReader) //nolint:noctx // To be removed in the future
	if err != nil {
		return fmt.Errorf("failed to create a token refresh request %q: %w", tokenURI, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := cfg.APIClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST request to %q failed: %v", req.RequestURI, err.Error())
	}

	// log error if it occurred
	if resp.StatusCode/100 != 2 {
		// log error before trying to parse body, more processing later
		log.Errorf("Request failed, status %q; more info to follow", resp.Status)
		// fall through
	}

	var respBytes []byte
	defer resp.Body.Close()
	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed reading response to POST to %q: %v", req.RequestURI, err.Error())
	}

	// parse tokens
	var tokenObject appTokens
	if err := json.Unmarshal(respBytes, &tokenObject); err != nil {
		return fmt.Errorf("failed to JSON parse the response as a token object: %w", err)
	}

	// PROBLEM, do we need to store the new refresh token and access token ?
	// we need to return them
	return nil
}

func oauthURIWithSuffix(cfg *AppdClient, suffix string) string {
	uri, err := url.JoinPath(cfg.URL, "auth", cfg.Tenant, oauth2ClientID, suffix)
	if err != nil {
		log.Fatalf("unexpected failure constructing oauth2 endpoint URI: %v; terminating (likely a bug)", err)
	}
	return uri
}

func getAuthorizationCodes(uri string) (*authCodes, error) {
	// start http server to receive the auth callback
	callbackServer, respChan, err := startCallbackServer()
	if err != nil {
		return nil, fmt.Errorf("could not start a local http server for auth: %v", err.Error())
	}
	defer func() {
		_ = stopCallbackServer(callbackServer) // no check needed, error should be logged
	}()

	if err = openBrowser(uri); err != nil {
		log.Errorf("Failed to automatically launch browser auth window: %v", err)
		log.Errorf("Please visit the following URL to login\n%v\n", uri)
	}

	authCode := <-respChan // nb: blocks until a callback is received on localhost with the correct path

	return &authCode, nil
}

func startCallbackServer() (*http.Server, chan authCodes, error) {
	// construct a channel for the response
	respChan := make(chan authCodes)

	// start server at oauthRedirectUri
	urlStruct, err := url.Parse(oauthRedirectURI)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse callback URL %q: %w", oauthRedirectURI, err)
	}
	server := &http.Server{
		Addr: urlStruct.Host,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callbackHandler(respChan, w, r)
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil && err.Error() != "http: Server closed" {
			log.Errorf("Failed to start auth http server on %v: %v", server.Addr, err)
		}
	}()
	return server, respChan, nil
}

func stopCallbackServer(server *http.Server) error {
	if err := server.Close(); err != nil {
		err = fmt.Errorf("error stopping the auth http server on %v: %w", server.Addr, err)
		log.Errorf("%v", err)
		return err
	}

	log.Infof("Stopped the auth http server on %v", server.Addr)
	return nil
}

func callbackHandler(respChan chan authCodes, w http.ResponseWriter, r *http.Request) {
	// compute expected response path
	respURI, err := url.Parse(oauthRedirectURI)
	if err != nil {
		log.Errorf("Unexpected failure to obtain expected callback path (likely a bug): %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	callbackPath := respURI.Path

	uri, err := url.Parse(r.RequestURI)
	if err != nil {
		log.Errorf("Unexpected failure to parse callback path received (malformed request?): %v", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// reject all requests except the callback
	if uri.Path != callbackPath {
		log.Infof("Failing unexpected request for %q", uri.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	values := uri.Query()

	codes := authCodes{
		Code:  safeExtractFirstValue(values, "code"),
		Scope: safeExtractFirstValue(values, "scope"),
		State: safeExtractFirstValue(values, "state"),
	}

	fmt.Fprint(w, "Login successful. You can close this browser window.")

	respChan <- codes
}

func safeExtractFirstValue(queryValues url.Values, field string) string {
	// note: url.Values is simply a map[string][]string

	// extract values by field name
	qv := queryValues[field]
	if qv == nil {
		log.Errorf("expected a value for auth response %q, received none", field)
		return ""
	}

	// extract value
	l := len(qv)
	if l < 1 {
		log.Errorf("expected a value for auth response %q, received none", field)
		return ""
	}
	if l > 1 {
		// log name and count but not values (values are likely secret)
		log.Warnf("expected a single value for auth response %q, received %v", field, l)
		// fall through, get just the first value
	}
	return qv[0]
}

// openBrowser opens a browser window at the provided url. It also captures stdout message displayed
// by the command (if any: xdg-open in Linux says things like "Opening in existing browser session.") so
// that our stdout is not polluted (as it may be being captured for yaml/json parsing)
func openBrowser(uri string) error {
	// redirect browser's package stdout to a pipe, saving the original stdout
	orig := browser.Stdout
	r, w, _ := os.Pipe()
	browser.Stdout = w
	defer func() {
		browser.Stdout = orig
	}()

	// start browser
	browserErr := browser.OpenURL(uri) // check error later
	w.Close()                          // no more writing

	// copy the output in a separate goroutine so printing can't block indefinitely
	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		if err != nil {
			log.Warnf("Error capturing browser launch output: %v; ignoring it", err)
			// fall through
		}
		outChan <- buf.String()
	}()

	// collect and log any message displayed
	outMsg := strings.TrimSpace(<-outChan)
	if outMsg != "" {
		log.Infof("Browser launch: %v", outMsg)
	}

	return browserErr
}
