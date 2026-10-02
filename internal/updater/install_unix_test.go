//go:build !windows

package updater

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallReplacesOnlyExecutable(t *testing.T) {
	dir := t.TempDir()
	source, target, dataset := filepath.Join(dir, "download"), filepath.Join(dir, "mju-dataset"), filepath.Join(dir, "dataset.jsonl")
	for path, value := range map[string]string{source: "new executable", target: "old executable", dataset: "preserve dataset"} {
		if err := os.WriteFile(path, []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := install(context.Background(), source, target, "2026.10.02.12"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(target)
	if string(got) != "new executable" {
		t.Fatal("update not installed")
	}
	got, _ = os.ReadFile(dataset)
	if string(got) != "preserve dataset" {
		t.Fatal("dataset modified")
	}
	fi, _ := os.Stat(target)
	if fi.Mode().Perm() != 0755 {
		t.Fatal("updated CLI is not executable")
	}
	paths, _ := filepath.Glob(filepath.Join(dir, ".mju-dataset-update-*"))
	if len(paths) != 0 {
		t.Fatal("staging files leaked")
	}
}
