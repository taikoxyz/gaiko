package version

import (
	"fmt"
	"runtime/debug"

	"github.com/ethereum/go-ethereum/version"
)

const ourPath = "github.com/taikoxyz/gaiko" // Path to our module

// These variables are set at build-time by the linker when the build is
// done by build/ci.go.
var gitCommit, gitDate string

// VCSInfo represents the git repository state.
type VCSInfo struct {
	Commit string // head commit hash
	Date   string // commit time in YYYYMMDD format
	Dirty  bool
}

// VCS returns version control information of the current executable.
func VCS() (VCSInfo, bool) {
	if gitCommit != "" {
		// Use information set by the build script if present.
		return VCSInfo{Commit: gitCommit, Date: gitDate}, true
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if buildInfo.Main.Path == ourPath {
			return buildInfoVCS(buildInfo)
		}
	}
	return VCSInfo{}, false
}

var Semantic = fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)

var WithMeta = func() string {
	v := Semantic
	if version.Meta != "" {
		v += "-" + version.Meta
	}
	return v
}()

func WithCommit(gitCommit, gitDate string) string {
	vsn := WithMeta
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	if (version.Meta != "stable") && (gitDate != "") {
		vsn += "-" + gitDate
	}
	return vsn
}
