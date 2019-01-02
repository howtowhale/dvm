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
	d := downloader.New(opts)

	binaryURL := buildDvmReleaseURL(version, dvmOS, dvmArch, "dvm-helper.exe")
	binaryPath := filepath.Join(opts.DvmDir, ".tmp", "dvm-helper.exe")
	err := d.DownloadFileWithChecksum(binaryURL, binaryPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

	psScriptURL := buildDvmReleaseURL(version, "dvm.ps1")
	psScriptPath := filepath.Join(opts.DvmDir, "dvm.ps1")
	err = d.DownloadFile(psScriptURL, psScriptPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

	cmdScriptURL := buildDvmReleaseURL(version, "dvm.cmd")
	cmdScriptPath := filepath.Join(opts.DvmDir, "dvm.cmd")
	err = d.DownloadFile(cmdScriptURL, cmdScriptPath)
	if err != nil {
		die("", err, retCodeRuntimeError)
	}

	writeUpgradeScript()
}

func writeUpgradeScript() {
	scriptPath := buildDvmOutputScriptPath()
	tmpBinaryPath := filepath.Join(opts.DvmDir, ".tmp", "dvm-helper.exe")
	binaryPath := filepath.Join(opts.DvmDir, "dvm-helper", "dvm-helper.exe")

	var contents string
	if opts.Shell == "powershell" {
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
	if opts.Shell != "powershell" && opts.Shell != "cmd" {
		die("The --shell flag or SHELL environment variable must be set when running on Windows. Available values are powershell and cmd.", nil, retCodeInvalidArgument)
	}
}

func getUserHomeDir() string {
	return os.Getenv("USERPROFILE")
}
