package dockerversion

import (
	"fmt"
	"net/http"
	"testing"

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

func TestExperimentalAlias(t *testing.T) {
	v := Parse(ExperimentalAlias)
	assert.Equal(t, ExperimentalAlias, v.Slug(),
		"The slug for the experimental version should be 'experimental'")
	assert.Equal(t, ExperimentalAlias, v.String(),
		"An empty alias should only print the alias")
	assert.Equal(t, ExperimentalAlias, v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "", v.Value(),
		"The value for an empty aliased version should be empty")
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

func TestSetAsExperimental(t *testing.T) {
	v := Parse("1.2.3")
	v.SetAsExperimental()
	assert.True(t, v.IsExperimental())
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
	}{
		// original download location, without compression
		Parse("1.10.3"): {fmt.Sprintf("https://get.docker.com/builds/%s/%s/docker-1.10.3", dockerOS, dockerArch), false},

		// original download location, without compression, prerelease
		Parse("1.10.0-rc1"): {fmt.Sprintf("https://test.docker.com/builds/%s/%s/docker-1.10.0-rc1", dockerOS, dockerArch), false},

		// compressed binaries
		Parse("1.11.0-rc1"): {fmt.Sprintf("https://test.docker.com/builds/%s/%s/docker-1.11.0-rc1.tgz", dockerOS, dockerArch), true},

		// original version scheme, prerelease binaries
		Parse("1.13.0-rc1"): {fmt.Sprintf("https://test.docker.com/builds/%s/%s/docker-1.13.0-rc1.tgz", dockerOS, dockerArch), true},

		// yearly notation, original download location, release location
		Parse("17.03.0-ce"): {fmt.Sprintf("https://get.docker.com/builds/%s/%s/docker-17.03.0-ce%s", dockerOS, dockerArch, archiveFileExt), true},

		// docker store download
		Parse("17.06.0-ce"): {fmt.Sprintf("https://download.docker.com/%s/static/stable/%s/docker-17.06.0-ce.tgz", mobyOS, dockerArch), true},

		// docker store download, prerelease
		Parse("17.07.0-ce-rc1"): {fmt.Sprintf("https://download.docker.com/%s/static/test/%s/docker-17.07.0-ce-rc1.tgz", mobyOS, dockerArch), true},

		// latest edge/experimental
		Parse("experimental"): {fmt.Sprintf("https://download.docker.com/%s/static/edge/%s/docker-17.06.0-ce.tgz", mobyOS, dockerArch), true},
	}

	for version, testcase := range testcases {
		t.Run(version.String(), func(t *testing.T) {
			gotURL, gotArchived := version.BuildDownloadURL("")
			if testcase.wantURL != gotURL {
				t.Fatalf("Expected %s to be downloaded from '%s', but got '%s'", version, testcase.wantURL, gotURL)
			}
			if testcase.wantArchived != gotArchived {
				t.Fatalf("Expected %s to use an archived download strategy", version)
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
