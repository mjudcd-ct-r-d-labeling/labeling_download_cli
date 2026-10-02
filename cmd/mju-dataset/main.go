// Command mju-dataset is the MJU Labeling Dataset Download CLI.
//
// Usage:
//
//	mju-dataset [--version | --update | --uninstall]
//
// The program prompts interactively for credentials and a local directory,
// then downloads the selected dataset from the labeling server.
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
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/listinput"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/manifest"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/secureinput"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/state"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/uninstall"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/updater"
	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version information and exit")
	update := flag.Bool("update", false, "Install the latest published CLI version")
	remove := flag.Bool("uninstall", false, "Remove the installed CLI and preserve downloaded datasets")
	// NOTE: --base-url / --server / --endpoint flags are intentionally omitted (SEC-001).
	flag.Parse()
	if flag.NArg() != 0 || (*showVersion && (*update || *remove)) || (*update && *remove) {
		fmt.Fprintln(os.Stderr, "Usage: mju-dataset [--version | --update | --uninstall]")
		os.Exit(2)
	}

	if *showVersion {
		version.Print()
		os.Exit(0)
	}
	if *remove {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := uninstall.Run(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Fprintln(os.Stderr, "Uninstall interrupted.")
				os.Exit(130)
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *update {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		if err := updater.Run(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Fprintln(os.Stderr, "Update interrupted.")
				os.Exit(130)
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// Guard: binary built without endpoint injection is unusable.
	if build.Endpoint() == "" {
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

	mode, err := promptMode(ctx)
	if err != nil {
		return err
	}
	var numbers []string
	if mode == 2 {
		value, readErr := secureinput.ReadLine("Classification number: ")
		if readErr != nil {
			return readErr
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("classification number is required")
		}
		numbers = []string{value}
	} else if mode == 3 {
		directory, readErr := promptListDirectory(ctx)
		if readErr != nil {
			return readErr
		}
		var fileCount int
		numbers, fileCount, err = listinput.LoadDirectory(directory)
		if err != nil {
			return err
		}
		fmt.Printf("Read %d classification numbers from %d list files.\n", len(numbers), fileCount)
	}
	fmt.Println()

	// ── Download directory ────────────────────────────────────────────────────
	downloadRoot, err := promptAbsPath(ctx)
	if err != nil {
		return err
	}
	fmt.Println()

	// ── Fetch download plan ───────────────────────────────────────────────────
	fmt.Print("Fetching file list from server... ")
	var entries []manifest.FileEntry
	var unavailable []string
	if mode == 1 {
		entries, err = manifest.Build(ctx, authed, downloadRoot)
	} else {
		var selection *manifest.Selection
		selection, err = manifest.BuildSelected(ctx, authed, downloadRoot, numbers)
		if err == nil {
			entries, unavailable = selection.Entries, selection.Unavailable
		}
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		fmt.Println()
		if mode != 1 {
			return err
		}
		return fmt.Errorf("Network error. Please try again.")
	}
	fmt.Println("done.")
	fmt.Println()
	if len(unavailable) > 0 {
		fmt.Printf("Unavailable: %d classification numbers or sessions.\n", len(unavailable))
		for i, reason := range unavailable {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(unavailable)-5)
				break
			}
			fmt.Println("  -", reason)
		}
		if err := writeUnavailableReport(downloadRoot, unavailable); err != nil {
			fmt.Println("Warning: could not save unavailable report.")
		} else {
			fmt.Println("Full report: " + filepath.Join(downloadRoot, state.DirName, "unavailable.txt"))
		}
		fmt.Println()
	} else if mode != 1 {
		_ = os.Remove(filepath.Join(downloadRoot, state.DirName, "unavailable.txt"))
	}

	if len(entries) == 0 {
		fmt.Println("No downloadable files are available for this selection.")
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
	if mode == 1 {
		fmt.Printf("Ready to download %d games / %d files.\n", gameCount, fileCount)
	} else {
		fmt.Printf("Ready to download %d games / %d sessions / %d files.\n", gameCount, manifest.UniqueSessions(entries), fileCount)
	}
	fmt.Print("Press Enter to start.")
	if _, readErr := secureinput.ReadLine(""); readErr != nil && !errors.Is(readErr, context.Canceled) {
		return readErr
	}
	fmt.Println()

	// ── Step 6: data_explain.md (non-fatal) ───────────────────────────────────
	downloader.FetchDataExplain(ctx, authed, downloadRoot)

	// ── Step 7: Main download loop ────────────────────────────────────────────
	sum := downloader.Run(ctx, authed, entries, fresh, downloadRoot)
	if err := ctx.Err(); err != nil {
		return err
	}

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

func promptMode(ctx context.Context) (int, error) {
	fmt.Println("[1] Download all data (existing export)")
	fmt.Println("[2] Download all sessions for one classification number")
	fmt.Println("[3] Download all sessions from list files in a directory")
	for {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		choice, err := secureinput.ReadLine("Select mode (1/2/3): ")
		if err != nil {
			return 0, err
		}
		switch strings.TrimSpace(choice) {
		case "1":
			return 1, nil
		case "2":
			return 2, nil
		case "3":
			return 3, nil
		default:
			fmt.Println("Please enter 1, 2, or 3.")
		}
	}
}

func promptListDirectory(ctx context.Context) (string, error) {
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		value, err := secureinput.ReadLine("List files directory (absolute path): ")
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if !filepath.IsAbs(value) {
			fmt.Println("Please enter an absolute directory path.")
			continue
		}
		info, err := os.Stat(value)
		if err != nil || !info.IsDir() {
			fmt.Println("List directory does not exist.")
			continue
		}
		return filepath.Clean(value), nil
	}
}

func writeUnavailableReport(root string, lines []string) error {
	directory := filepath.Join(root, state.DirName)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "unavailable.txt"), []byte(strings.Join(lines, "\n")+"\n"), 0644)
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
