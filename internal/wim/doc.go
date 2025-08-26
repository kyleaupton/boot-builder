// Package wim provides a thin bridge to wimlib for WIM operations. Files in this
// package are gated by the 'wimlib' build tag when they require cgo.
//
// This package dynamically links against libwim and registers a progress
// callback to surface structured progress to Go. It intentionally keeps
// the C-facing details contained so the exported API is idiomatic Go.
package wim
