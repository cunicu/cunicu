package pb

import (
	"strings"
	"time"

	"github.com/pion/ice/v2"
	icex "riasc.eu/wice/pkg/ice"
)

func TimeNow() *Timestamp {
	return Time(time.Now())
}

func Time(s time.Time) *Timestamp {
	t := &Timestamp{}
	t.Set(s)
	return t
}

func (t *Timestamp) Set(s time.Time) {
	t.Nanos = int32(s.Nanosecond())
	t.Seconds = s.Unix()
}

func (t *Timestamp) Time() time.Time {
	return time.Unix(t.Seconds, int64(t.Nanos))
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

func (s *ConnectionState) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(s.String())), nil
}
