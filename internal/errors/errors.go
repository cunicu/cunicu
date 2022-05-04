package errors

import "errors"

var (
	ErrNotSupported = errors.New("not supported on this platform")
)
