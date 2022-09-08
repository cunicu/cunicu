//go:build !unix

package wg

import (
	"github.com/stv0g/cunicu/pkg/errors"
)

func CleanupUserSockets() error {
	return errors.ErrNotSupported
}
