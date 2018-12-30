package dockerversion

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/howtowhale/dvm/dvm-helper/internal/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestStripLeadingV(t *testing.T) {
	v := Parse("v1.0.0")
	assert.Equal(t, "1.0.0", v.String(), "Leading v should be stripped from the string representation")
	assert.Equal(t, "1.0.0", v.Name(), "Leading v should be stripped from the name")
	assert.Equal(t, "1.0.0", v.Value(), "Leading v should be stripped from the version value")
}

func TestIsPrerelease(t *testing.T) {
	var v Version

	v = Parse("17.3.0-ce-rc1")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = Parse("1.12.4-rc1")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = Parse("1.12.4-beta.1")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = Parse("1.12.4-alpha-2")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = Parse("17.3.0-ce")
	assert.False(t, v.IsPrerelease(), "%s should NOT be a prerelease", v)
}

func TestLeadingZeroInVersion(t *testing.T) {
	v := Parse("v17.03.0-ce")

	assert.Equal(t, "17.03.0-ce", v.String(), "Leading zeroes in the version should be preserved")
}

func TestSystemAlias(t *testing.T) {
	v := Parse(SystemAlias)
	assert.Empty(t, v.Slug(),
		"The system alias should not have a slug")
	assert.Equal(t, SystemAlias, v.String(),
		"An empty alias should only print the alias")
	assert.Equal(t, SystemAlias, v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "", v.Value(),
		"The value for an empty aliased version should be empty")
}

func TestEdgeAlias(t *testing.T) {
	v := Parse(EdgeAlias)
	assert.Equal(t, EdgeAlias, v.Slug(),
		"The slug for the edge version should be 'edge'")
	assert.Equal(t, EdgeAlias, v.String(),
		"An empty alias should only print the alias")
	assert.Equal(t, EdgeAlias, v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "", v.Value(),
		"The value for an empty aliased version should be empty")
}

func TestEdgeAliasWithVersion(t *testing.T) {
	v := Parse("17.06.0-ce+02c1d87")
	v.SetAsEdge()
	assert.Equal(t, EdgeAlias, v.Slug(),
		"The slug for the edge version should be 'edge'")
	assert.Equal(t, "edge (17.06.0-ce+02c1d87)", v.String(),
		"The string representation should include the alias and version")
	assert.Equal(t, EdgeAlias, v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "17.06.0-ce+02c1d87", v.Value(),
		"The value for a populated alias should be the version")
}

func TestAlias(t *testing.T) {
	v := NewAlias("prod", "1.2.3")
	assert.Equal(t, "1.2.3", v.Slug(),
		"The slug for an aliased version should be its semver value")
	assert.Equal(t, "prod (1.2.3)", v.String(),
		"The string representation for an aliased version should include both alias and version")
	assert.Equal(t, "prod", v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "1.2.3", v.Value(),
		"The value for an aliased version should be its semver value")
}

func TestSemanticVersion(t *testing.T) {
	v := Parse("1.2.3")
	assert.Equal(t, "1.2.3", v.Slug(),
		"The slug for a a semantic version should be its semver value")
	assert.Equal(t, "1.2.3", v.String(),
		"The string representation for a semantic version should only include the semver value")
	assert.Equal(t, "1.2.3", v.Name(),
		"The name for a semantic version should be its semver value")
	assert.Equal(t, "1.2.3", v.Value(),
		"The value for a semantic version should be its semver value")
}

func TestSetAsEdge(t *testing.T) {
	v := Parse("1.2.3")
	v.SetAsEdge()
	assert.True(t, v.IsEdge())
}

func TestSetAsSystem(t *testing.T) {
	v := Parse("1.2.3")
	v.SetAsSystem()
	assert.True(t, v.IsSystem())
}

