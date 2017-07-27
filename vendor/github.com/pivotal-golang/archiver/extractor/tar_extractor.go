package extractor

import (
	"archive/tar"
	"os"
)

type tarExtractor struct{}

func NewTar() Extractor {
	return &tarExtractor{}
}

func (e *tarExtractor) Extract(src, dest string) error {
	fd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fd.Close()

	tarReader := tar.NewReader(fd)
	return extractTarArchive(tarReader, dest)
}
