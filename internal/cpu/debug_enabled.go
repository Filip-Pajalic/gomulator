//go:build debug

package cpu

// Debug build: DbgUpdate and DbgPrint are active
func stepDebugHook() bool {
	DbgUpdate()
	if !DbgPrint() {
		return false
	}
	DbgProgress()
	return true
}
