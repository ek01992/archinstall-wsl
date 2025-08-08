// Package config handles safe load/save of user configuration with defaults.
//
// This package uses small, package-level seams for filesystem operations to ease testing.
// NOTE: These seams are NOT concurrency-safe. Use internal/seams.With to serialize overrides
// in tests. Prefer dependency injection when introducing concurrency.
package config
