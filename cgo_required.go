package main

import _ "runtime/cgo"

// The blank import above enforces that CGO is enabled when building the main application
// binary. When CGO is disabled, the runtime/cgo package is unavailable, and the build
// will fail fast with a clear error instead of silently falling back to reduced
// functionality.
