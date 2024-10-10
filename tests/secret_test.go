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
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os/exec"
	"strings"
	"testing"
)

// Helper function to run kubectl-decode with the given input and format
func runKubectlUnsecret(input string, format string) (string, error) {
	cmd := exec.Command("../dist/kubectl-decode_linux_amd64_v1/kubectl-decode")
	var out bytes.Buffer
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

// Test for valid JSON Secret input
func TestParseJsonSecret(t *testing.T) {
	input := `{
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": "my-secret"},
        "data": {
            "key1": "dmFsdWUx",
            "key2": "dmFsdWUy"
        },
        "stringData": {"key3": "value3"}
    }`

	expectedOutput := `{
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": "my-secret"},
        "stringData": {
            "key1": "value1",
            "key2": "value2",
            "key3": "value3"
        }
    }`

	output, err := runKubectlUnsecret(input, "json")
	if err != nil {
		t.Fatalf("Error running kubectl-decode: %v", err)
	}
	compareJsonOutput(t, output, expectedOutput)
}

// Test for valid YAML Secret input
func TestParseYamlSecret(t *testing.T) {
	input := `
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-secret
    data:
      key1: dmFsdWUx
      key2: dmFsdWUy
    stringData:
      key3: value3
    `

	expectedOutput := `
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-secret
    stringData:
      key1: value1
      key2: value2
      key3: value3
    `

	output, err := runKubectlUnsecret(input, "yaml")
	if err != nil {
		t.Fatalf("Error running kubectl-decode: %v", err)
	}
	compareYamlOutput(t, output, expectedOutput)
}

// Test for invalid JSON input
func TestInvalidJson(t *testing.T) {
	input := `{ "apiVersion": "v1", "kind": "Secret", "data": { "key1": "dmFsdWUx" ` // invalid JSON
	_, err := runKubectlUnsecret(input, "json")
	if err == nil {
		t.Errorf("Expected an error for invalid JSON but got none")
	}
}

// Test for invalid YAML input
func TestInvalidYaml(t *testing.T) {
	input := `
    apiVersion: v1
    kind: Secret
    data:
      key1: "dmFsdWUx
    ` // invalid YAML
	_, err := runKubectlUnsecret(input, "yaml")
	if err == nil {
		t.Errorf("Expected an error for invalid YAML but got none")
	}
}

// Removed pending decision on validity of empty input given we pass through non-YAML and non-JSON text.
//  // Test for empty input
//  func TestEmptyInput(t *testing.T) {
//      input := ""
//      _, err := runKubectlUnsecret(input, "json")
//      if err == nil {
//          t.Errorf("Expected an error for empty input but got none")
//      }
//  }

// Test round-trip for JSON input
func TestRoundTripJson(t *testing.T) {
	input := `{
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": "my-secret"},
        "stringData": {
            "key1": "value1",
            "key2": "value2"
        }
    }`

	expectedOutput := input

	output, err := runKubectlUnsecret(input, "json")
	if err != nil {
		t.Fatalf("Error running kubectl-decode: %v", err)
	}
	compareJsonOutput(t, output, expectedOutput)
}

// Test round-trip for YAML input
func TestRoundTripYaml(t *testing.T) {
	input := `
    apiVersion: v1
    kind: Secret
    metadata:
      name: my-secret
    stringData:
      key1: value1
      key2: value2
    `

	expectedOutput := input

	output, err := runKubectlUnsecret(input, "yaml")
	if err != nil {
		t.Fatalf("Error running kubectl-decode: %v", err)
	}
	compareYamlOutput(t, output, expectedOutput)
}

// Helper to compare JSON outputs
func compareJsonOutput(t *testing.T, output, expectedOutput string) {
	var outputMap, expectedMap map[string]interface{}
	json.Unmarshal([]byte(output), &outputMap)
	json.Unmarshal([]byte(expectedOutput), &expectedMap)

	if !jsonEqual(outputMap, expectedMap) {
		t.Errorf("Expected JSON output does not match. Got:\n%v\nExpected:\n%v", output, expectedOutput)
	}
}

// Helper to compare YAML outputs
func compareYamlOutput(t *testing.T, output, expectedOutput string) {
	var outputMap, expectedMap map[string]interface{}
	yaml.Unmarshal([]byte(output), &outputMap)
	yaml.Unmarshal([]byte(expectedOutput), &expectedMap)

	if !jsonEqual(outputMap, expectedMap) {
		t.Errorf("Expected YAML output does not match. Got:\n%v\nExpected:\n%v", output, expectedOutput)
	}
}

// jsonEqual compares two generic JSON structures
func jsonEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
