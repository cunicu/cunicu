//go:build !linux

package wg

func KernelModuleExists() bool {
	return false
}
