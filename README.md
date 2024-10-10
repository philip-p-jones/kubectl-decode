# kubectl-decode
Decode Kubernetes Secret manifests to stringData for humans.

## Overview
`kubectl-decode` is a command-line tool for convenience when working with Kubernetes secrets. It decodes base64 `data` into human-readable `stringData`, and can be used as a simple plugin for `kubectl` to process data on standard input or to retrieve it from the Kubernetes API (as a simple wrapper around the `kubectl get` subcommand).

While similar in function to the `extract` subcommand in Red Hat OpenShift Client `oc`, `kubectl-decode` aims to provide Kubernetes API-compatible output, retaining the structure of the original resource for subsequent reuse.

## Usage

You can use `kubectl-decode` in two ways:

1. Piping the output of `kubectl get`:
   
   ```bash
   kubectl get secret my-secret -o yaml | kubectl-decode
   ```
   
2. As a kubectl plugin:
   
   ```bash
   kubectl decode get secret my-secret -o yaml
   ```

## Examples

### 1: Decoding a Secret from stdin

```bash
kubectl get secret example -o yaml | kubectl-decode
```

Output:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: default
stringData:
  key1: value1
  key2: value2
type: Opaque
```

### 2: Decoding a Secret retrieved from Kubernetes API

```bash
kubectl decode get secret example -o yaml
```

Output:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: example
  namespace: default
stringData:
  key1: value1
  key2: value2
type: Opaque
```

## Build

### Building locally

To build `kubectl-decode`, clone this repository and build it using Go:

```bash
git clone https://github.com/philip-p-jones/kubectl-decode.git
cd kubectl-decode
go build -o /usr/local/bin/kubectl-decode .
```

### Building and releasing using GoReleaser GitHub Actions

See .github/workflows/release.yml and .goreleaser.yaml

### Structure
The program is structured into several key components:

1. **Main Entry Point (`main.go`)**
   - The `main` function checks for command-line arguments. If the first argument is `get`, it delegates the command to `HandleGetCommand`. Otherwise, it reads from standard input (stdin).
   - For input from stdin, it attempts to parse the data as either JSON or YAML. If parsing fails, it prints the raw input, allowing the user to see what was provided.

2. **Input Handling (`format.go`)**
   - **AssertFormat**: States whether to treat the input is JSON or YAML based on the unmarshalling attempts in the ParseInput function or falling back to heuristic checks in its absence.
   - **ParseInput**: Attempts to unmarshal the input into a map. It first tries JSON, and if that fails, it tries YAML. If both fail, it returns a message indicating that no base64 data was found and suggests using `--output yaml|json`.
   - **OutputResult**: Outputs the parsed result in the appropriate format (YAML or JSON), depending on the detected input type.

3. **Resource Processing (`resource.go`)**
   - **ProcessResource**: This function takes a parsed resource map and checks for a `data` field containing base64-encoded strings. If found, it decodes them and replaces the original `data` field with a new `stringData` field.
   - **HandleGetCommand**: Executes the `kubectl get` command with the provided arguments and captures both stdout and stderr. It attempts to parse the command output and processes any resources found. If the output is neither JSON nor YAML, it prints the raw output for debugging.

4. **Decoding Functionality (`decode.go`)**
   - The program includes a helper function to decode base64-encoded data, converting it from a map of strings into a usable format.

5. **Versioning Information (`version.go`)**
   - Contains constants for the program's name and version, as well as a function to print this information.

### Workflow
1. **Command Invocation**: The user can invoke the program with a command such as:
   - `kubectl-decode get secret <name>`: This retrieves the specified secret from Kubernetes.
   - Alternatively, the user can pipe input into the program.

2. **Handling Input**:
   - If the user provides arguments, `HandleGetCommand` is called, which runs the `kubectl get` command and captures its output.
   - If there are no arguments, the program reads from stdin.

3. **Parsing and Decoding**:
   - The program attempts to parse the input. If it's valid JSON or YAML, it processes the resource, decoding any base64-encoded data found in the `data` field.
   - If parsing fails, the program outputs the raw input directly to assist the user.

4. **Output**:
   - The program formats the decoded output as either JSON or YAML based on the input type and writes it to stdout.

### Error Handling
- `kubectl-decode` provides basic error handling with messages for various failure scenarios:
  - If the wrapped `kubectl get` command fails, it returns the error along with the output.
  - If input parsing fails, it prints the raw input, allowing human-readable output from `kubectl get` to be viewed (e.g. with `--output plain`).

## Tests

To ensure that the `kubectl-decode` tool works as expected, we have included a set of unit tests that cover both JSON and YAML input/output parsing, error handling, and edge cases.

### Test Structure

- **`secret_test.go`**: Contains the Go test cases that validate the functionality of `kubectl-decode`. This includes tests for valid/invalid JSON and YAML inputs, edge cases like empty input, and round-trip tests to ensure idempotency.
- **`testdata/secret.yaml`**: Sample YAML input used in tests.
- **`testdata/secret.json`**: Sample JSON input used in tests.

### Running the Tests

To run the tests, ensure you have Go installed and navigate to the root directory of the project. Run the following command:

```bash
go test ./tests
```

This will execute all the tests in the `tests/` directory and output the results.

### Adding New Tests

If you need to add additional tests, place new test cases in `tests/secret_test.go` and include any new test data in the `testdata/` folder. Make sure to follow the existing test patterns and helper functions for consistency.
