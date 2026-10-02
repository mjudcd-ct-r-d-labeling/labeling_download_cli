package build

import (
	"encoding/base64"
	"testing"
)

func TestBuildEndpointsAreIndependent(t *testing.T) {
	oldDataset, oldRelease := endpointB64, releaseEndpointB64
	t.Cleanup(func() { endpointB64, releaseEndpointB64 = oldDataset, oldRelease })
	endpointB64 = base64.StdEncoding.EncodeToString([]byte("https://dataset.example"))
	releaseEndpointB64 = base64.StdEncoding.EncodeToString([]byte("https://distribution.example"))
	if Endpoint() != "https://dataset.example" || ReleaseEndpoint() != "https://distribution.example" {
		t.Fatal("dataset and distribution endpoints were not injected independently")
	}
	releaseEndpointB64 = "invalid base64"
	if ReleaseEndpoint() != "" || Endpoint() != "https://dataset.example" {
		t.Fatal("invalid distribution endpoint affected dataset endpoint")
	}
	releaseEndpointB64 = ""
	if ReleaseEndpoint() != "" {
		t.Fatal("missing endpoint was accepted")
	}
}
