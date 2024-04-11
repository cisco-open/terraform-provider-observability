// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

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
