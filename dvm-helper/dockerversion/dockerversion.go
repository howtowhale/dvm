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
	SemVer *semver.Version
	Alias  string
}

func New(semver *semver.Version) Version {
	return Version{SemVer: semver}
}

func Parse(value string) Version {
	semver, err := semver.NewVersion(value)
	if err != nil {
		return Version{Alias: value}
	}

	return New(semver)
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
	cutoff, _ := semver.NewConstraint(">= 1.11.0")
	return version.IsExperimental() || cutoff.Check(version.SemVer)
}

func (version Version) String() string {
	if version.Alias != "" {
		if version.SemVer != nil {
			return fmt.Sprintf("%s (%s)", version.Alias, version.SemVer.String())
		}
		return version.Alias
	}
	return version.SemVer.String()
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
