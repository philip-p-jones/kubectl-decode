/*
Copyright 2024 Philip Jones

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package format

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Global variable to store the detected format
var detectedFormat string

// AssertFormat determines if the input is JSON or YAML based on content.
// If the format has already been determined by ParseInput, it returns the cached value.
func AssertFormat(input []byte) string {
	// If the format is already determined, return it
	if detectedFormat != "" {
		return detectedFormat
	}

	// Fallback to heuristic detection if the format has not been set
	inputStr := strings.TrimSpace(string(input))
	if len(inputStr) > 0 {
		if inputStr[0] == '{' || inputStr[0] == '[' {
			return "json"
		}
		return "yaml"
	}

	// Default case if no format is detected
	return ""
}

// Helper function to convert map[interface{}]interface{} to map[string]interface{}
// This is necessary because YAML unmarshals into a more generic type
func convertMap(m interface{}) interface{} {
	switch v := m.(type) {
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			newMap[fmt.Sprintf("%v", key)] = convertMap(value)
		}
		return newMap
	case []interface{}:
		for i, item := range v {
			v[i] = convertMap(item)
		}
		return v
	default:
		return m
	}
}

// DecodeBase64Values decodes base64 encoded values in the "data" field and merges them into "stringData"
func DecodeBase64Values(inputData map[string]interface{}) error {
	data, hasData := inputData["data"].(map[string]interface{})
	stringData, hasStringData := inputData["stringData"].(map[string]interface{})

	if !hasStringData {
		// If no stringData exists, initialize it
		stringData = make(map[string]interface{})
		inputData["stringData"] = stringData
	}

	// Iterate over the "data" field and decode each base64 value
	if hasData {
		for key, value := range data {
			strValue, ok := value.(string)
			if !ok {
				return fmt.Errorf("data field contains a non-string value for key %s", key)
			}

			decodedValue, err := base64.StdEncoding.DecodeString(strValue)
			if err != nil {
				return fmt.Errorf("failed to decode base64 value for key %s: %v", key, err)
			}

			// If the key does not exist in stringData, add the decoded value
			if _, exists := stringData[key]; !exists {
				stringData[key] = string(decodedValue)
			}
		}

		// After decoding, remove the "data" field to match the expected output format
		delete(inputData, "data")
	}

	return nil
}

// ParseInput parses the input data into a map[string]interface{}.
// It also sets the detected format (json or yaml) once parsing is successful.
func ParseInput(input []byte) (map[string]interface{}, error) {
	// Attempt to unmarshal as JSON first
	var jsonData map[string]interface{}
	if err := json.Unmarshal(input, &jsonData); err == nil {
		detectedFormat = "json" // Set the global format to JSON
		return jsonData, nil
	}

	// If JSON parsing fails, attempt to unmarshal as YAML
	var yamlData map[interface{}]interface{}
	if err := yaml.Unmarshal(input, &yamlData); err == nil {
		detectedFormat = "yaml" // Set the global format to YAML
		return convertMap(yamlData).(map[string]interface{}), nil
	}

	// If both JSON and YAML parsing fail, return an error
	return nil, fmt.Errorf("failed to parse input as JSON or YAML. Use --output yaml|json to decode data.")
}

// OutputResult prints the output in the appropriate format (YAML or JSON).
// It relies on AssertFormat to determine how the data should be encoded.
func OutputResult(input []byte, inputData map[string]interface{}) error {
	// Decode base64 values in the "data" field
	if err := DecodeBase64Values(inputData); err != nil {
		return err
	}

	format := AssertFormat(input)

	if format == "json" {
		output, err := json.MarshalIndent(inputData, "", "  ")
		if err != nil {
			return fmt.Errorf("error encoding JSON: %v", err)
		}
		os.Stdout.Write(output)
	} else if format == "yaml" {
		output, err := yaml.Marshal(inputData)
		if err != nil {
			return fmt.Errorf("error encoding YAML: %v", err)
		}
		os.Stdout.Write(output)
	} else {
		return fmt.Errorf("unknown format: %s", format)
	}

	return nil
}
