package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// LoadTestData reads the relative path under the testdata directory
// and returns the contents.
func LoadTestData(src string) string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	testFile := filepath.Join(pwd, "testdata", src)
	content, err := ioutil.ReadFile(testFile)
	if err != nil {
		panic(err)
	}
	return string(content)
}
