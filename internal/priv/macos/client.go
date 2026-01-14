//go:build !darwin

package macos

// NewClient returns nil on non-darwin platforms.
func NewClient() Client { return nil }
