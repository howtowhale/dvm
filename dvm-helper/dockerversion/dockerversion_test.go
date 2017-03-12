package dockerversion_test

import (
	"testing"

	"github.com/howtowhale/dvm/dvm-helper/dockerversion"
	"github.com/stretchr/testify/assert"
)

func TestStripLeadingV(t *testing.T) {
	v := dockerversion.Parse("v1.0.0")
	assert.Equal(t, "1.0.0", v.String(), "Leading v should be stripped from the string representation")
	assert.Equal(t, "1.0.0", v.Name(), "Leading v should be stripped from the name")
	assert.Equal(t, "1.0.0", v.Value(), "Leading v should be stripped from the version value")
}

func TestIsPrerelease(t *testing.T) {
	var v dockerversion.Version

	v = dockerversion.Parse("17.3.0-ce-rc1")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = dockerversion.Parse("1.12.4-rc1")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = dockerversion.Parse("1.12.4-beta.1")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = dockerversion.Parse("1.12.4-alpha-2")
	assert.True(t, v.IsPrerelease(), "%s should be a prerelease", v)

	v = dockerversion.Parse("17.3.0-ce")
	assert.False(t, v.IsPrerelease(), "%s should NOT be a prerelease", v)
}

func TestPrereleaseUsesArchivedReleases(t *testing.T) {
	v := dockerversion.Parse("v1.12.5-rc1")

	assert.True(t, v.ShouldUseArchivedRelease())
}

func TestLeadingZeroInVersion(t *testing.T) {
	v := dockerversion.Parse("v17.03.0-ce")

	assert.Equal(t, "17.03.0-ce", v.String(), "Leading zeroes in the version should be preserved")
}

func TestSystemAlias(t *testing.T) {
	v := dockerversion.Parse(dockerversion.SystemAlias)
	assert.Empty(t, v.Slug(),
		"The system alias should not have a slug")
	assert.Equal(t, dockerversion.SystemAlias, v.String(),
		"An empty alias should only print the alias")
	assert.Equal(t, dockerversion.SystemAlias, v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "", v.Value(),
		"The value for an empty aliased version should be empty")
}

func TestExperimentalAlias(t *testing.T) {
	v := dockerversion.Parse(dockerversion.ExperimentalAlias)
	assert.Equal(t, dockerversion.ExperimentalAlias, v.Slug(),
		"The slug for the experimental version should be 'experimental'")
	assert.Equal(t, dockerversion.ExperimentalAlias, v.String(),
		"An empty alias should only print the alias")
	assert.Equal(t, dockerversion.ExperimentalAlias, v.Name(),
		"The name for an aliased version should be its alias")
	assert.Equal(t, "", v.Value(),
		"The value for an empty aliased version should be empty")
}

func TestAlias(t *testing.T) {
	v := dockerversion.NewAlias("prod", "1.2.3")
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
	v := dockerversion.Parse("1.2.3")
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
	v := dockerversion.Parse("1.2.3")
	v.SetAsExperimental()
	assert.True(t, v.IsExperimental())
}

func TestSetAsSystem(t *testing.T) {
	v := dockerversion.Parse("1.2.3")
	v.SetAsSystem()
	assert.True(t, v.IsSystem())
}