//go:build darwin && !flashitxpc

package macos

// NewClient returns nil on darwin builds when the flashitxpc tag is not set.
func NewClient() Client { return nil }
