package main

import (
	"fmt"
	"github.com/ryanuber/go-glob"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type requestHandler func(w http.ResponseWriter, r *http.Request)

func docker1_10_3_Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case glob.Glob("*/version", r.RequestURI):
		fmt.Fprintln(w, `{
     "Version": "swarm/1.2.3",
     "ApiVersion": "1.22"
}`)
	default:
		w.WriteHeader(404)
	}
}

func docker1_12_1_Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case glob.Glob("*/version", r.RequestURI):
		fmt.Fprintln(w, `{
     "Version": "1.12.1"
}`)
	default:
		w.WriteHeader(404)
	}
}

func githubReleasesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.RequestURI {
	case "/repos/docker/docker/releases?per_page=100":
		fmt.Fprintln(w, loadTestData("github-docker-releases.json"))
	default:
		w.WriteHeader(404)
	}
}

func createMockDVM(h requestHandler) (docker *httptest.Server, github *httptest.Server) {
	docker = httptest.NewServer(http.HandlerFunc(h))
	github = httptest.NewServer(http.HandlerFunc(githubReleasesHandler))
	githubUrlOverride = github.URL + "/"

	return
}

func TestDetectOldVersion(t *testing.T) {
	docker, github := createMockDVM(docker1_10_3_Handler)
	defer docker.Close()
	defer github.Close()

	os.Setenv("DOCKER_HOST", docker.URL)
	os.Setenv("DOCKER_TLS_VERIFY", "0")
	os.Unsetenv("DOCKER_CERT_PATH")

	debug = true

	detect()
	version := os.Getenv("DOCKER_VERSION")
	assert.Equal(t, version, "1.10.3")
}

func TestDetectVersion(t *testing.T) {
	docker, github := createMockDVM(docker1_12_1_Handler)
	defer docker.Close()
	defer github.Close()

	os.Setenv("DOCKER_HOST", docker.URL)
	os.Setenv("DOCKER_TLS_VERIFY", "0")
	os.Unsetenv("DOCKER_CERT_PATH")

	debug = true

	detect()
	version := os.Getenv("DOCKER_VERSION")
	assert.Equal(t, version, "1.12.1")
}

func loadTestData(src string) string {
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
