// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
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
		Seconds: int32(s.Unix()),
		Nanos:   int32(s.Nanosecond()),
	}
}

func (t *Timestamp) Time() time.Time {
	return time.Unix(int64(t.Seconds), int64(t.Nanos))
}

func (bi *BuildInfo) ToString() string {
	attrs := []string{
		fmt.Sprintf("os=%s", bi.Os),
		fmt.Sprintf("arch=%s", bi.Arch),
	}

	if bi.Commit != "" {
		attrs = append(attrs, fmt.Sprintf("commit=%s", bi.Commit[:8]))
	}

	if bi.Branch != "" {
		attrs = append(attrs, fmt.Sprintf("branch=%s", bi.Branch))
	}

	if bi.Date != nil {
		attrs = append(attrs, fmt.Sprintf("built-at=%s", bi.Date.Time().Format(time.RFC3339)))
	}

	if bi.BuiltBy != "" {
		attrs = append(attrs, fmt.Sprintf("built-by=%s", bi.BuiltBy))
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
