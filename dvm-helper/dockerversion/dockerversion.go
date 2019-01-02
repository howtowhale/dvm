package dockerversion

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/howtowhale/dvm/dvm-helper/internal/config"
	"github.com/howtowhale/dvm/dvm-helper/internal/downloader"
	"github.com/pkg/errors"
)

const SystemAlias = "system"
const EdgeAlias = "edge"

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
		v.semver = &semver
	} else {
		v.alias = value
	}
	return v
}

func (version Version) buildDownloadURL(mirror string, forcePrerelease bool) (url string, archived bool, checksumed bool, err error) {
	var releaseSlug, versionSlug, extSlug string

	var edgeVersion Version
	if version.IsEdge() {
		edgeVersion, err = findLatestEdgeVersion(mirror)
		if err != nil {
			return
		}
	}

	// Docker Store Download
	if version.shouldBeInDockerStore() {
		archived = true
		checksumed = false
		extSlug = archiveFileExt
		if mirror == "" {
			mirror = "download.docker.com"
		}
		if version.IsEdge() {
			releaseSlug = "edge"
			versionSlug = edgeVersion.String()
		} else if version.IsPrerelease() || forcePrerelease {
			releaseSlug = "test"
			versionSlug = version.String()
		} else {
			releaseSlug = "stable"
			versionSlug = version.String()
		}

		url = fmt.Sprintf("https://%s/%s/static/%s/%s/docker-%s%s",
			mirror, mobyOS, releaseSlug, dockerArch, versionSlug, extSlug)
		return
	} else { // Original Download
		archived = version.shouldBeArchived()
		checksumed = true
		versionSlug = version.String()
		if archived {
			extSlug = archiveFileExt
		} else {
			extSlug = binaryFileExt
		}
		if mirror == "" {
			mirror = "docker.com"
		}
		if version.IsPrerelease() {
			releaseSlug = "test"
		} else {
			releaseSlug = "get"
		}

		url = fmt.Sprintf("https://%s.%s/builds/%s/%s/docker-%s%s",
			releaseSlug, mirror, dockerOS, dockerArch, versionSlug, extSlug)
		return
	}
}

// Download a Docker release.
// version - the desired version.
// mirrorURL - optional alternate download location.
// binaryPath - full path to where the Docker client binary should be saved.
func (version Version) Download(opts config.DvmOptions, binaryPath string) error {
	err := version.download(false, opts, binaryPath)
	if err != nil && !version.IsPrerelease() && version.shouldBeInDockerStore() {
		// Docker initially publishes non-rc version versions to the test location
		// and then later republishes to the stable location
		// Retry stable versions against test to find "unstable" stable versions. :-)
		opts.Logger.Printf("Could not find a stable release for %s, checking for a test release\n", version)
		retryErr := version.download(true, opts, binaryPath)
		return errors.Wrapf(retryErr, "Attempted to fallback to downloading from the prerelease location after downloading from the stable location failed: %s", err.Error())
	}
	return err
}

func (version Version) download(forcePrerelease bool, opts config.DvmOptions, binaryPath string) error {
	url, archived, checksumed, err := version.buildDownloadURL(opts.MirrorURL, forcePrerelease)
	if err != nil {
		return errors.Wrapf(err, "Unable to determine the download URL for %s", version)
	}

	opts.Logger.Printf("Checking if %s can be found at %s", version, url)
	head, err := http.Head(url)
	if err != nil {
		return errors.Wrapf(err, "Unable to determine if %s is a valid version", version)
	}
	if head.StatusCode >= 400 {
		return errors.Errorf("Version %s not found (%v) - try `dvm ls-remote` to browse available versions", version, head.StatusCode)
	}

	d := downloader.New(opts)
	binaryName := filepath.Base(binaryPath)

	if archived {
		archivedFile := filepath.Join("docker", binaryName)
		if checksumed {
			return d.DownloadArchivedFileWithChecksum(url, archivedFile, binaryPath)
		}
		return d.DownloadArchivedFile(url, archivedFile, binaryPath)
	}

	if checksumed {
		return d.DownloadFileWithChecksum(url, binaryPath)
	}
	return d.DownloadFile(url, binaryPath)
}

func (version Version) shouldBeInDockerStore() bool {
	if version.IsEdge() {
		return true
	}

	dockerStoreCutoff, _ := semver.NewVersion("17.06.0-ce")

	return version.semver != nil && !version.semver.LessThan(dockerStoreCutoff)
}

func (version Version) shouldBeArchived() bool {
	archivedReleaseCutoff, _ := semver.NewVersion("1.11.0-rc1")
	return version.semver != nil && !version.semver.LessThan(archivedReleaseCutoff)
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

func (version Version) IsEdge() bool {
	return version.alias == EdgeAlias
}

func (version *Version) SetAsEdge() {
	version.alias = EdgeAlias
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

// Slug is the path segment under DVM_DIR where the binary is located
func (version Version) Slug() string {
	if version.IsSystem() {
		return ""
	}
	if version.IsEdge() {
		return version.alias
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
	if strings.HasPrefix(strings.ToLower(value), "v") {
		value = value[1:]
	}
	return value
}

func (version Version) InRange(r string) (bool, error) {
	c, err := semver.NewConstraint(r)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to parse range constraint: %s", r)
	}
	if version.semver == nil {
		return false, nil
	}
	return c.Matches(*version.semver) == nil, nil
}

// Compare compares Versions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v Version) Compare(o Version) int {
	if v.semver != nil && o.semver != nil {
		return v.semver.Compare(*o.semver)
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
