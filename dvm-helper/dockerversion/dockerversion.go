package dockerversion

import "fmt"
import "sort"
import "strings"
import "github.com/blang/semver"

const SystemAlias = "system"
const ExperimentalAlias = "experimental"

type Version struct {
	SemVer semver.Version
	Alias  string
}

func Parse(value string) Version {
	semver, err := semver.Parse(value)
	if err != nil {
		return Version{Alias: value}
	}

	return Version{SemVer: semver}
}

func (version Version) HasAlias() bool {
	return version.Alias != ""
}

func (version Version) HasSemVer() bool {
	return !(version.SemVer.Major == 0 && version.SemVer.Minor == 0 && version.SemVer.Patch == 0)
}

func (version Version) IsEmpty() bool {
	return !version.HasAlias() && !version.HasSemVer()
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
	cutoff := semver.MustParse("1.11.0")
	return version.IsExperimental() || version.SemVer.GTE(cutoff)
}

func (version Version) String() string {
	if version.HasAlias() {
		if version.HasSemVer() {
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
	return v.SemVer.Compare(o.SemVer)
}

// Equals checks if v is equal to o.
func (v Version) Equals(o Version) bool {
	semverMatch := v.Compare(o) == 0
	// Enables distinguishing between X.Y.Z and system (X.Y.Z)
	systemMatch := v.IsSystem() == o.IsSystem()
	aliasmatch := v.HasAlias() && strings.Compare(v.Alias, o.Alias) == 0
	return (semverMatch && systemMatch) || aliasmatch
}

// LT checks if v is less than o.
func (v Version) LT(o Version) bool {
	return v.Compare(o) == -1
}

type Versions []Version

func (s Versions) Len() int {
	return len(s)
}

func (s Versions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Versions) Less(i, j int) bool {
	return s[i].LT(s[j])
}

func Sort(versions []Version) {
	sort.Sort(Versions(versions))
}
