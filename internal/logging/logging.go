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
package logging

import (
	"log"
	"os"
)

// debugMode checks if DEBUG environment variable is set.
var debugMode = os.Getenv("DEBUG") != ""

// debugLog is a helper function to print debug messages only if debug mode is enabled.
func DebugLog(format string, args ...interface{}) {
	if debugMode {
		log.Printf(format, args...)
	}
}
