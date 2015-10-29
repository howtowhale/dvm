package path

import (
	"fmt"
	"os"
	"regexp"
)

const pathEnvVar string = "PATH"

// Get the PATH environment variable value
func Get() string {
	return os.Getenv(pathEnvVar)
}

// Set the PATH environment variable value
func Set(value string) {
	os.Setenv(pathEnvVar, value)
}

// Prepend the specified value to the PATH environment variable
func Prepend(value string) {
	originalPath := Get()
	newPath := fmt.Sprintf("%s%s%s", value, separator, originalPath)
	Set(newPath)
}

// Remove any values which match the specified regular expression
// from the PATH environment variable
func Remove(regexValue string) {
	regex, _ := regexp.Compile(regexValue)
	newPath := regex.ReplaceAllString(Get(), "")
	Set(newPath)
}
