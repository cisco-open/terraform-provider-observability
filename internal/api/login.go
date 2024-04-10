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
