package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// CompareChecksum validates the SHA256 checksum for a binary against its checksum file
func CompareChecksum(filePath string, checksumPath string) (bool, error) {
	knownChecksum, err := readChecksum(checksumPath)
	if err != nil {
		return false, err
	}

	checksum, err := calculateChecksum(filePath)
	if err != nil {
		return false, err
	}

	return strings.Compare(knownChecksum, checksum) == 0, nil
}

func readChecksum(checksumPath string) (string, error) {
	contents, err := ioutil.ReadFile(checksumPath)
	if err != nil {
		return "", err
	}
	checksum := strings.Split(string(contents), " ")[0]
	return checksum, nil
}

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	checksum := hash.Sum(nil)

	// convert from bytes to ascii hex string
	return fmt.Sprintf("%x", checksum), nil
}