func TestVersion_BuildDownloadURL(t *testing.T) {
	testcases := map[Version]struct {
		wantURL      string
		wantArchived bool
		wantChecksum bool
	}{
		// original download location, without compression
		Parse("1.10.3"): {
			wantURL:      fmt.Sprintf("https://get.docker.com/builds/%s/%s/docker-1.10.3", dockerOS, dockerArch),
			wantArchived: false,
			wantChecksum: true,
		},

		// original download location, without compression, prerelease
		/* test.docker.com has been removed by docker
		Parse("1.10.0-rc1"): {
			wantURL:      fmt.Sprintf("https://test.docker.com/builds/%s/%s/docker-1.10.0-rc1", dockerOS, dockerArch),
			wantArchived: false,
			wantChecksum: true,
		},
		*/

		// compressed binaries
		/* test.docker.com has been removed by docker
		Parse("1.11.0-rc1"): {
			wantURL:      fmt.Sprintf("https://test.docker.com/builds/%s/%s/docker-1.11.0-rc1.tgz", dockerOS, dockerArch),
			wantArchived: true,
			wantChecksum: true,
		},
		*/

		// original version scheme, prerelease binaries
		/* test.docker.com has been removed by docker
		Parse("1.13.0-rc1"): {
			wantURL:      fmt.Sprintf("https://test.docker.com/builds/%s/%s/docker-1.13.0-rc1.tgz", dockerOS, dockerArch),
			wantArchived: true,
			wantChecksum: true,
		},
		*/

		// yearly notation, original download location, release location
		Parse("17.03.0-ce"): {
			wantURL:      fmt.Sprintf("https://get.docker.com/builds/%s/%s/docker-17.03.0-ce%s", dockerOS, dockerArch, archiveFileExt),
			wantArchived: true,
			wantChecksum: true,
		},

		// docker store download (no more checksums)
		Parse("17.06.0-ce"): {
			wantURL:      fmt.Sprintf("https://download.docker.com/%s/static/stable/%s/docker-17.06.0-ce.tgz", mobyOS, dockerArch),
			wantArchived: true,
			wantChecksum: false,
		},

		// docker store download, prerelease
		Parse("17.07.0-ce-rc1"): {
			wantURL:      fmt.Sprintf("https://download.docker.com/%s/static/test/%s/docker-17.07.0-ce-rc1.tgz", mobyOS, dockerArch),
			wantArchived: true,
			wantChecksum: false,
		},
	}

	for version, testcase := range testcases {
		t.Run(version.String(), func(t *testing.T) {
			gotURL, gotArchived, gotChecksumed, err := version.buildDownloadURL("", false)
			if err != nil {
				t.Fatal(err)
			}

			if testcase.wantURL != gotURL {
				t.Fatalf("Expected %s to be downloaded from '%s', but got '%s'", version, testcase.wantURL, gotURL)
			}
			if testcase.wantArchived != gotArchived {
				t.Fatalf("Expected archive for %s to be %v, got %v", version, testcase.wantArchived, gotArchived)
			}
			if testcase.wantChecksum != gotChecksumed {
				t.Fatalf("Expected checksum for %s to be %v, got %v", version, testcase.wantChecksum, gotChecksumed)
			}

			response, err := http.DefaultClient.Head(gotURL)
			if err != nil {
				t.Fatalf("%#v", errors.Wrapf(err, "Unable to download release from %s", gotURL))
			}

			if response.StatusCode != 200 {
				t.Fatalf("Unexpected status code (%d) when downloading %s", response.StatusCode, gotURL)
			}
		})
	}
}

func TestVersion_DownloadEdgeRelease(t *testing.T) {
	version := Parse("edge")
	tempDir, _ := ioutil.TempDir("", "dvmtest")
	opts := config.NewDvmOptions()
	opts.DvmDir = filepath.Join(tempDir, ".dvm")
	destPath := filepath.Join(opts.DvmDir, "docker")

	err := version.Download(opts, destPath)
	if err != nil {
		t.Fatalf("%#v", err)
	}
}
