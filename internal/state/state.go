// Package state manages the persistent download state stored under
// {downloadRoot}/.mju-dataset-download/.
//
// Security invariant: credentials and the server endpoint are never written
// to state.json or download.log (FR-AUTH-005, FR-UX-004, SEC-006).
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DirName is the hidden metadata directory created inside the download root.
	DirName = ".mju-dataset-download"

	stateFile = "state.json"
	// LogFile is the append-only log of file-level download events.
	LogFile = "download.log"
)

// State is the persisted progress record for a download session.
// It is used to decide which files can be skipped on resume.
type State struct {
	// Completed contains relative paths (CN/filename) of successfully downloaded files.
	Completed []string `json:"completed"`
	// Failed contains relative paths of files that failed on the last run.
	Failed    []string  `json:"failed"`
	LastRun   time.Time `json:"last_run"`
	CLIVersion string   `json:"cli_version"`
}

// Load reads state.json from downloadRoot.  Missing or malformed files
// return an empty State (safe default for first run).
func Load(downloadRoot string) *State {
	path := filepath.Join(downloadRoot, DirName, stateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return &State{}
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return &State{}
	}
	return &s
}

// Save persists s to state.json inside downloadRoot.
func Save(downloadRoot string, s *State) error {
	dir := filepath.Join(downloadRoot, DirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, stateFile), data, 0644)
}

// AppendLog appends a single log entry to download.log.
// The entry must not contain credentials or the server endpoint.
func AppendLog(downloadRoot string, ts time.Time, level, filename, detail string) {
	dir := filepath.Join(downloadRoot, DirName)
	_ = os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, LogFile)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	line := fmt.Sprintf("%s  %-7s  %s", ts.UTC().Format(time.RFC3339), level, filename)
	if detail != "" {
		line += "  " + detail
	}
	_, _ = fmt.Fprintln(f, line)
}
