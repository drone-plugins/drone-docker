package docker

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// Default tags returns a set of default suggested tags based on
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
