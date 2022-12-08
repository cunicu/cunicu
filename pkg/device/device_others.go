//go:build !(linux || freebsd || darwin)

package device

func FindKernelDevice(name string) (Device, error) {
	return nil, errNotSupported
}

func NewKernelDevice(name string) (Device, error) {
	return nil, errNotSupported
}
