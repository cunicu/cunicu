package pb

import (
	"time"
)

var Success = &Error{
	Code: Error_SUCCESS,
}

var NotSupported = &Error{
	Code:    Error_ENOTSUP,
	Message: "not supported yet",
}

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
