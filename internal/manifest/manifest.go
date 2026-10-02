// Package manifest builds the download plan from the server's classification list.
package manifest

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/client"
)

// FileEntry describes a single file to download.
type FileEntry struct {
	CN           string // e.g. "GC-2024-0001"
	SessionID    string // empty for the original full-download mode
	FileType     string // "gameplay" | "inputlogs" | "labeling"
	Filename     string // e.g. "GC-2024-0001_gameplay.mp4"
	DownloadPath string // server path: "/exports/file/GC-2024-0001/gameplay"
	LocalDir     string // absolute local directory for this CN
	Size         int64  // bytes; 0 = unknown (server does not yet provide this)
	SHA256       string // hex digest; empty = no server-side checksum available
	ETag         string // exact ETag from the selected session manifest
}

var safeCN = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
var safeSessionID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
var safeSHA256 = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

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
		if !safeCN.MatchString(cn) {
			return nil, fmt.Errorf("invalid classification number in server response")
		}
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

type selectedResponse struct {
	Classifications []struct {
		ClassificationNumber string `json:"classification_number"`
		Status               string `json:"status"`
		Reason               string `json:"reason"`
		Sessions             []struct {
			SessionID            string   `json:"session_id"`
			ClassificationNumber string   `json:"classification_number"`
			Status               string   `json:"status"`
			Reason               string   `json:"reason"`
			MissingFileTypes     []string `json:"missing_file_types"`
			Files                []struct {
				FileType     string `json:"file_type"`
				FileName     string `json:"file_name"`
				DownloadPath string `json:"download_path"`
				SizeBytes    int64  `json:"size_bytes"`
				SHA256       string `json:"sha256"`
				Revision     string `json:"revision"`
				ETag         string `json:"etag"`
			} `json:"files"`
		} `json:"sessions"`
	} `json:"classifications"`
}

// Selection contains a complete download plan and reasons for requested data
// that could not be downloaded. The server returns every requested number.
type Selection struct {
	Entries       []FileEntry
	Unavailable   []string
	ReadySessions int
}

// BuildSelected resolves all sessions for each requested classification number.
// The server accepts 100 numbers per request, so large lists are batched here.
func BuildSelected(ctx context.Context, c *client.Client, downloadRoot string, numbers []string) (*Selection, error) {
	plan := &Selection{Entries: []FileEntry{}, Unavailable: []string{}}
	if len(numbers) == 0 {
		return plan, nil
	}
	for start := 0; start < len(numbers); start += 100 {
		end := start + 100
		if end > len(numbers) {
			end = len(numbers)
		}
		var response selectedResponse
		if err := c.PostJSON(ctx, "/exports/classifications/manifest", map[string][]string{"classification_numbers": numbers[start:end]}, &response); err != nil {
			return nil, fmt.Errorf("could not fetch session manifest: %w", err)
		}
		if len(response.Classifications) != end-start {
			return nil, fmt.Errorf("incomplete session manifest from server")
		}
		for i, item := range response.Classifications {
			cn := numbers[start+i]
			if item.ClassificationNumber != cn {
				return nil, fmt.Errorf("invalid classification number in session manifest")
			}
			if item.Status != "found" {
				plan.Unavailable = append(plan.Unavailable, fmt.Sprintf("%q: %s", cn, item.Reason))
				continue
			}
			if !safeCN.MatchString(cn) {
				return nil, fmt.Errorf("invalid classification number in session manifest")
			}
			for _, session := range item.Sessions {
				if session.ClassificationNumber != cn || !safeSessionID.MatchString(session.SessionID) {
					return nil, fmt.Errorf("invalid session ID in session manifest")
				}
				if session.Status != "ready" {
					plan.Unavailable = append(plan.Unavailable, cn+"/"+session.SessionID+": "+session.Reason+" ("+strings.Join(session.MissingFileTypes, ", ")+")")
					continue
				}
				if len(session.Files) != 3 {
					return nil, fmt.Errorf("ready session has an incomplete file list")
				}
				seenTypes := make(map[string]bool)
				for _, file := range session.Files {
					if file.FileType != "gameplay" && file.FileType != "inputlogs" && file.FileType != "labeling" || seenTypes[file.FileType] {
						return nil, fmt.Errorf("invalid file type in session manifest")
					}
					seenTypes[file.FileType] = true
					ext := ".jsonl"
					if file.FileType == "gameplay" {
						ext = ".mp4"
					}
					expectedName := cn + "_" + file.FileType + ext
					expectedPath := "/exports/sessions/" + session.SessionID + "/files/" + file.FileType + "?revision=" + file.Revision
					if file.FileName != expectedName || file.DownloadPath != expectedPath || !safeSHA256.MatchString(file.Revision) || file.SizeBytes < 0 || !safeSHA256.MatchString(file.SHA256) || file.ETag != `"`+file.Revision+`"` {
						return nil, fmt.Errorf("invalid file metadata in session manifest")
					}
					plan.Entries = append(plan.Entries, FileEntry{
						CN: cn, SessionID: session.SessionID, FileType: file.FileType,
						Filename: file.FileName, DownloadPath: file.DownloadPath,
						LocalDir: filepath.Join(downloadRoot, cn, session.SessionID),
						Size:     file.SizeBytes, SHA256: strings.ToLower(file.SHA256), ETag: file.ETag,
					})
				}
				plan.ReadySessions++
			}
		}
	}
	return plan, nil
}

func UniqueSessions(entries []FileEntry) int {
	seen := make(map[string]bool)
	for _, entry := range entries {
		if entry.SessionID != "" {
			seen[entry.SessionID] = true
		}
	}
	return len(seen)
}

// UniqueGames returns the number of distinct classification numbers in entries.
func UniqueGames(entries []FileEntry) int {
	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		seen[e.CN] = struct{}{}
	}
	return len(seen)
}
