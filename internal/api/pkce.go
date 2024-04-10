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
