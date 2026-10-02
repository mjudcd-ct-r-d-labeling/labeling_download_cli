package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/manifest"
)

func TestVerifyDownloadedFileChecksSizeAndSHA256(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.part")
	data := []byte("session data")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	entry := manifest.FileEntry{Size: int64(len(data)), SHA256: hex.EncodeToString(digest[:])}
	if err := verifyDownloadedFile(entry, path); err != nil {
		t.Fatalf("valid file rejected: %v", err)
	}
	entry.SHA256 = hex.EncodeToString(make([]byte, sha256.Size))
	if err := verifyDownloadedFile(entry, path); err == nil {
		t.Fatal("corrupt checksum accepted")
	}
	entry.Size++
	if err := verifyDownloadedFile(entry, path); err == nil {
		t.Fatal("wrong size accepted")
	}
}
