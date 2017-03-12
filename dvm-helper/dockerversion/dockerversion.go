package dockerversion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
)

const SystemAlias = "system"
const ExperimentalAlias = "experimental"

type Version struct {
	// Since Docker versions aren't valid versions, (it has leading zeroes)
	// We must save the original string representation
	Raw string
	SemVer *semver.Version
	Alias  string
}

func Parse(value string) Version {
	v := Version{Raw: value}
	semver, err := semver.NewVersion(value)
	if err == nil {
		v.SemVer = semver
	} else {
		v.Alias = value
	}
	return v
}

func (version Version) IsPrerelease() bool {
	if version.SemVer == nil {
		return false
	}

	tag := version.SemVer.Prerelease()

	preTags := []string{"rc", "alpha", "beta"}
	for i := 0; i < len(preTags); i++ {
		if strings.Contains(tag, preTags[i]) {
			return true
		}
	}

	return false
}

func (version Version) IsEmpty() bool {
	return version.Alias == "" && version.SemVer == nil
}

func (version Version) IsSystem() bool {
	return version.Alias == SystemAlias
}

func (version *Version) SetAsSystem() {
	version.Alias = SystemAlias
}

func (version Version) IsExperimental() bool {
	return version.Alias == ExperimentalAlias
}

func (version *Version) SetAsExperimental() {
	version.Alias = ExperimentalAlias
}

func (version Version) ShouldUseArchivedRelease() bool {
	cutoff, _ := semver.NewConstraint(">= 1.11.0-rc1")
	return version.IsExperimental() || cutoff.Check(version.SemVer)
}

func (version Version) String() string {
	raw := version.formatRaw()
	if version.Alias != "" && version.SemVer != nil {
		return fmt.Sprintf("%s (%s)", version.Alias, raw)
	}
	return raw
}
func (version Version) formatRaw() string {
	value := version.Raw
	prefix := value[0:1]
	if strings.ToLower(prefix) == "v" {
		value = value[1:]
	}
	return value
}

// Compare compares Versions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v Version) Compare(o Version) int {
	if v.SemVer != nil && o.SemVer != nil {
		return v.SemVer.Compare(o.SemVer)
	}

	return strings.Compare(v.Alias, o.Alias)
}

// Equals checks if v is equal to o.
func (v Version) Equals(o Version) bool {
	semverMatch := v.Compare(o) == 0
	// Enables distinguishing between X.Y.Z and system (X.Y.Z)
	systemMatch := v.IsSystem() == o.IsSystem()
	aliasmatch := v.Alias != "" && v.Alias == o.Alias
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
