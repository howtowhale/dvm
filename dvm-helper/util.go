package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func exportEnvironmentVariable(name string) string {
	value := os.Getenv(name)

	if opts.Shell == "powershell" {
		return fmt.Sprintf("$env:%s=\"%s\"\r\n", name, value)
	}

	if opts.Shell == "cmd" {
		return fmt.Sprintf("%s=%s\r\n", name, value)
	}

	// default to bash
	return fmt.Sprintf("export %s=\"%s\"\n", name, value)
}

func ensureParentDirectoryExists(filePath string) {
	dir := filepath.Dir(filePath)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		die("Unable to create directory %s.", err, retCodeRuntimeError, dir)
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
	if !opts.Debug {
		return
	}

	color.Cyan(format, a...)
}

func writeInfo(format string, a ...interface{}) {
	if opts.Silent {
		return
	}

	color.White(format, a...)
}

func writeWarning(format string, a ...interface{}) {
	if opts.Silent {
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
