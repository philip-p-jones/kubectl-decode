package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/philip-p-jones/kubectl-decode/internal/format"
	"github.com/philip-p-jones/kubectl-decode/internal/resource"
)

var debugMode = os.Getenv("DEBUG") != ""

// CommandExecutor interface for executing commands
type CommandExecutor interface {
	CombinedOutput(name string, arg ...string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using exec package
type RealCommandExecutor struct{}

// CombinedOutput executes a command and returns its output
func (r *RealCommandExecutor) CombinedOutput(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput()
}

func Execute() {
	// Check for subcommands
	if len(os.Args) > 1 && os.Args[1] == "get" {
		executor := &RealCommandExecutor{}
		if err := resource.HandleGetCommand(os.Args[2:], executor); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Read input from stdin
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		inputData, err := format.ParseInput(input)
		if err != nil {
			fmt.Print(string(input)) // Print the input data directly
			os.Exit(1)
		}

		// Convert inputData to map[interface{}]interface{}
		interfaceInputData := make(map[interface{}]interface{})
		for k, v := range inputData {
			interfaceInputData[k] = v
		}

		// Process the input resource
		if err := resource.ProcessResource(interfaceInputData); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing resource: %v\n", err)
			os.Exit(1)
		}

		// Convert interfaceInputData back to map[string]interface{}
		stringInputData := make(map[string]interface{})
		for k, v := range interfaceInputData {
			if strKey, ok := k.(string); ok {
				stringInputData[strKey] = v
			}
		}

		// Output result
		if err := format.OutputResult(input, stringInputData); err != nil {
			fmt.Fprintf(os.Stderr, "Error outputting result: %v\n", err)
			os.Exit(1)
		}
	}
}
