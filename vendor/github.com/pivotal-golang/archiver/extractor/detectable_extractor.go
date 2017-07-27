package extractor

import (
	"fmt"
)

type detectableExtractor struct{}

func NewDetectable() Extractor {
	return &detectableExtractor{}
}

func (e *detectableExtractor) Extract(src, dest string) error {
	srcType, err := mimeType(src)
	if err != nil {
		return err
	}

	switch srcType {
	case "application/zip":
		err := NewZip().Extract(src, dest)
		if err != nil {
			return err
		}
	case "application/x-gzip":
		err := NewTgz().Extract(src, dest)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s is an unsupported archive type: %s", src, srcType)
	}

	return nil
}
