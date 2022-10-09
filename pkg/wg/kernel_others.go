//go:build !(linux || freebsd)

package wg

func KernelModuleExists() bool {
	return false
}
