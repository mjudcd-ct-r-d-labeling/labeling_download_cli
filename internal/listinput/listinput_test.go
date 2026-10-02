package listinput

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestLoadDirectoryReadsAndDeduplicatesFormats(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.csv"), []byte("\ufeffclassification_number,ignored\nGC-001,x\nGC-002,y\n"), 0600); err != nil {
		t.Fatal(err)
	}
	book := excelize.NewFile()
	if err := book.SetSheetRow("Sheet1", "A1", &[]interface{}{"ignored", "classification_number"}); err != nil {
		t.Fatal(err)
	}
	if err := book.SetSheetRow("Sheet1", "A2", &[]interface{}{"x", "GC-003"}); err != nil {
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
	got, count, err := LoadDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 || !reflect.DeepEqual(got, []string{"GC-001", "GC-002", "GC-003", "GC-004"}) {
		t.Fatalf("got count=%d numbers=%v", count, got)
	}
}

func TestLoadDirectoryRejectsMalformedList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.csv"), []byte("wrong_header\nGC-001\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, _, err := LoadDirectory(dir)
	if err == nil || !strings.Contains(err.Error(), "classification_number column is missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}
