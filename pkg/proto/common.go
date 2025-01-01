// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package proto

import (
	"fmt"
	"strings"
	"time"
)

func TimeNow() *Timestamp {
	return Time(time.Now())
}

func Time(s time.Time) *Timestamp {
	return &Timestamp{
		Seconds: int32(s.Unix()),       //nolint:gosec
		Nanos:   int32(s.Nanosecond()), //nolint:gosec
	}
}

func (t *Timestamp) Time() time.Time {
	return time.Unix(int64(t.Seconds), int64(t.Nanos))
}

func (bi *BuildInfo) ToString() string {
	attrs := []string{
		"os=" + bi.Os,
		"arch=" + bi.Arch,
	}

	if len(bi.Commit) >= 8 {
		attrs = append(attrs, "commit="+bi.Commit[:8])
	}

	if bi.Branch != "" {
		attrs = append(attrs, "branch="+bi.Branch)
	}

	if bi.Date != nil {
		attrs = append(attrs, "built-at="+bi.Date.Time().Format(time.RFC3339))
	}

	if bi.BuiltBy != "" {
		attrs = append(attrs, "built-by="+bi.BuiltBy)
	}

	return fmt.Sprintf("%s (%s)", bi.Version, strings.Join(attrs, ", "))
}

func (bi *BuildInfos) ToString() string {
	lines := ""

	if bi.Client != nil {
		lines += fmt.Sprintf("client: %s\n", bi.Client.ToString())
	}

	if bi.Daemon != nil {
		lines += fmt.Sprintf("daemon: %s\n", bi.Daemon.ToString())
	}

	return lines
}
