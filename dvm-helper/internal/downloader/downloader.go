package downloader

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/howtowhale/dvm/dvm-helper/checksum"
	"github.com/howtowhale/dvm/dvm-helper/internal/config"
	"github.com/pivotal-golang/archiver/extractor"
	"github.com/pkg/errors"
)

// Client is capable of downloading archived and checksumed files.
type Client struct {
	log *log.Logger
	tmp string
}

// New creates a downloader client.
// l - optional logger for debug output
func New(opts config.DvmOptions) Client {
	return Client{
		log: opts.Logger,
		tmp: filepath.Join(opts.DvmDir, ".tmp"),
	}
}

func (d Client) ensureParentDirectoryExists(path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	return errors.Wrapf(err, "Unable to create parent directory %s", path)
}

// DownloadFile saves a file without any additional processing.
func (d Client) DownloadFile(url string, destPath string) error {
	err := d.ensureParentDirectoryExists(destPath)
	if err != nil {
		return err
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return errors.Wrapf(err, "Unable to create %s", destPath)
	}
	defer destFile.Close()
	os.Chmod(destPath, 0755)

	d.log.Printf("Downloading %s to %s\n", url, destPath)

	response, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "Unable to download %s", url)
	}

	if response.StatusCode != 200 {
		return errors.Wrapf(err, "Unable to download %s (Status %d)", url, response.StatusCode)
	}
	defer response.Body.Close()

	_, err = io.Copy(destFile, response.Body)
	return errors.Wrapf(err, "Unable to write to %s", destPath)
}

// Download file saves a file after verifying the checksum found at url + ".sh256".
func (d Client) DownloadFileWithChecksum(url string, destPath string) error {
	fileName := filepath.Base(destPath)
	tmpPath := filepath.Join(d.tmp, fileName)
	err := d.DownloadFile(url, tmpPath)
	if err != nil {
		return err
	}

	checksumURL := url + ".sha256"
	checksumPath := filepath.Join(d.tmp, fileName+".sha256")
	err = d.DownloadFile(checksumURL, checksumPath)
	if err != nil {
		return err
	}

	isValid, err := checksum.CompareChecksum(tmpPath, checksumPath)
	if err != nil {
		return errors.Wrapf(err, "Unable to calculate checksum of %s", tmpPath)
	}
	if !isValid {
		return errors.Wrapf(err, "The checksum of %s failed to match %s", tmpPath, checksumPath)
	}

	// Copy to final location, if different
	if destPath != tmpPath {
		err = d.ensureParentDirectoryExists(destPath)
		if err != nil {
			return err
		}

		err = os.Rename(tmpPath, destPath)
		if err != nil {
			return errors.Wrapf(err, "Unable to copy %s to %s", tmpPath, destPath)
		}
	}

	// Cleanup temp files
	if err = os.Remove(checksumPath); err != nil {
		d.log.Println(errors.Wrapf(err, "Unable to remove temporary file %s", checksumPath))
	}

	return nil
}

// DownloadArchivedFile downloads the archive, decompresses it and saves the specified file to the destination path.
// url - URL of the archived file, e.g. a gzip, zip or tar file
// archivedFile - relative path to the desired file in the archive
// destPath - location where the archivedFile should be saved
func (d Client) DownloadArchivedFile(url string, archivedFile string, destPath string) error {
	archiveName := path.Base(url)
	tmpPath := filepath.Join(d.tmp, archiveName)

	err := d.DownloadFile(url, tmpPath)
	if err != nil {
		return err
	}

	return d.extractArchive(tmpPath, archiveName, archivedFile, destPath)
}

// DownloadArchivedFileWithChecksum first verifies the checksum found at url + ".sh256",
// decompresses the archive, and then saves the specified file to the destination path.
// url - URL of the archived file, e.g. a gzip, zip or tar file
// archivedFile - relative path to the desired file in the archive
// destPath - location where the archivedFile should be saved
func (d Client) DownloadArchivedFileWithChecksum(url string, archivedFile string, destPath string) error {
	archiveName := path.Base(url)
	tmpPath := filepath.Join(d.tmp, archiveName)

	err := d.DownloadFileWithChecksum(url, tmpPath)
	if err != nil {
		return err
	}

	return d.extractArchive(tmpPath, archiveName, archivedFile, destPath)
}

func (d Client) extractArchive(tmpPath string, archiveName string, archivedFile string, destPath string) error {
	// Extract the archive
	archivePath := filepath.Join(d.tmp, strings.TrimSuffix(archiveName, filepath.Ext(archiveName)))
	x := extractor.NewDetectable()
	x.Extract(tmpPath, archivePath)

	// Copy the archived file to the final destination
	archivedFilePath := filepath.Join(archivePath, archivedFile)

	err := d.ensureParentDirectoryExists(destPath)
	if err != nil {
		return err
	}

	err = os.Rename(archivedFilePath, destPath)
	if err != nil {
		return errors.Wrapf(err, "Unable to copy %s to %s", archivedFilePath, destPath)
	}

	// Cleanup temp files
	if err = os.Remove(tmpPath); err != nil {
		d.log.Println(errors.Wrapf(err, "Unable to remove temporary file %s", tmpPath))
	}
	if err = os.RemoveAll(archivePath); err != nil {
		d.log.Println(errors.Wrapf(err, "Unable to remove temporary directory %s", archivePath))
	}

	return nil
}
