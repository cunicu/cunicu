// Package errors defines custom project-wide errors
package errors

import "errors"

var (
	ErrNotSupported   = errors.New("not supported on this platform")
	ErrAlreadyStopped = errors.New("already stopped")
)
