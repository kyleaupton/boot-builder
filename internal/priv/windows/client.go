//go:build !windows

package windows

// NewClient returns nil on non-Windows platforms.
func NewClient() Client { return nil }
