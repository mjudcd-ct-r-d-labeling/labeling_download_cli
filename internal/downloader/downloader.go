// Package downloader streams files from the labeling server with progress
// display, retry, resume support, and atomic rename.
package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/client"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/manifest"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/resume"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/state"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version"
)

const (
	maxRetries  = 3
	readBufSize = 32 * 1024
)

// Summary holds the final outcome of a Run call.
type Summary struct {
	Success int
	Skipped int
	Failed  []string
}

// ScanExisting counts files in entries that are already fully present on disk.
func ScanExisting(entries []manifest.FileEntry) int {
	count := 0
	for _, e := range entries {
		if s, _ := resume.Check(e); s == resume.StatusComplete {
			count++
		}
	}
	return count
}

// FetchDataExplain downloads GET /download/data_explain.md and saves it to
// downloadRoot/data_explain.md.  Errors are non-fatal and only printed as a
// warning.
func FetchDataExplain(ctx context.Context, c *client.Client, downloadRoot string) {
	data, err := c.GetBytes(ctx, "/download/data_explain.md")
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Println("  Warning: could not fetch data_explain.md")
		}
		return
	}
	dest := filepath.Join(downloadRoot, "data_explain.md")
	if err := os.WriteFile(dest, data, 0644); err != nil {
		fmt.Println("  Warning: could not write data_explain.md")
	}
}

// Run iterates entries, skipping complete files (unless fresh=true), and
// downloads the rest.  It respects ctx cancellation (Ctrl+C).
// stateDir is the path to the .mju-dataset-download directory.
func Run(ctx context.Context, c *client.Client, entries []manifest.FileEntry, fresh bool, downloadRoot string) Summary {
	st := state.Load(downloadRoot)
	st.LastRun = time.Now().UTC()
	st.CLIVersion = version.String()
	if fresh {
		st.Completed = nil
		st.Failed = nil
	}

	completedSet := make(map[string]bool, len(st.Completed))
	for _, p := range st.Completed {
		completedSet[p] = true
	}

	var sum Summary
	total := len(entries)
	newFailed := make([]string, 0)

	for i, entry := range entries {
		if ctx.Err() != nil {
			break
		}

		relPath := filepath.Join(entry.CN, entry.Filename)
		label := fmt.Sprintf("[%d/%d] %s", i+1, total, entry.Filename)

		fileStatus, partSize := resume.Check(entry)

		// Skip already verified files (resume mode).
		if fileStatus == resume.StatusComplete && !fresh && completedSet[relPath] {
			fmt.Printf("  SKIP   %s\n", entry.Filename)
			sum.Skipped++
			continue
		}
		// Re-check corrupt files even without fresh flag.
		if fileStatus == resume.StatusCorrupt {
			_ = resume.RemoveCorrupt(entry)
			partSize = 0
			fileStatus = resume.StatusMissing
		}

		if fresh {
			// Remove existing .part and final file for a clean start.
			_ = os.Remove(resume.FinalPath(entry))
			_ = os.Remove(resume.PartPath(entry))
			partSize = 0
		}

		// Ensure local CN sub-directory exists.
		if err := os.MkdirAll(entry.LocalDir, 0755); err != nil {
			fmt.Printf("  FAIL   %s (cannot create directory)\n", entry.Filename)
			newFailed = append(newFailed, relPath)
			state.AppendLog(downloadRoot, time.Now().UTC(), "FAIL", entry.Filename, "directory creation error")
			continue
		}

		err := downloadWithRetry(ctx, c, entry, partSize, label)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// Preserve state on interrupt so the user can resume.
				break
			}
			fmt.Printf("\n  FAIL   %s\n", entry.Filename)
			newFailed = append(newFailed, relPath)
			state.AppendLog(downloadRoot, time.Now().UTC(), "FAIL", entry.Filename, "download failed after retries")
		} else {
			sum.Success++
			completedSet[relPath] = true
			state.AppendLog(downloadRoot, time.Now().UTC(), "SUCCESS", entry.Filename, "")
		}
	}

	// Rebuild completed list from the running set.
	completed := make([]string, 0, len(completedSet))
	for p := range completedSet {
		completed = append(completed, p)
	}
	st.Completed = completed
	st.Failed = newFailed
	_ = state.Save(downloadRoot, st)

	sum.Failed = newFailed
	return sum
}

// downloadWithRetry retries up to maxRetries times with exponential backoff.
func downloadWithRetry(ctx context.Context, c *client.Client, entry manifest.FileEntry, partSize int64, label string) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt > 0 {
			wait := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			fmt.Printf("  Retry %d/%d in %v...\n", attempt, maxRetries-1, wait)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}
		lastErr = downloadFile(ctx, c, entry, partSize, label)
		if lastErr == nil {
			return nil
		}
		if errors.Is(lastErr, context.Canceled) {
			return lastErr
		}
		// Reset part offset so the next attempt retries from scratch when
		// the server does not support Range (handled inside downloadFile).
		partSize = 0
	}
	return lastErr
}

// downloadFile performs a single streaming download attempt.
func downloadFile(ctx context.Context, c *client.Client, entry manifest.FileEntry, offset int64, label string) error {
	resp, err := c.GetStream(ctx, entry.DownloadPath, offset)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Server returned 200 (not 206) despite our Range request → reset offset.
	if resp.StatusCode == http.StatusOK && offset > 0 {
		offset = 0
		_ = os.Remove(resume.PartPath(entry))
	}

	partPath := resume.PartPath(entry)
	var openFlags int
	if offset > 0 {
		openFlags = os.O_APPEND | os.O_WRONLY
	} else {
		openFlags = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	}
	f, err := os.OpenFile(partPath, openFlags, 0644)
	if err != nil {
		return fmt.Errorf("cannot open temporary file")
	}

	// Determine total for the progress bar.
	totalSize := entry.Size
	if resp.ContentLength > 0 {
		totalSize = offset + resp.ContentLength
	}

	bar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription(label),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(25),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionOnCompletion(func() { fmt.Println() }),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	if offset > 0 {
		_ = bar.Add64(offset)
	}

	buf := make([]byte, readBufSize)
	for {
		if ctx.Err() != nil {
			_ = f.Close()
			return ctx.Err()
		}
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			wn, werr := f.Write(buf[:n])
			_ = bar.Add(wn)
			if werr != nil {
				_ = f.Close()
				return fmt.Errorf("write error")
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = f.Close()
			return fmt.Errorf("Network error. Please try again.")
		}
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("file close error")
	}

	// Atomic rename: .part → final filename (FR-RESUME-002).
	if err := os.Rename(partPath, resume.FinalPath(entry)); err != nil {
		return fmt.Errorf("rename error")
	}
	return nil
}
