//go:build !windows

package uninstall

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRemovePreservesOtherFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "mju-dataset")
	data := filepath.Join(dir, "dataset.jsonl")
	metadata := filepath.Join(dir, ".mju-dataset-download")
	if err := os.Mkdir(metadata, 0700); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(metadata, "state.json")
	for _, path := range []string{target, data, state} {
		if err := os.WriteFile(path, []byte("keep"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := remove(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("CLI still exists")
	}
	for _, path := range []string{data, state} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != "keep" {
			t.Fatalf("modified %s: %v", path, err)
		}
	}
}

func TestCanceledRemovalPreservesExecutable(t *testing.T) {
	target := filepath.Join(t.TempDir(), "mju-dataset")
	if err := os.WriteFile(target, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := remove(ctx, target); !errors.Is(err, context.Canceled) {
		t.Fatalf("error: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal("canceled uninstall removed executable")
	}
}
