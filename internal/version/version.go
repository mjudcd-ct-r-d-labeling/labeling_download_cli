// Package version exposes build-time metadata injected via ldflags.
// The server Endpoint is intentionally excluded from all output here.
package version

import (
	"fmt"
	"runtime"
)

// Variables are set at build time via:
//
//	-ldflags "-X 'github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version.Version=1.0.0'
//	          -X 'github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version.Commit=abc123'
//	          -X 'github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version.BuildDate=2026-05-25'"
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Print writes version information to stdout.
// Server endpoint is never included.
func Print() {
	fmt.Printf("mju-dataset %s (commit %s, built %s, %s/%s)\n",
		Version, Commit, BuildDate, runtime.GOOS, runtime.GOARCH)
}

// String returns the version string.
func String() string {
	return Version
}
