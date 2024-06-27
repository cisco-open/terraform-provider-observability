// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package plugingenerator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"strings"
	"sync"
	"text/template"
	"unicode"

	"github.com/cisco-open/terraform-provider-observability/internal/api"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/sync/errgroup"
)

const (
	registeredObjectTypeJSON = "object_types.json"
	terraformBaseImportPath  = "github.com/hashicorp/terraform-plugin-framework"
	objectTemplatePath       = "/tools/plugingenerator/templates/"
	tmpFileName              = "resource_knowledge_object.go.tmpl"
	tmpRegistrarFileName     = "resource_gen_handlers.go.tmpl"
	resourceGenFuncFile      = "resource_gen_handlers.go"
)

// typeName: generic schema for
type SchemaTypeStore map[string]map[string]any

// RegisteredTypes represents the structure of the JSON data
type RegisteredTypes struct {
	Version                 string   `json:"version"`
	FullyQualifiedTypeNames []string `json:"fullyQualifiedTypeNames"`
}

type TypePayload struct {
	JSONSchema       `mapstructure:"jsonSchema, squash"`
	SecureProperties []string `mapstructure:"secureProperties"`
}

type JSONSchema struct {
	Description string                     `mapstructure:"description"`
	Properties  map[string]PropertyPayload `mapstructure:"properties"`
	Required    []string                   `mapstructure:"required"`
}

type PropertyPayload struct {
	Description string `mapstructure:"description"`
	Type        string `mapstructure:"type"`
}

type SchemaObjectFields struct {
	Description string
	Properties  []Property
}

type Property struct {
	Name        string
	Required    bool
	Sensitive   bool
	Type        string
	Description string
}

// PopulateSchemaTypeStore reads object_types.json file containing registered object types,
// retrieves schemas for each type from the API client, and populates a SchemaTypeStore.
// It returns the populated SchemaTypeStore or an error if any operation fails.
func PopulateSchemaTypeStore(appdClient *api.AppdClient) (SchemaTypeStore, error) {
	// read the file
	dataBytes, err := os.ReadFile(registeredObjectTypeJSON)
	if err != nil {
		return nil, fmt.Errorf("error during read file: %w", err)
	}

	// read the json object types we are interested in registering and generating code for
	var schemaTypes RegisteredTypes
	err = json.Unmarshal(dataBytes, &schemaTypes)
	if err != nil {
		return nil, fmt.Errorf("error during unmarshalling: %w", err)
	}

	// store the schemas key: fullyQualifiedTypeName: the actual schema

	var g errgroup.Group
	var m sync.Mutex
	schemaTypesStore := make(SchemaTypeStore)
	for _, fqtn := range schemaTypes.FullyQualifiedTypeNames {
		g.Go(func() error {
			schema, err := appdClient.GetType(fqtn)
			if err != nil {
				return fmt.Errorf("error during get type api call: %w", err)
			}

			var result map[string]any
			// unmarshall the schema
			err = json.Unmarshal(schema, &result)
			if err != nil {
				return fmt.Errorf("error during unmarshall: %w", err)
			}

			m.Lock()
			schemaTypesStore[fqtn] = result
			m.Unlock()
			return nil
		})
	}

	return schemaTypesStore, g.Wait()
}

