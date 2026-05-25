// Package resume implements local file verification and .part-based resume logic.
package resume

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/manifest"
)

// FileStatus categorises the state of a local file for a given FileEntry.
type FileStatus int

const (
	// StatusMissing: neither the final file nor a .part file exists.
	StatusMissing FileStatus = iota
	// StatusPartial: a .part file exists (download was interrupted).
	StatusPartial
	// StatusCorrupt: the final file exists but its size or SHA-256 does not match.
	StatusCorrupt
	// StatusComplete: the final file exists and passes all available checks.
	StatusComplete
)

// Check returns the FileStatus for entry and, when StatusPartial, the byte
// count already written to the .part file (used as the Range resume offset).
func Check(entry manifest.FileEntry) (FileStatus, int64) {
	final := FinalPath(entry)
	part := PartPath(entry)

	// Check final file first.
	if fi, err := os.Stat(final); err == nil {
		if entry.Size > 0 && fi.Size() != entry.Size {
			return StatusCorrupt, 0
		}
		if entry.SHA256 != "" {
			if h, err := hashFile(final); err == nil && h != entry.SHA256 {
				return StatusCorrupt, 0
			}
		}
		return StatusComplete, fi.Size()
	}

	// Check .part file.
	if fi, err := os.Stat(part); err == nil {
		return StatusPartial, fi.Size()
	}

	return StatusMissing, 0
}

// RemoveCorrupt renames the corrupt final file to .corrupt so the original
// name is free for re-download without permanently discarding the data.
func RemoveCorrupt(entry manifest.FileEntry) error {
	return os.Rename(FinalPath(entry), FinalPath(entry)+".corrupt")
}

// PartPath returns the .part path for entry (in-progress download target).
func PartPath(entry manifest.FileEntry) string {
	return filepath.Join(entry.LocalDir, entry.Filename+".part")
}

// FinalPath returns the final file path for entry.
func FinalPath(entry manifest.FileEntry) string {
	return filepath.Join(entry.LocalDir, entry.Filename)
}

// hashFile computes the SHA-256 hex digest of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
