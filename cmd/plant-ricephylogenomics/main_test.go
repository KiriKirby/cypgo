package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRicePhylogenomicsSources(t *testing.T) {
	root := filepath.Join("..", "..", "raw")
	rows, err := buildRows(filepath.Join(root, "plants-ricephylogenomics-RP450Protein.txt"), filepath.Join(root, "plants-ricephylogenomics-RP450cDNA.txt"), filepath.Join(root, "plants-ricephylogenomics-RP450Genomic.txt"))
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	if len(rows) != 341 {
		t.Fatalf("rows=%d", len(rows))
	}
	for _, tc := range []struct {
		i          int
		prefix, id string
	}{{0, "LOC_Os11g32240.1", "LOC_Os11g32240.1"}, {170, "LOC_Os04g", ""}, {340, "LOC_Os", ""}} {
		if !strings.HasPrefix(rows[tc.i][3], tc.prefix) || rows[tc.i][7] == "" || rows[tc.i][5] != proteinURL || !strings.Contains(rows[tc.i][6], cdnaURL) || !strings.Contains(rows[tc.i][6], genomicURL) {
			t.Fatalf("row %d=%q", tc.i, rows[tc.i])
		}
	}
}
