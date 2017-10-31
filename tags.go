package docker

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// DefaultTagSuffix returns a set of default suggested tags
// based on the commit ref with an attached suffix.
func DefaultTagSuffix(ref, suffix string) []string {
	tags := DefaultTags(ref)
	if len(suffix) == 0 {
		return tags
	}
	for i, tag := range tags {
		if tag == "latest" {
			tags[i] = suffix
		} else {
			tags[i] = fmt.Sprintf("%s-%s", tag, suffix)
		}
	}
	return tags
}

// DefaultTags returns a set of default suggested tags based on
// the commit ref.
func DefaultTags(ref string) []string {
	if !strings.HasPrefix(ref, "refs/tags/") {
		return []string{"latest"}
	}
	v := stripTagPrefix(ref)
	version, err := semver.NewVersion(v)
	if err != nil {
		return []string{"latest"}
	}
	if version.PreRelease != "" || version.Metadata != "" {
		return []string{
			version.String(),
		}
	}
	if version.Major == 0 {
		return []string{
			fmt.Sprintf("%d.%d", version.Major, version.Minor),
			fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch),
		}
	}
	return []string{
		fmt.Sprint(version.Major),
		fmt.Sprintf("%d.%d", version.Major, version.Minor),
		fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch),
	}
}

func stripTagPrefix(ref string) string {
	ref = strings.TrimPrefix(ref, "refs/tags/")
	ref = strings.TrimPrefix(ref, "v")
	return ref
}
