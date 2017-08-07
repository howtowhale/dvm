package extractor_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/archiver/extractor"
	"code.cloudfoundry.org/archiver/extractor/test_helper"
)

var _ = Describe("Extractor", func() {
	var extractor Extractor

	var extractionDest string
	var extractionSrc string

	BeforeEach(func() {
		var err error

		archive, err := ioutil.TempFile("", "extractor-archive")
		Expect(err).NotTo(HaveOccurred())

		extractionDest, err = ioutil.TempDir("", "extracted")
		Expect(err).NotTo(HaveOccurred())

		extractionSrc = archive.Name()

		extractor = NewDetectable()
	})

	AfterEach(func() {
		os.RemoveAll(extractionSrc)
		os.RemoveAll(extractionDest)
	})

	archiveFiles := []test_helper.ArchiveFile{
		{
			Name: "./",
			Dir:  true,
		},
		{
			Name: "./some-file",
			Body: "some-file-contents",
		},
		{
			Name: "./empty-dir/",
			Dir:  true,
		},
		{
			Name: "./nonempty-dir/",
			Dir:  true,
		},
		{
			Name: "./nonempty-dir/file-in-dir",
			Body: "file-in-dir-contents",
		},
		{
			Name: "./legit-exe-not-a-virus.bat",
			Mode: 0644,
			Body: "rm -rf /",
		},
		{
			Name: "./some-symlink",
			Link: "some-file",
			Mode: 0755,
		},
	}

	extractionTest := func() {
		err := extractor.Extract(extractionSrc, extractionDest)
		Expect(err).NotTo(HaveOccurred())

		fileContents, err := ioutil.ReadFile(filepath.Join(extractionDest, "some-file"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fileContents)).To(Equal("some-file-contents"))

		fileContents, err = ioutil.ReadFile(filepath.Join(extractionDest, "nonempty-dir", "file-in-dir"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fileContents)).To(Equal("file-in-dir-contents"))

		executable, err := os.Open(filepath.Join(extractionDest, "legit-exe-not-a-virus.bat"))
		Expect(err).NotTo(HaveOccurred())

		executableInfo, err := executable.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(executableInfo.Mode()).To(Equal(os.FileMode(0644)))

		emptyDir, err := os.Open(filepath.Join(extractionDest, "empty-dir"))
		Expect(err).NotTo(HaveOccurred())

		emptyDirInfo, err := emptyDir.Stat()
		Expect(err).NotTo(HaveOccurred())

		Expect(emptyDirInfo.IsDir()).To(BeTrue())

		target, err := os.Readlink(filepath.Join(extractionDest, "some-symlink"))
		Expect(err).NotTo(HaveOccurred())
		Expect(target).To(Equal("some-file"))

		symlinkInfo, err := os.Lstat(filepath.Join(extractionDest, "some-symlink"))
		Expect(err).NotTo(HaveOccurred())

		Expect(symlinkInfo.Mode() & 0755).To(Equal(os.FileMode(0755)))
	}

	Context("when the file is a zip archive", func() {
		BeforeEach(func() {
			test_helper.CreateZipArchive(extractionSrc, archiveFiles)
		})

		Context("when 'unzip' is on the PATH", func() {
			BeforeEach(func() {
				_, err := exec.LookPath("unzip")
				Expect(err).NotTo(HaveOccurred())
			})

			It("extracts the ZIP's files, generating directories, and honoring file permissions and symlinks", extractionTest)
		})

		Context("when 'unzip' is not in the PATH", func() {
			var oldPATH string

			BeforeEach(func() {
				oldPATH = os.Getenv("PATH")
				os.Setenv("PATH", "/dev/null")

				_, err := exec.LookPath("unzip")
				Expect(err).To(HaveOccurred())
			})

			AfterEach(func() {
				os.Setenv("PATH", oldPATH)
			})

			It("extracts the ZIP's files, generating directories, and honoring file permissions and symlinks", extractionTest)
		})
	})

	Context("when the file is a tgz archive", func() {
		BeforeEach(func() {
			test_helper.CreateTarGZArchive(extractionSrc, archiveFiles)
		})

		Context("when 'tar' is on the PATH", func() {
			BeforeEach(func() {
				_, err := exec.LookPath("tar")
				Expect(err).NotTo(HaveOccurred())
			})

			It("extracts the TGZ's files, generating directories, and honoring file permissions and symlinks", extractionTest)
		})

		Context("when 'tar' is not in the PATH", func() {
			var oldPATH string

			BeforeEach(func() {
				oldPATH = os.Getenv("PATH")
				os.Setenv("PATH", "/dev/null")

				_, err := exec.LookPath("tar")
				Expect(err).To(HaveOccurred())
			})

			AfterEach(func() {
				os.Setenv("PATH", oldPATH)
			})

			It("extracts the TGZ's files, generating directories, and honoring file permissions and symlinks", extractionTest)
		})
	})

	Context("when the file is a tar archive", func() {
		BeforeEach(func() {
			extractor = NewTar()
			test_helper.CreateTarArchive(extractionSrc, archiveFiles)
		})

		It("extracts the TAR's files, generating directories, and honoring file permissions and symlinks", extractionTest)
	})
})
