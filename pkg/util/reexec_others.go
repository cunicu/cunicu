//go:build !unix

package util

import "github.com/stv0g/cunicu/pkg/errors"

const ReexecSelfSupported = false

func ReexecSelf() error {
	return errors.ErrNotSupported
}
