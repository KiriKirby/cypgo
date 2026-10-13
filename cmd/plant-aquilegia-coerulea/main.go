package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceFile = "plants-Aquilegia.coerulea.xlsx"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Aquilegia.coerulea.xlsx"
const sourceHash = "4388faebc0e2602cebbc484814f84946c58cfa73cf3b489895bc686f881ed343"
const sheetName = "Sorted by CYP name"

type record struct {
	Row                                              int
	ID, BestHit, PercentID, Symbol, Sequence, Status string
}

var digitP = regexp.MustCompile(`(?i)(?:\dP|pseudo)$`)

func main() {
	input := flag.String("input", filepath.Join("raw", sourceFile), "downloaded Aquilegia workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "aquilegia-coerulea.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-aquilegia-coerulea.md"), "audit")
	flag.Parse()
	wb, err := plantxlsx.Read(*input, sheetName)
	if err != nil {
		panic(err)
	}
	rs, err := reviewRows(wb)
	if err != nil {
		panic(err)
	}
	hash, err := fileHash(*input)
	if err != nil {
		panic(err)
	}
	if hash != sourceHash {
		panic(fmt.Errorf("Aquilegia source hash changed: %s", hash))
	}
	if err := writeCSV(*out, rs, hash); err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, rs, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Aquilegia coerulea current workbook: %d assigned literal sequences accepted; 13 unassigned candidates excluded\n", len(rs))
}

func reviewRows(wb *plantxlsx.Workbook) ([]record, error) {
	if wb.Dimension != "A1:K565" || len(wb.HiddenRows) > 0 || len(wb.HiddenColumns) > 0 || len(wb.MergedCells) > 0 || wb.TableParts != 0 {
		return nil, fmt.Errorf("Aquilegia workbook structure changed")
	}
	h := wb.Rows[1]
	if h["G"] != "Seq. ID" || h["H"] != "best hit " || h["I"] != "%ID" || h["J"] != "CYP name" || h["K"] != "" {
		return nil, fmt.Errorf("Aquilegia headers changed")
	}
	rs := make([]record, 0, 551)
	for row := 2; row <= 552; row++ {
		v := wb.Rows[row]
		r := record{Row: row, ID: strings.TrimSpace(v["G"]), BestHit: strings.TrimSpace(v["H"]), PercentID: strings.TrimSpace(v["I"]), Symbol: strings.TrimSpace(v["J"]), Sequence: strings.TrimSpace(v["K"])}
		if r.ID == "" || r.BestHit == "" || r.PercentID == "" || r.Symbol == "" || r.Sequence == "" {
			return nil, fmt.Errorf("Aquilegia assigned row %d missing value", row)
		}
		if !strings.HasPrefix(r.BestHit, "CYP") || !strings.HasPrefix(r.Symbol, "CYP") {
			return nil, fmt.Errorf("Aquilegia assigned row %d non-CYP", row)
		}
		var s []string
		if digitP.MatchString(r.BestHit) {
			s = append(s, "source-best-hit-pseudogene-label")
		}
		if digitP.MatchString(r.Symbol) {
			s = append(s, "source-pseudogene-label")
		}
		if strings.HasSuffix(r.Sequence, "*") {
			r.Sequence = strings.TrimSuffix(r.Sequence, "*")
			s = append(s, "source-terminal-stop")
		}
		if strings.Contains(r.Sequence, "*") {
			s = append(s, "internal-stop")
		}
		if strings.Contains(r.Sequence, "-") {
			s = append(s, "source-gap")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			s = append(s, "ambiguous-X-or-x")
		}
		if strings.Contains(r.Sequence, "O") {
			s = append(s, "nonstandard-O")
		}
		if len(r.Sequence) < 350 {
			s = append(s, "short-sequence")
		}
		for _, aa := range r.Sequence {
			if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXYx*-", aa) {
				return nil, fmt.Errorf("Aquilegia row %d unexpected residue %q", row, aa)
			}
		}
		r.Status = strings.Join(s, ";")
		rs = append(rs, r)
	}
	for row := 553; row <= 565; row++ {
		v := wb.Rows[row]
		if strings.TrimSpace(v["G"]) == "" || strings.TrimSpace(v["K"]) == "" || strings.TrimSpace(v["H"]) != "" || strings.TrimSpace(v["I"]) != "" || strings.TrimSpace(v["J"]) != "" {
			return nil, fmt.Errorf("Aquilegia unassigned row %d changed", row)
		}
	}
	return rs, nil
}

func writeCSV(path string, rs []record, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Aquilegia current workbook row %d; best hit=%s; percent identity=%s; assigned CYP=%s; literal sequence column=K", r.Row, r.BestHit, r.PercentID, r.Symbol)
		_ = w.Write([]string{"plants", "Aquilegia coerulea", r.Symbol, r.ID, fmt.Sprintf("aquilegia-coerulea-current:row-%04d:%s", r.Row, r.ID), sourceURL, note, r.Sequence, sourceFile, hash, sheetName, fmt.Sprint(r.Row), r.Status})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, rs []record, wb *plantxlsx.Workbook, hash string) error {
	counts := map[string]int{}
	seqs := map[string]int{}
	for _, r := range rs {
		seqs[r.Sequence]++
		for _, s := range strings.Split(r.Status, ";") {
			if s != "" {
				counts[s]++
			}
		}
	}
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	dupes := 0
	for _, n := range seqs {
		if n > 1 {
			dupes++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Plant resource review: Aquilegia coerulea current workbook\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Sheet: `%s`, `%s`; no hidden rows/columns, merges, or tables\n- Accepted assigned rows: `%d` (`2-552`)\n- Excluded unassigned candidates: `13` (`553-565`)\n- Duplicate literal-sequence groups retained: `%d`\n- Review status: `complete`\n\nRows 2-552 have explicit values in G/J/K for sequence ID, assigned CYP, and literal protein. Rows 553-565 retain IDs and sequence-like text but H/I/J are all empty, so they remain excluded rather than receiving guessed assignments. Literal O, X/x, gaps, internal stops and short proteins are retained; only one terminal stop is removed.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n", sourceFile, sourceURL, hash, sheetName, wb.Dimension, len(rs), dupes)
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	b.WriteString("\n## Representative records\n\n| Row | ID | CYP | Length | Status |\n|---:|---|---|---:|---|\n")
	for _, i := range []int{0, len(rs) / 2, len(rs) - 1} {
		r := rs[i]
		fmt.Fprintf(&b, "| %d | %s | %s | %d | %s |\n", r.Row, r.ID, r.Symbol, len(r.Sequence), r.Status)
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
func fileHash(path string) (string, error) {
	d, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(d)), nil
}
