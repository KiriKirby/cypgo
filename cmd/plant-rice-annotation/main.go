// Resource-specific importer for David Nelson's legacy rice annotation workbook.
// The workbook has one explicit table (Sheet1, header row 6, records rows 7-418).
// It is intentionally kept separate from the rice FASTA resource: annotation
// rows remain available even when the workbook has no literal protein column.
package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shakinm/xlsReader/xls"
)

const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/rice_annotation.P450.xls"
const expectedSHA256 = "3532c8470d07188dc00d5135caa7772af80429d9b5ea471a2bb3ac25214a91c1"

func cell(sheet *xls.Sheet, row, col int) string {
	r, err := sheet.GetRow(row)
	if err != nil {
		return ""
	}
	c, err := r.GetCol(col)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(c.GetString(), "\r", " "), "\n", " "))
}

func buildRows(input string) ([][]string, string, error) {
	data, err := os.ReadFile(input)
	if err != nil {
		return nil, "", err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if hash != expectedSHA256 {
		return nil, hash, fmt.Errorf("rice annotation workbook SHA-256 changed: got %s, want %s", hash, expectedSHA256)
	}
	book, err := xls.OpenFile(input)
	if err != nil {
		return nil, hash, err
	}
	if book.GetNumberSheets() != 3 {
		return nil, hash, fmt.Errorf("rice annotation workbook: expected 3 worksheets (one data sheet plus two empty sheets), got %d", book.GetNumberSheets())
	}
	sheet, err := book.GetSheet(0)
	if err != nil {
		return nil, hash, err
	}
	if sheet.GetName() != "Sheet1" {
		return nil, hash, fmt.Errorf("rice annotation workbook: unexpected sheet %q", sheet.GetName())
	}
	for i, want := range []string{"Sheet2", "Sheet3"} {
		empty, err := book.GetSheet(i + 1)
		if err != nil || empty.GetName() != want || cell(empty, 0, 0) != "" {
			return nil, hash, fmt.Errorf("rice annotation workbook: worksheet %d must be empty %s", i+2, want)
		}
	}
	// Validate the exact reviewed header layout before accepting any rows.
	expected := []string{"Locus", "Gene Name", "Gene Description", "Alt Gene Name", "GenBank Acc Genomic", "GenBank Acc cDNA", "GenBank Acc Protein", "Mutant", "Criteria", "Source", "Comment"}
	for i, want := range expected {
		if got := cell(sheet, 5, i); got != want {
			return nil, hash, fmt.Errorf("rice annotation header column %d: got %q, want %q", i+1, got, want)
		}
	}
	rows := make([][]string, 0, 412)
	for row := 6; row <= 417; row++ { // Excel rows 7-418; xlsReader indexes from zero.
		gene := strings.TrimSpace(cell(sheet, row, 1))
		if gene == "" {
			return nil, hash, fmt.Errorf("rice annotation row %d has empty Gene Name", row+1)
		}
		desc := cell(sheet, row, 2)
		alt, genomic, cdna, protein := cell(sheet, row, 3), cell(sheet, row, 4), cell(sheet, row, 5), cell(sheet, row, 6)
		mutant, criteria, source, comment := cell(sheet, row, 7), cell(sheet, row, 8), cell(sheet, row, 9), cell(sheet, row, 10)
		note := fmt.Sprintf("rice_annotation.P450.xls Sheet1 row %d; Gene Description=%s; Alt Gene Name=%s; GenBank Acc Genomic=%s; GenBank Acc cDNA=%s; GenBank Acc Protein=%s; Mutant=%s; Criteria=%s; Source=%s; Comment=%s", row+1, desc, alt, genomic, cdna, protein, mutant, criteria, source, comment)
		rows = append(rows, []string{"plants", "Oryza sativa", gene, gene, fmt.Sprintf("rice-annotation:sheet1:row-%04d", row+1), sourceURL, note, "", filepath.Base(input), hash, fmt.Sprintf("Sheet1!row-%d", row+1), "annotation-only;no-sequence"})
	}
	if len(rows) != 412 {
		return nil, hash, fmt.Errorf("rice annotation rows: got %d, want 412", len(rows))
	}
	return rows, hash, nil
}

func main() {
	in := flag.String("input", "raw/plants-rice_annotation.P450.xls", "source workbook")
	out := flag.String("out", "sources/reviewed/plants/rice-annotation.csv", "reviewed CSV")
	flag.Parse()
	rows, _, err := buildRows(*in)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "review_status"})
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("rice annotation: %d records, 0 sequences\n", len(rows))
}
