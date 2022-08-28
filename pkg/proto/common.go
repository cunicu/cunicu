package proto

import (
	"fmt"
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

func (t *Timestamp) Set(s time.Time) {
	t.Nanos = int32(s.Nanosecond())
	t.Seconds = int32(s.Unix())
}

func (t *Timestamp) Time() time.Time {
	return time.Unix(int64(t.Seconds), int64(t.Nanos))
}

func (bi *BuildInfo) ToString() string {
	commit := bi.Commit
	if len(commit) > 8 {
		commit = commit[:8]
	}

	date := "unknown"
	if bi.Date != nil {
		date = bi.Date.Time().Format(time.RFC3339)
	}

	return fmt.Sprintf("%s (%s, %s/%s, %s)", bi.Version, commit, bi.Os, bi.Arch, date)
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
