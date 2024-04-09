package api

import (
	"fmt"
	"net/http"
)

type AppdClient struct {
	Username     string
	Password     string
	Tenant       string
	AuthMethod   string
	URL          string
	Token        string
	RefreshToken string
	SecretFile   string
	APIClient    *http.Client
}

func (ac *AppdClient) Login() error {
	var authErr error
	switch ac.AuthMethod {
	case authMethodOAuth:
		authErr = ac.oauthLogin()
	case headless:
		// TODO: implement the headless authentication using username and password
	case servicePrincipal:
		authErr = ac.servicePrincipalLogin()
	default:
		panic(fmt.Sprintf("bug: unhandled authentication method %q", ac.AuthMethod))
	}
	if authErr != nil {
		return authErr
	}

	// PROBLEM: we should return the login credentials to terraform for storing purposes ?
	return nil
}
