// +build !windows

package main

const binaryFileExt string = ""

func getCleanDvmPathRegex() string {
	versionDir := getVersionDir("")
	return versionDir + `/(\d+\.\d+\.\d+|experimental):`
}

func validateShellFlag() {
	// we don't care about the shell flag on non-Windows platforms
}
