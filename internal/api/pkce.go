// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/oauth2"
)

// GenerateCodeVerifier generates a random code verifier string
func GenerateCodeVerifier() (string, error) {
	const codeLenBytes = 32
	verifier := make([]byte, codeLenBytes) // Generate a 256-bit (32-byte) random string
	_, err := rand.Read(verifier)
	if err != nil {
		return "", err
	}
	return base64URLEncode(verifier), nil
}

// GenerateCodeChallenge generates a code challenge from a code verifier
func GenerateCodeChallenge(verifier string) string {
	hashed := sha256.Sum256([]byte(verifier))
	return base64URLEncode(hashed[:])
}

// ValidateCodeVerifier validates a code verifier against a code challenge
func ValidateCodeVerifier(verifier, challenge string) error {
	calculatedChallenge := GenerateCodeChallenge(verifier)
	if calculatedChallenge != challenge {
		return errors.New("invalid code verifier")
	}
	return nil
}

// Method sets the appropriate hashing algorithm used in generating the code challenge
func Method(hashingType string) oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("code_challenge_method", hashingType)
}

// Verifier sets the appropriate verifier
func Verifier(verifier string) oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("code_verifier", verifier)
}

// Challenge sets the appropriate code challenge
func Challenge(challenge string) oauth2.AuthCodeOption {
	return oauth2.SetAuthURLParam("code_challenge", challenge)
}

func base64URLEncode(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
