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

// authentication types
const (
	authMethodOAuth  = "oauth"
	headless         = "headless"
	servicePrincipal = "service-principal"
	// TODO add new types of authentication method here...
)

// oath related data
const (
	oauth2ClientID      = "default"
	oauth2AuthURISuffix = "oauth2/authorize" // API for obtaining authorization codes
	//nolint:gosec // This is not a hard coded secret
	oauth2TokenURISuffix = "oauth2/token" // API for exchanging the auth code for a token
	oauthRedirectURI     = "http://127.0.0.1:3101/callback"
	SHA256Hash           = "S256" // the SHA-256 hashing alorithm used to generate the code challenge for PKCE
)

const (
	typeAPIPath   = "/knowledge-store/v1/types/"
	objectAPIPath = "/knowledge-store/v1/objects/"
)
