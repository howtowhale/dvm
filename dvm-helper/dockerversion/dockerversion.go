package dockerversion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

const SystemAlias = "system"
const ExperimentalAlias = "experimental"

type Version struct {
	// Since Docker versions aren't valid versions, (it has leading zeroes)
	// We must save the original string representation
	raw    string
	semver *semver.Version
	alias  string
}

func NewAlias(alias string, value string) Version {
	v := Parse(value)
	v.alias = alias
	return v
}

func Parse(value string) Version {
	v := Version{raw: value}
	semver, err := semver.NewVersion(value)
	if err == nil {
		v.semver = semver
	} else {
		v.alias = value
	}
	return v
}

func (version Version) IsPrerelease() bool {
	if version.semver == nil {
		return false
	}

	tag := version.semver.Prerelease()

	preTags := []string{"rc", "alpha", "beta"}
	for i := 0; i < len(preTags); i++ {
		if strings.Contains(tag, preTags[i]) {
			return true
		}
	}

	return false
}

func (version Version) IsEmpty() bool {
	return version.semver == nil
}

func (version Version) IsAlias() bool {
	return version.alias != ""
}

func (version Version) IsSystem() bool {
	return version.alias == SystemAlias
}

func (version *Version) SetAsSystem() {
	version.alias = SystemAlias
}

func (version Version) IsExperimental() bool {
	return version.alias == ExperimentalAlias
}

func (version *Version) SetAsExperimental() {
	version.alias = ExperimentalAlias
}

func (version Version) ShouldUseArchivedRelease() bool {
	cutoff, _ := semver.NewConstraint(">= 1.11.0-rc1")
	return version.IsExperimental() || cutoff.Check(version.semver)
}

func (version Version) String() string {
	if version.alias != "" && version.semver != nil {
		return fmt.Sprintf("%s (%s)", version.alias, version.formatRaw())
	}
	return version.formatRaw()
}

func (version Version) Value() string {
	if version.semver == nil {
		return ""
	}
	return version.formatRaw()
}

func (version Version) Name() string {
	if version.alias != "" {
		return version.alias
	}
	return version.formatRaw()
}

func (version Version) formatRaw() string {
	value := version.raw
	prefix := value[0:1]
	if strings.ToLower(prefix) == "v" {
		value = value[1:]
	}
	return value
}

func (version Version) InRange(r string) (bool, error) {
	c, err := semver.NewConstraint(r)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to parse range constraint: %s", r)
	}
	return c.Check(version.semver), nil
}

// Compare compares Versions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v Version) Compare(o Version) int {
	if v.semver != nil && o.semver != nil {
		return v.semver.Compare(o.semver)
	}

	return strings.Compare(v.alias, o.alias)
}

// Equals checks if v is equal to o.
func (v Version) Equals(o Version) bool {
	semverMatch := v.Compare(o) == 0
	// Enables distinguishing between X.Y.Z and system (X.Y.Z)
	systemMatch := v.IsSystem() == o.IsSystem()
	aliasmatch := v.alias != "" && v.alias == o.alias
	return (semverMatch && systemMatch) || aliasmatch
}

type Versions []Version

func (s Versions) Len() int {
	return len(s)
}

func (s Versions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Versions) Less(i, j int) bool {
	return s[i].Compare(s[j]) == -1
}

func Sort(versions []Version) {
	sort.Sort(Versions(versions))
}
