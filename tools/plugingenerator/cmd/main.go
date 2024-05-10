// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"log"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/cisco-open/terraform-provider-observability/internal/api"
	"github.com/cisco-open/terraform-provider-observability/tools/plugingenerator"
)

var (
	tenant          string
	url             string
	secretsFileName string
	rootOfTheRepo   string
)

const (
	servicePrincipal         = "service-principal"
	registeredObjectTypeJSON = "object_types.json"
)

func main() {
	// parse the flags
	flag.StringVar(&tenant, "tenant-id", "0eb4e853-34fb-4f77-b3fc-b9cd3b462366", "tenant id used in the environment")
	flag.StringVar(&url, "url", "https://aiops-dev.saas.appd-test.com", "url for the api environment")
	//nolint:lll // To be removed
	flag.StringVar(&secretsFileName, "secret-file-name", "aiops_secret.json", "file name containing the secrets used to authenticate, it should be present in your home directory")
	flag.StringVar(&rootOfTheRepo, "repo-root", "", "path to the root of the repo")
	flag.Parse()

	// Get current user information
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Construct the full path
	secretsFilePath := filepath.Join(currentUser.HomeDir, secretsFileName)

	appdClient := &api.AppdClient{
		AuthMethod: servicePrincipal,
		Tenant:     tenant,
		URL:        url,
		SecretFile: secretsFilePath,
		APIClient:  http.DefaultClient,
	}

	// there is no point in going forward, just exit
	err = appdClient.Login()
	if err != nil {
		log.Fatal(err.Error())
	}

	schemaTypesStore, err := plugingenerator.PopulateSchemaTypeStore(appdClient)
	if err != nil {
		log.Fatal(err.Error())
	}

	// generate all the resources based on type schema for each object
	err = schemaTypesStore.GenerateObjectFile(rootOfTheRepo)
	if err != nil {
		log.Fatal(err.Error())
	}

	// generate the function used to register all the previously created resources
	err = schemaTypesStore.GenerateRegistrarFunc(rootOfTheRepo)
	if err != nil {
		log.Fatal(err.Error())
	}
}
