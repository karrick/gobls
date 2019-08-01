// +build !gobls_debug

package gobls

// debug is a no-op for release builds
func debug(_ string, _ ...interface{}) {}
