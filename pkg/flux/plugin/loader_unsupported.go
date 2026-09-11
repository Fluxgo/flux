//go:build !linux && !darwin && !freebsd

package plugin

// loadSharedObject is a no-op on platforms where Go's plugin package is not
// supported. The rest of Flux remains fully usable on those platforms.
func (m *Manager) loadSharedObject(_ string, _ map[string]Config) error {
	return nil
}
