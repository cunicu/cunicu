//go:build !unix

package util

import "riasc.eu/wice/pkg/errors"

const ReexecSelfSupported = false

func ReexecSelf() error {
	return errors.ErrNotSupported
}
