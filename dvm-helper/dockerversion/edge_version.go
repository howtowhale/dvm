package dockerversion

import (
	"fmt"
	"net/http"
	"net/url"

	"bytes"
	"regexp"

	"github.com/pkg/errors"
)

var hrefRegex = regexp.MustCompile("href=\"docker-(.*).tgz\"")

func findLatestEdgeVersion(mirrorURL string) (Version, error) {
	if mirrorURL == "" {
		mirrorURL = "https://download.docker.com"
	}

	mirror, err := url.Parse(mirrorURL)
	if err != nil {
		return Version{}, errors.Wrapf(err, "Unable to parse the mirror URL: %s", mirrorURL)
	}

	edgeReleasesUrl := fmt.Sprintf("%s://%s/%s/static/edge/%s", mirror.Scheme, mirror.Host, mobyOS, dockerArch)
	response, err := http.Get(edgeReleasesUrl)
	if err != nil {
		return Version{}, errors.Wrapf(err, "Unable to list edge releases at %s", edgeReleasesUrl)
	}
	defer response.Body.Close()

	b := bytes.Buffer{}
	_, err = b.ReadFrom(response.Body)
	if err != nil {
	}
	errors.Wrapf(err, "Unable to read the listing of edge releases at %s", edgeReleasesUrl)

	matches := hrefRegex.FindAllStringSubmatch(b.String(), -1)
	var results []Version
	for _, match := range matches {
		version := Parse(match[1])
		if version.semver == nil {
			continue
		}
		results = append(results, version)
	}

	Sort(results)

	if len(results) == 0 {
		return Version{}, errors.Errorf("No valid edge versions were found at %s", edgeReleasesUrl)
	}

	last := len(results) - 1
	return results[last], nil
}
