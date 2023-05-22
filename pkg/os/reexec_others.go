//go:build !unix

package os

const ReexecSelfSupported = false

func ReexecSelf() error {
	return errNotSupported
}
