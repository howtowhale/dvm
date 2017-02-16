package main

import "fmt"
import "io"
import "net/http"
import "os"
import "path"
import "path/filepath"
import "strings"
import "github.com/fatih/color"
import "github.com/pivotal-golang/archiver/extractor"
import "github.com/getcarina/dvm/dvm-helper/checksum"

func exportEnvironmentVariable(name string) string {
	value := os.Getenv(name)

	if shell == "powershell" {
		return fmt.Sprintf("$env:%s=\"%s\"\r\n", name, value)
	}

	if shell == "cmd" {
		return fmt.Sprintf("%s=%s\r\n", name, value)
	}

	// default to bash
	return fmt.Sprintf("export %s=\"%s\"\n", name, value)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureParentDirectoryExists(filePath string) {
	dir := filepath.Dir(filePath)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		die("Unable to create directory %s.", err, retCodeRuntimeError, dir)
	}
}

func downloadFile(url string, destPath string) {
	ensureParentDirectoryExists(destPath)

	destFile, err := os.Create(destPath)
	if err != nil {
		die("Unable to create to %s.", err, retCodeRuntimeError, destPath)
	}
	defer destFile.Close()
	os.Chmod(destPath, 0755)

	writeDebug("Downloading %s", url)

	response, err := http.Get(url)
	if err != nil {
		die("Unable to download %s.", err, retCodeRuntimeError, url)
	}

	if response.StatusCode != 200 {
		die("Unable to download %s. (Status %d)", nil, retCodeRuntimeError, url, response.StatusCode)
	}
	defer response.Body.Close()

	_, err = io.Copy(destFile, response.Body)
	if err != nil {
		die("Unable to write to %s.", err, retCodeRuntimeError, destPath)
	}
}

func downloadFileWithChecksum(url string, destPath string) {
	fileName := filepath.Base(destPath)
	tmpPath := filepath.Join(dvmDir, ".tmp", fileName)
	downloadFile(url, tmpPath)

	checksumURL := url + ".sha256"
	checksumPath := filepath.Join(dvmDir, ".tmp", (fileName + ".sh256"))
	downloadFile(checksumURL, checksumPath)

	checksum.CompareChecksum(tmpPath, checksumPath)
	isValid, err := checksum.CompareChecksum(tmpPath, checksumPath)
	if err != nil {
		die("Unable to calculate checksum of %s.", err, retCodeRuntimeError, tmpPath)
	}
	if !isValid {
		die("The checksum of %s failed to match %s.", nil, retCodeRuntimeError, tmpPath, checksumPath)
	}

	// Copy to final location, if different
	if destPath != tmpPath {
		ensureParentDirectoryExists(destPath)
		err = os.Rename(tmpPath, destPath)
		if err != nil {
			die("Unable to copy %s to %s.", err, retCodeRuntimeError, tmpPath, destPath)
		}
	}

	// Cleanup temp files
	if err = os.Remove(checksumPath); err != nil {
		writeWarning("Unable to remove temporary file: %s.", checksumPath)
	}
}

func downloadArchivedFileWithChecksum(url string, archivedFile string, destPath string) {
	archiveName := path.Base(url)
	tmpPath := filepath.Join(dvmDir, ".tmp", archiveName)
	downloadFileWithChecksum(url, tmpPath)

	// Extract the archive
	archivePath := filepath.Join(dvmDir, ".tmp", strings.TrimSuffix(archiveName, filepath.Ext(archiveName)))
	extractor := extractor.NewDetectable()
	extractor.Extract(tmpPath, archivePath)

	// Copy the archived file to the final destination
	archivedFilePath := filepath.Join(archivePath, archivedFile)
	ensureParentDirectoryExists(destPath)
	err := os.Rename(archivedFilePath, destPath)
	if err != nil {
		die("Unable to copy %s to %s.", err, retCodeRuntimeError, archivedFilePath, destPath)
	}

	// Cleanup temp files
	if err = os.Remove(tmpPath); err != nil {
		writeWarning("Unable to remove temporary file: %s\n%s", tmpPath, err)
	}
	if err = os.RemoveAll(archivePath); err != nil {
		writeWarning("Unable to remove temporary directory: %s\n%s", archivePath, err)
	}
}

func writeFile(path string, contents string) {
	writeDebug("Writing to %s...", path)
	writeDebug(contents)

	ensureParentDirectoryExists(path)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		die("Unable to create %s", err, retCodeRuntimeError, path)
	}

	_, err = io.WriteString(file, contents)
	if err != nil {
		die("Unable to write to %s", err, retCodeRuntimeError, path)
	}

	file.Close()
}

func writeDebug(format string, a ...interface{}) {
	if !debug {
		return
	}

	color.Cyan(format, a...)
}

func writeInfo(format string, a ...interface{}) {
	if silent {
		return
	}

	color.White(format, a...)
}

func writeWarning(format string, a ...interface{}) {
	if silent {
		return
	}

	color.Yellow(format, a...)
}

func writeError(format string, err error, a ...interface{}) {
	color.Set(color.FgRed)
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	color.Unset()
}

func die(format string, err error, exitCode int, a ...interface{}) {
	writeError(format, err, a...)
	os.Exit(exitCode)
}
