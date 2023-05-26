// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package buildinfo provides access to build-time information such as the build date and version control details
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/stv0g/cunicu/pkg/proto"
)

//nolint:gochecknoglobals
var (
	// set via ldflags -X / goreleaser or from debug.ReadBuildInfo()
	Version = "dev"
	Commit  = "none"
	Tag     = ""
	Branch  = ""
	BuiltBy = "manual"
	DateStr = ""
	Date    *time.Time
	Dirty   bool
)

func init() { //nolint:gochecknoinits
	if Version == "dev" {
		_, Commit, Dirty, Date = ReadVCSInfos()
	} else {
		Dirty = strings.Contains(Version, "-dirty")
		if bd, err := time.Parse(time.RFC3339, DateStr); err == nil && !bd.IsZero() {
			Date = &bd
		}
	}
}

func BuildInfo() *proto.BuildInfo {
	bi := &proto.BuildInfo{
		Version: Version,
		Commit:  Commit,
		Tag:     Tag,
		Branch:  Branch,
		BuiltBy: BuiltBy,
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Dirty:   Dirty,
	}

	if Date != nil {
		bi.Date = proto.Time(*Date)
	}

	return bi
}

func UserAgent() string {
	return fmt.Sprintf("cunicu/%s (%s/%s rev %s)", Version, runtime.GOOS, runtime.GOARCH, Commit)
}

func ReadVCSInfos() (bool, string, bool, *time.Time) {
	if info, ok := debug.ReadBuildInfo(); ok {
		commit := ""
		dirty := false
		var btime *time.Time

		for _, v := range info.Settings {
			switch v.Key {
			case "vcs.revision":
				commit = v.Value
			case "vcs.modified":
				dirty = v.Value == "true"
			case "vcs.time":
				if bd, err := time.Parse(time.RFC3339, v.Value); err == nil && !bd.IsZero() {
					btime = &bd
				}
			}
		}

		if dirty {
			commit += "-dirty"
		}

		return true, commit, dirty, btime
	}

	return false, "", false, nil
}
