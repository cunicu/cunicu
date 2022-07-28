//go:build !linux

package routes

func (s *Syncer) syncKernel() {
	s.logger.Warn("Kernel to WireGuard route synchronization is not supported on this platform.")
}

func (s *Syncer) watchKernel() {

}
