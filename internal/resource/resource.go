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
package resource

import (
	"fmt"
	"log"

	"github.com/philip-p-jones/kubectl-decode/internal/decode"
	"github.com/philip-p-jones/kubectl-decode/internal/format"
	"github.com/philip-p-jones/kubectl-decode/internal/logging"
)

// CommandExecutor is an interface for executing commands.
type CommandExecutor interface {
	CombinedOutput(name string, arg ...string) ([]byte, error)
}

// ProcessResource processes a single Kubernetes resource (base64 decoding)
func ProcessResource(resource map[interface{}]interface{}) error {
	// Log the initial input
	logging.DebugLog("Processing input resource: %+v\n", resource)

	if dataMap, ok := resource["data"].(map[interface{}]interface{}); ok {
		stringDataMap := make(map[string]string)
		for k, v := range dataMap {
			if strKey, ok := k.(string); ok {
				if strValue, ok := v.(string); ok {
					stringDataMap[strKey] = strValue
				}
			}
		}

		// Debug: Log the constructed stringDataMap
		logging.DebugLog("Constructed stringDataMap: %+v\n", stringDataMap)

		// Decode the base64 values
		decodedData, err := decode.DecodeDataMap(stringDataMap)
		if err != nil {
			log.Printf("Error decoding data: %v\n", err)
			return err
		}

		// Log the decoded data type and contents
		logging.DebugLog("Decoded data: %+v (type: %T)\n", decodedData, decodedData)

		// Remove the original data field and add stringData
		delete(resource, "data")
		resource["stringData"] = decodedData // Ensure decodedData is of the right type

		// Debug: Log the updated resource
		logging.DebugLog("Updated resource with stringData: %+v\n", resource)
	} else {
		logging.DebugLog("Warning: 'data' field not found or is not a map.")
	}

	return nil
}

// HandleGetCommand executes the "kubectl get" command and processes the output
func HandleGetCommand(args []string, executor CommandExecutor) error {
	if len(args) < 1 {
		return fmt.Errorf("resource type must be specified")
	}

	// Execute the kubectl command using the provided executor
	cmdOutput, err := executor.CombinedOutput("kubectl", append([]string{"get"}, args...)...)
	if err != nil {
		return fmt.Errorf("failed to execute kubectl command: %v\nOutput: %s", err, string(cmdOutput))
	}

	// Attempt to parse the command output
	inputData, err := format.ParseInput(cmdOutput)
	if err != nil {
		fmt.Println(string(cmdOutput))
		return fmt.Errorf("input appeared to be neither json nor yaml: %v", err)
	}

	// Convert inputData to map[interface{}]interface{}
	interfaceInputData := make(map[interface{}]interface{})
	for k, v := range inputData {
		interfaceInputData[k] = v
	}

	// Debug: Log the parsed inputData
	logging.DebugLog("Parsed inputData: %+v\n", interfaceInputData)

	// Process resources
	if kind, ok := interfaceInputData["kind"].(string); ok && kind == "List" {
		if items, ok := interfaceInputData["items"].([]interface{}); ok {
			for i, item := range items {
				if itemMap, ok := item.(map[interface{}]interface{}); ok {
					if err := ProcessResource(itemMap); err != nil {
						return fmt.Errorf("error processing resource in items[%d]: %v", i, err)
					}
				}
			}
		}
	} else {
		if err := ProcessResource(interfaceInputData); err != nil {
			return fmt.Errorf("error processing resource: %v", err)
		}
	}

	// Convert interfaceInputData back to map[string]interface{}
	stringInputData := make(map[string]interface{})
	for k, v := range interfaceInputData {
		if strKey, ok := k.(string); ok {
			stringInputData[strKey] = v
		}
	}

	return format.OutputResult(cmdOutput, stringInputData)
}
