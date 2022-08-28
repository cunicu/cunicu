// Package buildinfo provides access to build-time information such as the build date and version control details
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"riasc.eu/wice/pkg/pb"
)

var (
	// set via ldflags -X / goreleaser or from debug.ReadBuildInfo()
	Version      = "dev"
	Commit       = "none"
	Tag          = "unknown"
	Branch       = "unknown"
	BuiltBy      = "unknown"
	BuiltDateStr = "unknown"
	BuiltDate    *time.Time
	Dirty        bool
)

func init() {
	if Version == "dev" {
		_, Commit, Dirty, BuiltDate = ReadVCSInfos()
	} else {
		Dirty = strings.Contains(Version, "-dirty")
		if bd, err := time.Parse(time.RFC3339, BuiltDateStr); err == nil && !bd.IsZero() {
			BuiltDate = &bd
		}
	}
}

func BuildInfo() *pb.BuildInfo {
	bi := &pb.BuildInfo{
		Version: Version,
		Commit:  Commit,
		Tag:     Tag,
		Branch:  Branch,
		BuiltBy: BuiltBy,
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Dirty:   Dirty,
	}

	if BuiltDate != nil {
		bi.Date = pb.Time(*BuiltDate)
	}

	return bi
}

func UserAgent() string {
	return fmt.Sprintf("wice/%s (%s/%s rev %s)", Version, runtime.GOOS, runtime.GOARCH, Commit)
}

func ReadVCSInfos() (bool, string, bool, *time.Time) {
	if info, ok := debug.ReadBuildInfo(); ok {
		rev := "unknown"
		dirty := false
		var btime *time.Time

		for _, v := range info.Settings {
			switch v.Key {
			case "vcs.revision":
				rev = v.Value
			case "vcs.modified":
				dirty = v.Value == "true"
			case "vcs.time":
				if bd, err := time.Parse(time.RFC3339, v.Value); err == nil && !bd.IsZero() {
					btime = &bd
				}
			}
		}

		if dirty {
			rev += "-dirty"
		}

		return true, rev, dirty, btime
	}

	return false, "", false, nil
}
