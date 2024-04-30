// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package api_test

import (
	"testing"

	"github.com/cisco-open/terraform-provider-observability/internal/api"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := api.GenerateCodeVerifier()
	if err != nil {
		t.Errorf("Error generating code verifier: %v", err)
	}
	if len(verifier) != 43 { // Base64URL encoding adds padding, so the length should be 43
		t.Errorf("Invalid code verifier length: expected 43, got %d", len(verifier))
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "test_verifier"
	challenge := api.GenerateCodeChallenge(verifier)
	expectedChallenge := "0Ku4rR8EgR1w3HyHLBCxVLtPsAAks5HOlpmTEt0XhVA"
	if challenge != expectedChallenge {
		t.Errorf("Generated code challenge does not match expected. Expected: %s, Got: %s", expectedChallenge, challenge)
	}
}

func TestValidateCodeVerifier(t *testing.T) {
	verifier := "test_verifier"
	challenge := api.GenerateCodeChallenge(verifier)
	err := api.ValidateCodeVerifier(verifier, challenge)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}
