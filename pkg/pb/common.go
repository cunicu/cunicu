package pb

import (
	"strings"
	"time"

	"github.com/pion/ice/v2"
	icex "riasc.eu/wice/pkg/ice"
)

var (
	Success = &Error{
		Code: Error_SUCCESS,
	}

	ErrNotSupported = &Error{
		Code:    Error_ENOTSUP,
		Message: "not supported yet",
	}

	ErrNotAuthorized = &Error{
		Code:    Error_EPERM,
		Message: "not authorized",
	}
)

func NewError(e error) *Error {
	return &Error{
		Code:    Error_EUNKNOWN,
		Message: e.Error(),
	}
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Ok() bool {
	return e.Code == Error_SUCCESS
}

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

	case icex.ConnectionStateClosing:
		return ConnectionState_CLOSING
	case icex.ConnectionStateConnecting:
		return ConnectionState_CONNECTING
	case icex.ConnectionStateIdle:
		return ConnectionState_IDLE
	case icex.ConnectionStateUnknown:
		return ConnectionState_UNKNOWN
	}

	return -1
}

func (s *ConnectionState) ConnectionState() ice.ConnectionState {
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
	}

	return -1
}

func (s *ConnectionState) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(s.String())), nil
}
