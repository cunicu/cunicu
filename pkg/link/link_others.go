//go:build !(linux || freebsd || darwin)

package link

func CreateWireGuardLink(_ string) (Link, error) {
	return nil, errNotSupported
}

func FindLink(_ string) (Link, error) {
	return nil, errNotSupported
}
