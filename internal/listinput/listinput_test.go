package listinput

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestLoadFileReadsOnlySelectedFileAndDeduplicatesFormats(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.csv"), []byte("\ufeffclassification_number,ignored\nGC-001,x\n GC-002 ,y\nGC-001,z\n,x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	book := excelize.NewFile()
	if err := book.SetSheetRow("Sheet1", "A1", &[]interface{}{"ignored", "classification_number"}); err != nil {
		t.Fatal(err)
	}
	if err := book.SetSheetRow("Sheet1", "A2", &[]interface{}{"x", " GC-003 "}); err != nil {
		t.Fatal(err)
	}
	if err := book.SaveAs(filepath.Join(dir, "b.xlsx")); err != nil {
		t.Fatal(err)
	}
	if err := book.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "c.json"), []byte(`{"classification_numbers":["GC-002"," GC-004 ",""]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "ignored.csv"), []byte("classification_number\nGC-999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// A malformed sibling must not affect loading the selected file.
	if err := os.WriteFile(filepath.Join(dir, "bad.csv"), []byte("wrong_header\nGC-999\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		want []string
	}{
		{"a.csv", []string{"GC-001", "GC-002"}},
		{"b.xlsx", []string{"GC-003"}},
		{"c.json", []string{"GC-002", "GC-004"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LoadFile(filepath.Join(dir, tc.name))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got numbers=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadFileRejectsMalformedList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.csv"), []byte("wrong_header\nGC-001\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFile(filepath.Join(dir, "bad.csv"))
	if err == nil || !strings.Contains(err.Error(), "classification_number column is missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFileRejectsInvalidPathsAndContents(t *testing.T) {
	dir := t.TempDir()
	for _, tc := range []struct {
		name      string
		content   string
		wantError string
	}{
		{"unsupported.txt", "classification_number\nGC-001\n", "must be CSV, XLSX, or JSON"},
		{"empty.json", `{"classification_numbers":["", " "]}`, "no classification numbers"},
		{"missing-array.json", `{}`, "array is missing"},
		{"malformed.json", `{`, "invalid list file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name)
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadFile(path); err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
	if _, err := LoadFile(dir); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("directory error: %v", err)
	}
	if _, err := LoadFile(filepath.Join(dir, "missing.csv")); err == nil {
		t.Fatal("expected missing file error")
	}
	path := filepath.Join(dir, "oversized.csv")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	err = file.Truncate(maxFileSize + 1)
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("create oversized file: %v, %v", err, closeErr)
	}
	if _, err := LoadFile(path); err == nil || !strings.Contains(err.Error(), "50 MiB") {
		t.Fatalf("oversized file error: %v", err)
	}
}
