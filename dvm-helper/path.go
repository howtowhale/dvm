package main

import (
	"fmt"
	"os"
	"regexp"
)

const pathEnvVar string = "PATH"

// Get the PATH environment variable value
func getPath() string {
	return os.Getenv(pathEnvVar)
}

// Set the PATH environment variable value
func setPath(value string) {
	os.Setenv(pathEnvVar, value)
}

// Prepend the specified value to the PATH environment variable
func prependPath(value string) {
	originalPath := getPath()
	newPath := fmt.Sprintf("%s%c%s", value, os.PathListSeparator, originalPath)
	setPath(newPath)
}

// Remove any values which match the specified regular expression
// from the PATH environment variable
func removePath(regexValue string) {
	regex, _ := regexp.Compile(regexValue)
	newPath := regex.ReplaceAllString(getPath(), "")
	setPath(newPath)
}

// Build a script which sets the PATH environment variable
func buildPathScript() string {
	path := getPath()

	if shell == "powershell" {
		return fmt.Sprintf(`$env:PATH="%s"`, path)
	}

	if shell == "cmd" {
		return fmt.Sprintf("PATH=%s", path)
	}

	// default to bash
	return fmt.Sprintf(`export PATH="%s"`, path)
}
