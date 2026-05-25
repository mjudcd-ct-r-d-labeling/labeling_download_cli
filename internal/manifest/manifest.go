// Package manifest builds the download plan from the server's classification list.
package manifest

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/client"
)

// FileEntry describes a single file to download.
type FileEntry struct {
	CN           string // e.g. "GC-2024-0001"
	FileType     string // "gameplay" | "inputlogs" | "labeling"
	Filename     string // e.g. "GC-2024-0001_gameplay.mp4"
	DownloadPath string // server path: "/exports/file/GC-2024-0001/gameplay"
	LocalDir     string // absolute local directory for this CN
	Size         int64  // bytes; 0 = unknown (server does not yet provide this)
	SHA256       string // hex digest; empty = no server-side checksum available
}

// fileTypes defines the ordered set of file types per classification number.
var fileTypes = []struct {
	name string
	ext  string
}{
	{"gameplay", ".mp4"},
	{"inputlogs", ".jsonl"},
	{"labeling", ".jsonl"},
}

// listResponse matches GET /exports/list response body.
type listResponse struct {
	ClassificationNumbers []string `json:"classification_numbers"`
	Total                 int      `json:"total"`
}

// Build fetches the server classification list and returns a flat slice of
// FileEntry values ready for download.  downloadRoot is the user-supplied
// absolute path where data will be saved.
func Build(ctx context.Context, c *client.Client, downloadRoot string) ([]FileEntry, error) {
	var resp listResponse
	if err := c.GetJSON(ctx, "/exports/list", &resp); err != nil {
		return nil, fmt.Errorf("fetching export list: %w", err)
	}

	if len(resp.ClassificationNumbers) == 0 {
		return nil, nil
	}

	entries := make([]FileEntry, 0, len(resp.ClassificationNumbers)*len(fileTypes))
	for _, cn := range resp.ClassificationNumbers {
		for _, ft := range fileTypes {
			entries = append(entries, FileEntry{
				CN:           cn,
				FileType:     ft.name,
				Filename:     cn + "_" + ft.name + ft.ext,
				DownloadPath: "/exports/file/" + cn + "/" + ft.name,
				LocalDir:     filepath.Join(downloadRoot, cn),
			})
		}
	}
	return entries, nil
}

// UniqueGames returns the number of distinct classification numbers in entries.
func UniqueGames(entries []FileEntry) int {
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		seen[e.CN] = struct{}{}
	}
	return len(seen)
}
