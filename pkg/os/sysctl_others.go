//go:build !linux

package os

func SetSysctl(_ string, _ any) error {
	return errNotSupported
}

func SetSysctlMap(_ map[string]any) error {
	return errNotSupported
}
