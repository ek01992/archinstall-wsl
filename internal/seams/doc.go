// Package seams provides helpers to safely override global seam variables in tests.
//
// Seams are package-level variables (e.g., functions for running commands or reading files)
// used to make code testable without performing real system operations. These globals are
// NOT concurrency-safe in production. Tests that override seams should use With to
// serialize overrides and ensure the original value is restored even if the test fails.
//
// Prefer dependency injection in production code when introducing concurrency or when
// test isolation needs to be guaranteed across packages.
package seams
