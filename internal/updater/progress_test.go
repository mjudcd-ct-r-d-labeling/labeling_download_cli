package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProgressReportsPartialTransfersAndThrottlesLogs(t *testing.T) {
	var out bytes.Buffer
	p := &downloadProgress{out: &out, total: 1024, phase: "Downloading", started: time.Now()}
	p.draw(true)
	p.Write(make([]byte, 256))
	if strings.Count(out.String(), "\n") != 1 {
		t.Fatal("progress logged every write")
	}
	p.lastDraw = time.Now().Add(-3 * time.Second)
	p.Write(make([]byte, 256))
	if !strings.Contains(out.String(), "50%  512 B / 1.0 KiB") {
		t.Fatalf("missing partial progress: %s", &out)
	}
	p.finish(context.Canceled)
	if !strings.Contains(out.String(), "Interrupted") || strings.Contains(out.String(), "Verified") || strings.Contains(out.String(), "\x1b") {
		t.Fatalf("incorrect redirected cancellation output: %s", &out)
	}
}

func TestTerminalProgressReusesLines(t *testing.T) {
	var out bytes.Buffer
	p := &downloadProgress{out: &out, total: 1024, phase: "Downloading", started: time.Now(), terminal: true}
	p.draw(true)
	p.Write(make([]byte, 1024))
	p.finish(nil)
	if !strings.Contains(out.String(), "\x1b[2A") || !strings.Contains(out.String(), "100%") || !strings.Contains(out.String(), "Verified") {
		t.Fatalf("incorrect terminal output: %q", out.String())
	}
}

func TestDownloadProgressOnlyReportsVerifiedAfterIntegrityCheck(t *testing.T) {
	data := []byte("downloaded CLI binary")
	hash := sha256.Sum256(data)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(data)
	}))
	defer srv.Close()
	for _, valid := range []bool{true, false} {
		t.Run(map[bool]string{true: "verified", false: "checksum failure"}[valid], func(t *testing.T) {
			info := release{URL: srv.URL, Size: int64(len(data)), SHA256: hex.EncodeToString(hash[:])}
			if !valid {
				info.SHA256 = strings.Repeat("0", 64)
			}
			// Capture actual redirected CLI output without invoking installation.
			out, err := os.CreateTemp(t.TempDir(), "output")
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			stdout := os.Stdout
			os.Stdout = out
			defer func() { os.Stdout = stdout }()
			path, downloadErr := download(context.Background(), srv.Client(), info)
			os.Stdout = stdout
			if path != "" {
				defer os.Remove(path)
			}
			if (downloadErr == nil) != valid {
				t.Fatalf("unexpected download result: %v", downloadErr)
			}
			output, err := os.ReadFile(out.Name())
			if err != nil {
				t.Fatal(err)
			}
			text := string(output)
			if !strings.Contains(text, "Verifying size and SHA-256") || strings.Contains(text, "\x1b") || strings.Contains(text, srv.URL) {
				t.Fatalf("incorrect progress output: %s", text)
			}
			if strings.Contains(text, "Status Verified") != valid || strings.Contains(text, "Status Failed") == valid {
				t.Fatalf("transfer completion confused with verification: %s", text)
			}
		})
	}
}
