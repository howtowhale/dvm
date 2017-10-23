// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/howtowhale/dvm/dvm-helper/internal/downloader"
)

const dvmOS string = "Windows"
const binaryFileExt string = ".exe"

func upgradeSelf(version string) {
	d := downloader.New(getDebugLogger())

	binaryURL := buildDvmReleaseURL(version, dvmOS, dvmArch, "dvm-helper.exe")
	binaryPath := filepath.Join(dvmDir, ".tmp", "dvm-helper.exe")
	err := d.DownloadFileWithChecksum(binaryURL, binaryPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

	psScriptURL := buildDvmReleaseURL(version, "dvm.ps1")
	psScriptPath := filepath.Join(dvmDir, "dvm.ps1")
	err = d.DownloadFile(psScriptURL, psScriptPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

	cmdScriptURL := buildDvmReleaseURL(version, "dvm.cmd")
	cmdScriptPath := filepath.Join(dvmDir, "dvm.cmd")
	err = d.DownloadFile(cmdScriptURL, cmdScriptPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

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
	return escapedVersionDir + `\\[^:]+;`
}

func validateShellFlag() {
	if shell != "powershell" && shell != "cmd" {
		die("The --shell flag or SHELL environment variable must be set when running on Windows. Available values are powershell and cmd.", nil, retCodeInvalidArgument)
	}
}

func getUserHomeDir() string {
	return os.Getenv("USERPROFILE")
}
