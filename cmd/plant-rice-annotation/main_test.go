package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRiceAnnotationWorkbook(t *testing.T) {
	path := filepath.Join("..", "..", "raw", "plants-rice_annotation.P450.xls")
	if _, err := os.Stat(path); err != nil {
		t.Skip("review source workbook is not present")
	}
	rows, hash, err := buildRows(path)
	if err != nil {
		t.Fatal(err)
	}
	if hash != expectedSHA256 || len(rows) != 412 {
		t.Fatalf("hash=%s rows=%d", hash, len(rows))
	}
	for _, tc := range []struct {
		index         int
		symbol, block string
	}{
		{0, "CYP51G1", "Sheet1!row-7"},
		{205, "CYP81L4", "Sheet1!row-212"},
		{411, "CYP735A4", "Sheet1!row-418"},
	} {
		row := rows[tc.index]
		if row[2] != tc.symbol || row[10] != tc.block || row[7] != "" || row[11] != "annotation-only;no-sequence" {
			t.Fatalf("row %d=%q", tc.index, row)
		}
		if !strings.Contains(row[6], "Gene Description=") || row[5] != sourceURL {
			t.Fatalf("row %d lacks provenance: %q", tc.index, row)
		}
	}
}
