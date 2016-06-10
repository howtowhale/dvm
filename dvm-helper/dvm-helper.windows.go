// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)
import "strings"

const dockerOS string = "Windows"
const dvmOS string = "Windows"
const binaryFileExt string = ".exe"
const archiveFileExt string = ".zip"

func upgradeSelf(version string) {
	binaryURL := buildDvmReleaseURL(version, dvmOS, dvmArch, "dvm-helper.exe")
	binaryPath := filepath.Join(dvmDir, ".tmp", "dvm-helper.exe")
	downloadFileWithChecksum(binaryURL, binaryPath)

	psScriptURL := buildDvmReleaseURL(version, "dvm.ps1")
	psScriptPath := filepath.Join(dvmDir, "dvm.ps1")
	downloadFile(psScriptURL, psScriptPath)

	cmdScriptURL := buildDvmReleaseURL(version, "dvm.cmd")
	cmdScriptPath := filepath.Join(dvmDir, "dvm.cmd")
	downloadFile(cmdScriptURL, cmdScriptPath)

	writeUpgradeScript()
}

func writeUpgradeScript() {
	scriptPath := buildDvmOutputScriptPath()
	tmpBinaryPath := filepath.Join(dvmDir, ".tmp", "dvm-helper.exe")
	binaryPath := filepath.Join(dvmDir, "dvm-helper", "dvm-helper.exe")

	var contents string
	if shell == "powershell" {
		contents = fmt.Sprintf("cp -force '%s' '%s'", tmpBinaryPath, binaryPath)
	} else { // cmd
		contents = fmt.Sprintf("cp /Y '%s' '%s'", tmpBinaryPath, binaryPath)
	}

	writeFile(scriptPath, contents)
}

func getCleanPathRegex() string {
	versionDir := getVersionsDir()
	escapedVersionDir := strings.Replace(versionDir, `\`, `\\`, -1)
	return escapedVersionDir + `\\(\d+\.\d+\.\d+|experimental);`
}

func validateShellFlag() {
	if shell != "powershell" && shell != "cmd" {
		die("The --shell flag or SHELL environment variable must be set when running on Windows. Available values are powershell and cmd.", nil, retCodeInvalidArgument)
	}
}

func getUserHomeDir() string {
	return os.Getenv("USERPROFILE")
}
