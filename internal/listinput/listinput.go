// Package listinput reads classification numbers from a directory of list files.
package listinput

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

const maxFileSize = 50 << 20

// LoadDirectory reads supported files in the specified directory only. Each
// number is returned once, in first-seen order across alphabetically sorted files.
func LoadDirectory(dir string) ([]string, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot read list directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	seen := make(map[string]bool)
	numbers := make([]string, 0)
	fileCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".csv" && ext != ".xlsx" && ext != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Size() > maxFileSize {
			return nil, 0, fmt.Errorf("invalid list file: %s", entry.Name())
		}
		var values []string
		switch ext {
		case ".csv":
			values, err = readCSV(path)
		case ".xlsx":
			values, err = readXLSX(path)
		case ".json":
			values, err = readJSON(path)
		}
		if err != nil {
			return nil, 0, fmt.Errorf("invalid list file %s: %w", entry.Name(), err)
		}
		fileCount++
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value != "" && !seen[value] {
				seen[value] = true
				numbers = append(numbers, value)
			}
		}
	}
	if fileCount == 0 {
		return nil, 0, fmt.Errorf("no CSV, XLSX, or JSON list files found")
	}
	if len(numbers) == 0 {
		return nil, fileCount, fmt.Errorf("list files contain no classification numbers")
	}
	return numbers, fileCount, nil
}

func readCSV(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("CSV must be UTF-8")
	}
	text := strings.TrimPrefix(string(data), "\ufeff")
	r := csv.NewReader(strings.NewReader(text))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	return fromRows(rows)
}

func readXLSX(path string) ([]string, error) {
	book, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer book.Close()
	sheet := book.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("first worksheet is missing")
	}
	rows, err := book.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	return fromRows(rows)
}

func fromRows(rows [][]string) ([]string, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("header row is missing")
	}
	column := -1
	for i, value := range rows[0] {
		if strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")) == "classification_number" {
			column = i
			break
		}
	}
	if column < 0 {
		return nil, fmt.Errorf("classification_number column is missing")
	}
	values := make([]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if column < len(row) {
			values = append(values, row[column])
		}
	}
	return values, nil
}

func readJSON(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var content struct {
		ClassificationNumbers []string `json:"classification_numbers"`
	}
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}
	if content.ClassificationNumbers == nil {
		return nil, fmt.Errorf("classification_numbers array is missing")
	}
	return content.ClassificationNumbers, nil
}
