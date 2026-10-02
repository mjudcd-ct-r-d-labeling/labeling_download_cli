package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"2026.10.02.12", "2026.10.02.9", 1},
		{"2026.09.30.99", "2026.10.01.1", -1},
		{"2026.10.02.1", "2026.10.02.1", 0},
		{"1.2.3", "1.2.3.0", 0},
	} {
		got, err := compareVersions(tc.a, tc.b)
		if err != nil || got != tc.want {
			t.Fatalf("compare %s %s: %d %v", tc.a, tc.b, got, err)
		}
	}
	for _, bad := range []string{"dev", "../../bin", "1.2", "1.2.3;sh", "18446744073709551616.1.1"} {
		if _, err := compareVersions(bad, "1.2.3"); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}

func TestLatestVersionScansPagesAndIgnoresUnpublishedNames(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("update must not send dataset credentials")
		}
		if r.URL.Query().Get("page") == "1" {
			tags := make([]map[string]string, 100)
			for i := range tags {
				tags[i] = map[string]string{"name": "2026.10.02.9"}
			}
			_ = json.NewEncoder(w).Encode(tags)
		} else {
			fmt.Fprint(w, `[{"name":"2026.10.02.12"},{"name":"preview"},{"name":"999999999999999999999.1.1"}]`)
		}
	}))
	defer srv.Close()
	got, err := latestVersion(context.Background(), srv.Client(), srv.URL)
	if err != nil || got != "2026.10.02.12" {
		t.Fatalf("latest: %s %v", got, err)
	}
}

func TestDownloadIntegrity(t *testing.T) {
	data := []byte("verified executable bytes")
	h := sha256.Sum256(data)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(data) }))
	defer srv.Close()
	good := release{URL: srv.URL, Size: int64(len(data)), SHA256: hex.EncodeToString(h[:])}
	for _, tc := range []struct {
		name string
		size int64
		hash string
		ok   bool
	}{
		{"valid", good.Size, good.SHA256, true},
		{"short", good.Size + 1, good.SHA256, false},
		{"excess", good.Size - 1, good.SHA256, false},
		{"wrong hash", good.Size, strings.Repeat("0", 64), false},
		{"missing hash", good.Size, "", false},
		{"missing size", 0, good.SHA256, false},
		{"oversized", maxBinarySize + 1, good.SHA256, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			info := good
			info.Size = tc.size
			info.SHA256 = tc.hash
			path, err := download(context.Background(), srv.Client(), info)
			if tc.ok {
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(path)
				got, err := os.ReadFile(path)
				if err != nil || string(got) != string(data) {
					t.Fatalf("download: %q %v", got, err)
				}
			} else {
				if err == nil {
					os.Remove(path)
					t.Fatal("accepted corrupt download")
				}
				if path != "" {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatal("failed download was not cleaned up")
					}
				}
			}
		})
	}
}

func TestRequestErrorsDoNotExposeURL(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) }))
	defer srv.Close()
	_, err := request(context.Background(), srv.Client(), srv.URL+"/?signed-secret=value")
	if err == nil || strings.Contains(err.Error(), srv.URL) || strings.Contains(err.Error(), "signed-secret") {
		t.Fatalf("unsafe error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = request(ctx, srv.Client(), srv.URL)
	if err != context.Canceled {
		t.Fatalf("cancellation: %v", err)
	}
}

func TestPlatformBinary(t *testing.T) {
	for _, goos := range []string{"darwin", "linux", "windows"} {
		_, name, err := platformBinary(goos, "amd64")
		if err != nil {
			t.Fatal(err)
		}
		if (goos == "windows") != strings.HasSuffix(name, ".exe") {
			t.Fatalf("incorrect filename: %s", name)
		}
	}
	if _, _, err := platformBinary("windows", "arm64"); err == nil {
		t.Fatal("unsupported platform accepted")
	}
}

func TestStagePreservesOriginal(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "download")
	target := filepath.Join(dir, "installed")
	if err := os.WriteFile(source, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	path, err := stage(source, target)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	old, _ := os.ReadFile(target)
	if string(old) != "old" {
		t.Fatal("staging modified installed CLI")
	}
	newBytes, _ := os.ReadFile(path)
	if string(newBytes) != "new" {
		t.Fatal("incorrect staging")
	}
}
