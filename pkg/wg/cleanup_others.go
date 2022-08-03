//go:build !unix

package wg

import (
	"riasc.eu/wice/pkg/errors"
)

func CleanupUserSockets() error {
	return errors.ErrNotSupported
}
