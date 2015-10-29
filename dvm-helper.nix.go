// +build !windows

package main

import (
	"fmt"
	"os"
)

const binaryFileExt string = ""

func prependDvmVersionToPath(version string) {
	versionDir := getVersionDir(version)
	path := fmt.Sprintf("%s:%s", versionDir, os.Getenv("PATH"))
	os.Setenv("PATH", path)
}

func getCleanDvmPathRegex() string {
	return getVersionDir("") + `/(\d+\.\d+\.\d+|experimental):`
}

func validateShellFlag() {
	// we don't care about the shell flag on non-Windows platforms
}
