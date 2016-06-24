package extractor

import (
	"net/http"
	"os"
)

func mimeType(src string) (string, error) {
	fd, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	data := make([]byte, 512)

	_, err = fd.Read(data)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(data), nil
}
