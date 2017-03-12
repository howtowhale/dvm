package dockerversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripLeadingV(t *testing.T) {
	v := Parse("v1.0.0")
	assert.Equal(t, "1.0.0", v.String())
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

func TestPrereleasesUseArchivedReleases(t *testing.T) {
	v := Parse("v1.12.5-rc1")

	assert.True(t, v.ShouldUseArchivedRelease())
}

func TestLeadingZeroInVersion(t *testing.T) {
	v := Parse("v17.03.0-ce")

	assert.Equal(t, "17.03.0-ce", v.String(), "Leading zeroes in the version should be preserved")
}

func TestSystemAlias(t *testing.T) {
	v := Parse(SystemAlias)
	assert.Equal(t, SystemAlias, v.String())
}

func TestUserAlias(t *testing.T) {
	v := Parse("1.2.3")
	v.Alias = "prod"
	assert.Equal(t, "prod (1.2.3)", v.String())
}
