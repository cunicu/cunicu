package pb

import (
	"fmt"
	"strings"
	"time"

	"github.com/pion/ice/v2"
	icex "riasc.eu/wice/pkg/ice"
	t "riasc.eu/wice/pkg/util/terminal"
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

func NewConnectionState(s icex.ConnectionState) ConnectionState {
	switch s {
	case ice.ConnectionStateNew:
		return ConnectionState_NEW
	case ice.ConnectionStateChecking:
		return ConnectionState_CHECKING
	case ice.ConnectionStateConnected:
		return ConnectionState_CONNECTED
	case ice.ConnectionStateCompleted:
		return ConnectionState_COMPLETED
	case ice.ConnectionStateFailed:
		return ConnectionState_FAILED
	case ice.ConnectionStateDisconnected:
		return ConnectionState_DISCONNECTED
	case ice.ConnectionStateClosed:
		return ConnectionState_CLOSED

	case icex.ConnectionStateCreating:
		return ConnectionState_CREATING
	case icex.ConnectionStateClosing:
		return ConnectionState_CLOSING
	case icex.ConnectionStateConnecting:
		return ConnectionState_CONNECTING
	case icex.ConnectionStateIdle:
		return ConnectionState_IDLE
	}

	return -1
}

func (s *ConnectionState) ConnectionState() icex.ConnectionState {
	switch *s {
	case ConnectionState_NEW:
		return ice.ConnectionStateNew
	case ConnectionState_CHECKING:
		return ice.ConnectionStateChecking
	case ConnectionState_CONNECTED:
		return ice.ConnectionStateConnected
	case ConnectionState_COMPLETED:
		return ice.ConnectionStateCompleted
	case ConnectionState_FAILED:
		return ice.ConnectionStateFailed
	case ConnectionState_DISCONNECTED:
		return ice.ConnectionStateDisconnected
	case ConnectionState_CLOSED:
		return ice.ConnectionStateClosed

	case ConnectionState_CREATING:
		return icex.ConnectionStateCreating
	case ConnectionState_CLOSING:
		return icex.ConnectionStateClosing
	case ConnectionState_CONNECTING:
		return icex.ConnectionStateConnecting
	case ConnectionState_IDLE:
		return icex.ConnectionStateIdle
	}

	return -1
}

func (s *ConnectionState) Color() string {
	switch *s {
	case ConnectionState_CHECKING:
		return t.FgYellow
	case ConnectionState_CONNECTED:
		return t.FgGreen
	case ConnectionState_FAILED:
		fallthrough
	case ConnectionState_DISCONNECTED:
		return t.FgRed
	case ConnectionState_NEW:
		fallthrough
	case ConnectionState_COMPLETED:
		fallthrough
	case ConnectionState_CLOSED:
		fallthrough
	case ConnectionState_CREATING:
		fallthrough
	case ConnectionState_CLOSING:
		fallthrough
	case ConnectionState_CONNECTING:
		fallthrough
	case ConnectionState_IDLE:
		return t.FgWhite
	}

	return t.FgDefault
}

func (s *ConnectionState) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(s.String())), nil
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
