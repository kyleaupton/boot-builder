//go:build darwin && !osinstallxpc

package macos

// NewClient returns nil on darwin builds when the osinstallxpc tag is not set.
func NewClient() Client { return nil }
