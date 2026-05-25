// Package build holds the server endpoint injected at build time via ldflags.
// It is intentionally left empty in source; CI sets it via:
//
//	-ldflags "-X 'github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/build.Endpoint=...'"
//
// The value must never be printed in help text, error messages, logs, or panic output.
package build

// Endpoint is the base URL of the labeling server.
// Left empty by default; build pipeline injects the real value.
var Endpoint string
