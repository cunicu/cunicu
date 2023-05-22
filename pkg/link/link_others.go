//go:build !(linux || freebsd || darwin)

package link

func FindKernelDevice(name string) (Device, error) {
	return nil, errNotSupported
}

func NewKernelDevice(name string) (Device, error) {
	return nil, errNotSupported
}
