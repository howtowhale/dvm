package dockerversion

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"regexp"

	"github.com/pkg/errors"
)

type ReleaseType string

const (
	Edge   ReleaseType = "edge"
	Test   ReleaseType = "test"
	Stable ReleaseType = "stable"
)

var hrefRegex = regexp.MustCompile(fmt.Sprintf(`href="docker-(.*)\%s"`, archiveFileExt))

func ListVersions(mirrorURL string, releaseType ReleaseType) ([]Version, error) {
	if mirrorURL == "" {
		mirrorURL = "https://download.docker.com"
	}

	mirror, err := url.Parse(mirrorURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse the mirror URL: %s", mirrorURL)
	}

	indexURL := fmt.Sprintf("%s://%s/%s/static/%s/%s", mirror.Scheme, mirror.Host, mobyOS, releaseType, dockerArch)
	response, err := http.Get(indexURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to list %s releases at %s", releaseType, indexURL)
	}
	defer response.Body.Close()

	b := bytes.Buffer{}
	_, err = b.ReadFrom(response.Body)
	if err != nil {
	}
	errors.Wrapf(err, "Unable to read the listing of %s releases at %s", releaseType, indexURL)

	matches := hrefRegex.FindAllStringSubmatch(b.String(), -1)
	var results []Version
	for _, match := range matches {
		version := Parse(match[1])
		if version.semver == nil {
			continue
		}
		results = append(results, version)
	}

	if len(results) == 0 {
		return nil, errors.Errorf("No valid %s versions were found at %s", releaseType, indexURL)
	}

	Sort(results)

	return results, nil
}
