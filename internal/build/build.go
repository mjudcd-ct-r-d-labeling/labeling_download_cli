// Package build holds the server endpoint injected at build time via ldflags.
// It is intentionally left empty in source; CI sets it via:
//
//	-ldflags "-X 'github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/build.endpointB64=...'"
//
// The value is stored base64-encoded to prevent trivial extraction with strings(1).
// It must never be printed in help text, error messages, logs, or panic output.
package build

import "encoding/base64"

// endpointB64 holds the base64-encoded server base URL.
// Left empty by default; build pipeline injects the real value.
var endpointB64 string

// releaseEndpointB64 is the distribution API address injected from API_BASE.
// It is separate from the dataset endpoint and must not be printed.
var releaseEndpointB64 string

// ReleaseEndpoint returns the build-time distribution API address.
func ReleaseEndpoint() string {
	return decodeEndpoint(releaseEndpointB64)
}

// Endpoint returns the decoded server base URL.
// Returns an empty string when the binary was built without an endpoint.
func Endpoint() string {
	return decodeEndpoint(endpointB64)
}

func decodeEndpoint(encoded string) string {
	if encoded == "" {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	return string(b)
}
