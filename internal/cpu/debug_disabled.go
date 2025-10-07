//go:build !debug

package cpu

// Release build: No debug overhead
func stepDebugHook() bool {
	return true // Always continue
}
