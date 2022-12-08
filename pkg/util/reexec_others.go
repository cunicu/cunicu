//go:build !unix

package util

const ReexecSelfSupported = false

func ReexecSelf() error {
	return errNotSupported
}
