// Command mju-dataset is the MJU Labeling Dataset Download CLI.
//
// Usage:
//
//	mju-dataset [--version]
//
// The program prompts interactively for credentials and a local directory,
// then downloads the full dataset from the labeling server.
//
// Security note: --base-url / --server / --endpoint options are intentionally
// absent.  The server address is injected at build time and never exposed to
// the user (SEC-001).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/auth"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/build"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/client"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/downloader"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/manifest"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/secureinput"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version information and exit")
	// NOTE: --base-url / --server / --endpoint flags are intentionally omitted (SEC-001).
	flag.Parse()

	if *showVersion {
		version.Print()
		os.Exit(0)
	}

	// Guard: binary built without endpoint injection is unusable.
	if build.Endpoint == "" {
		fmt.Fprintln(os.Stderr, "This binary was not built with a server endpoint.")
		fmt.Fprintln(os.Stderr, "Please use an official release from the project page.")
		os.Exit(1)
	}

	// Honour Ctrl+C / SIGTERM with graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintln(os.Stderr, "\nDownload interrupted. Run again to resume.")
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

// run executes the full interactive download flow.
func run(ctx context.Context) error {
	fmt.Println("MJU Labeling Dataset Downloader")
	fmt.Println()

	// ── Step 1: Authentication ────────────────────────────────────────────────
	c := client.New()
	token, err := auth.Authenticate(ctx, c)
	if err != nil {
		return err
	}
	fmt.Println()
	authed := c.WithToken(token)

	// ── Step 2: Download directory ────────────────────────────────────────────
	downloadRoot, err := promptAbsPath(ctx)
	if err != nil {
		return err
	}
	fmt.Println()

	// ── Step 3: Fetch download plan ───────────────────────────────────────────
	fmt.Print("Fetching file list from server... ")
	entries, err := manifest.Build(ctx, authed, downloadRoot)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		fmt.Println()
		return fmt.Errorf("Network error. Please try again.")
	}
	fmt.Println("done.")
	fmt.Println()

	if len(entries) == 0 {
		fmt.Println("No downloadable files are available on the server.")
		return nil
	}

	gameCount := manifest.UniqueGames(entries)
	fileCount := len(entries)

	// ── Step 4: Resume / Fresh selection ─────────────────────────────────────
	existing := downloader.ScanExisting(entries)
	fresh := false
	if existing > 0 {
		fresh, err = promptResumeOrFresh(ctx, existing)
		if err != nil {
			return err
		}
		fmt.Println()
	}

	// ── Step 5: Confirm and start ─────────────────────────────────────────────
	fmt.Printf("Ready to download %d games / %d files.\n", gameCount, fileCount)
	fmt.Print("Press Enter to start.")
	if _, readErr := secureinput.ReadLine(""); readErr != nil && !errors.Is(readErr, context.Canceled) {
		return readErr
	}
	fmt.Println()

	// ── Step 6: data_explain.md (non-fatal) ───────────────────────────────────
	downloader.FetchDataExplain(ctx, authed, downloadRoot)

	// ── Step 7: Main download loop ────────────────────────────────────────────
	sum := downloader.Run(ctx, authed, entries, fresh, downloadRoot)

	// ── Step 8: Summary ───────────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("Done.  Success: %d  Skipped: %d  Failed: %d\n",
		sum.Success, sum.Skipped, len(sum.Failed))

	if len(sum.Failed) > 0 {
		fmt.Println("\nFailed files:")
		for _, f := range sum.Failed {
			fmt.Println("  -", f)
		}
		return fmt.Errorf("Completed with errors.")
	}
	return nil
}

// promptAbsPath prompts for a download directory, enforcing absolute paths,
// creating missing directories on confirmation, and testing write permission.
func promptAbsPath(ctx context.Context) (string, error) {
	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		raw, err := secureinput.ReadLine("Download directory (absolute path): ")
		if err != nil {
			return "", err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		// Reject relative paths (FR-PATH-001).
		if !filepath.IsAbs(raw) {
			fmt.Println("Please enter an absolute path (e.g. /home/user/mju_dataset or C:\\Users\\name\\mju_dataset).")
			continue
		}

		clean := filepath.Clean(raw)

		fi, statErr := os.Stat(clean)
		if os.IsNotExist(statErr) {
			// Offer to create (FR-PATH-002).
			fmt.Printf("Directory does not exist. Create %q? (y/N): ", clean)
			ans, _ := secureinput.ReadLine("")
			if strings.ToLower(strings.TrimSpace(ans)) != "y" {
				continue
			}
			if mkErr := os.MkdirAll(clean, 0755); mkErr != nil {
				fmt.Println("Cannot write to the selected directory.")
				continue
			}
		} else if statErr != nil {
			fmt.Println("Cannot write to the selected directory.")
			continue
		} else if !fi.IsDir() {
			fmt.Println("Path exists but is not a directory. Please choose a directory.")
			continue
		}

		// Write permission test (FR-PATH-003).
		probe := filepath.Join(clean, ".mju-write-probe")
		if wErr := os.WriteFile(probe, []byte("ok"), 0600); wErr != nil {
			fmt.Println("Cannot write to the selected directory.")
			continue
		}
		_ = os.Remove(probe)

		return clean, nil
	}
}

// promptResumeOrFresh shows the Resume / Fresh options and returns
// fresh=true when the user chooses to overwrite existing data.
func promptResumeOrFresh(ctx context.Context, existingCount int) (bool, error) {
	fmt.Printf("Existing dataset files were found (%d verified).\n", existingCount)
	fmt.Println("[1] Resume - skip verified files, download missing/corrupted ones")
	fmt.Println("[2] Fresh  - remove/overwrite existing files and download everything again")

	for {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		choice, err := secureinput.ReadLine("Select option (1/2): ")
		if err != nil {
			return false, err
		}
		switch strings.TrimSpace(choice) {
		case "1":
			return false, nil
		case "2":
			// Second confirmation before destructive overwrite (FR-PATH-005).
			fmt.Print("This will overwrite all existing dataset files. Are you sure? (y/N): ")
			confirm, _ := secureinput.ReadLine("")
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				return true, nil
			}
			fmt.Println("Cancelled. Returning to options.")
		default:
			fmt.Println("Please enter 1 or 2.")
		}
	}
}