// GenerateObjectFile generates Go files for Terraform resource objects based on the schema stored in SchemaTypeStore.
// The files are created using templates and written to the specified root repository path.
// This method returns an error if any step in the process fails.
func (sts SchemaTypeStore) GenerateObjectFile(rootRepoPath string) error {
	// generate the templateMapping
	for fqtn, schema := range sts {
		var buffer bytes.Buffer

		schemaObjectFields, err := getObjectSchemaFieldValues(schema)
		if err != nil {
			return fmt.Errorf("failed to extract schema type fields: %w", err)
		}

		// data will be passed in the template and will contain all information
		// needed to generate each file
		data := struct {
			Fqtn                    string
			TerraformBaseImportPath string
			PascalCaseObjectName    string
			SnakeCaseObjectName     string
			Payload                 *SchemaObjectFields
		}{
			Fqtn:                    fqtn,
			TerraformBaseImportPath: terraformBaseImportPath,
			PascalCaseObjectName:    transformTypeNameToPascalCase(fqtn),
			SnakeCaseObjectName:     transformTypeNameToSnakeCase(fqtn),
			Payload:                 schemaObjectFields,
		}

		// create template and parse the template
		templFilePath := rootRepoPath + objectTemplatePath + tmpFileName
		t := template.Must(template.New(tmpFileName).Funcs(
			template.FuncMap{
				"capFirstChar":     capFirstChar,
				"toLower":          strings.ToLower,
				"camelToSnakeCase": transformCamelToSnakeCase,
			}).ParseFiles(templFilePath))

		if err = t.Execute(&buffer, data); err != nil {
			return fmt.Errorf("failed to execute template %w", err)
		}

		// format the contents
		var formattedContent []byte
		formattedContent, err = format.Source(buffer.Bytes())
		if err != nil {
			return fmt.Errorf("failed to format source %w", err)
		}

		// open the generated file path
		snakeCaseFqtn := transformTypeNameToSnakeCase(fqtn)
		path := rootRepoPath + "/internal/provider/" + "resource_" + snakeCaseFqtn + ".go"
		writer, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}

		// generate the actual contents
		if _, err := writer.Write(formattedContent); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}

// GenerateRegistrarFunc generates a Go function that registers all generated resource objects.
// It uses a template to create the function and writes it to the specified root repository path.
// This method returns an error if any step in the process fails.
func (sts SchemaTypeStore) GenerateRegistrarFunc(rootRepoPath string) error {
	var data []string
	for fqtn := range sts {
		data = append(data, transformTypeNameToPascalCase(fqtn))
	}

	var buffer bytes.Buffer
	// create template and parse the template
	templFilePath := rootRepoPath + objectTemplatePath + tmpRegistrarFileName
	t := template.Must(template.New(tmpRegistrarFileName).ParseFiles(templFilePath))

	if err := t.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute template %w", err)
	}

	// format the contents
	var formattedContent []byte
	formattedContent, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format source %w", err)
	}

	// open the generated file path
	path := rootRepoPath + "/internal/provider/" + resourceGenFuncFile
	writer, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}

	// generate the actual contents
	if _, err := writer.Write(formattedContent); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

func getObjectSchemaFieldValues(schemaPayload map[string]any) (*SchemaObjectFields, error) {
	var payload TypePayload
	err := mapstructure.Decode(schemaPayload, &payload)
	if err != nil {
		return nil, err
	}

	var properties []Property
	for k, val := range payload.Properties {
		prop := Property{
			Name:        k,
			Description: val.Description,
			Type:        capFirstChar(val.Type),
		}
		properties = append(properties, prop)
	}

	// set the required fields
	for _, reqStrField := range payload.Required {
		for idx, propertyStruct := range properties {
			if reqStrField == propertyStruct.Name {
				properties[idx].Required = true
			}
		}
	}

	// set the sensitive fields
	for _, sensitiveStrField := range payload.SecureProperties {
		// extract only sensitiveFieldName
		// platform returns $.sensitiveFieldName
		sensitiveStrField = strings.Split(sensitiveStrField, ".")[1]
		for idx, propertyStruct := range properties {
			if sensitiveStrField == propertyStruct.Name {
				properties[idx].Sensitive = true
			}
		}
	}

	return &SchemaObjectFields{
		Description: payload.Description,
		Properties:  properties,
	}, nil
}

func transformTypeNameToSnakeCase(input string) string {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return input // return the original string if it doesn't have exactly one colon
	}

	return strings.ToLower(parts[0]) + "_" + strings.ToLower(parts[1])
}

func transformCamelToSnakeCase(input string) string {
	var result []rune
	for i, r := range input {
		if unicode.IsUpper(r) {
			// If it's not the first letter, add an underscore before the uppercase letter.
			if i > 0 {
				result = append(result, '_')
			}
			// Convert the uppercase letter to lowercase.
			r = unicode.ToLower(r)
		}
		result = append(result, r)
	}
	return string(result)
}

func transformTypeNameToPascalCase(input string) string {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return input // return the original string if it doesn't have exactly one colon
	}

	return capFirstChar(parts[0]) + capFirstChar(parts[1])
}

func capFirstChar(s string) string {
	runes := []rune(s)

	if len(runes) > 0 {
		if !unicode.IsUpper(runes[0]) {
			runes[0] = unicode.ToUpper(runes[0])
		}
	}

	return string(runes)
}
