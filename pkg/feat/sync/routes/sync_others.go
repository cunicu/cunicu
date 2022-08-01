//go:build !linux

package routes

func (s *RouteSynchronization) syncKernel() {
	s.logger.Error("Kernel to WireGuard route synchronization is not supported on this platform.")
}

func (s *RouteSynchronization) watchKernel() {
	s.logger.Error("Kernel to WireGuard route synchronization is not supported on this platform.")
}
