//go:build !unix

package wg

func CleanupUserSockets() error {
	return errNotSupported
}
