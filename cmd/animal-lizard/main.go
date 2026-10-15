package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const sourceFile = "animals-Anolis.carolinensis.xlsx"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Anolis.carolinensis.xlsx"
const sourceSheet = "Sheet2"

var keyRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type record struct {
	Row                                                                                                             int
	Model, UniProt, SpeciesCode, Range, SourceSpecies, Hit, Identity, Alignment, Comment, Symbol, Sequence, ID, Key string
	Status                                                                                                          []string
}

func main() {
	w, e := plantxlsx.ReadSheet(filepath.Join("raw", sourceFile), sourceSheet, false)
	if e != nil {
		panic(e)
	}
	rs, e := review(w.Rows)
	if e != nil {
		panic(e)
	}
	b, e := os.ReadFile(filepath.Join("raw", sourceFile))
	if e != nil {
		panic(e)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if e = writeCSV(rs, h); e != nil {
		panic(e)
	}
	if e = writeAudit(rs, w, h); e != nil {
		panic(e)
	}
	fmt.Printf("Anolis carolinensis: %d literal proteins\n", len(rs))
}
func review(rows map[int]map[string]string) ([]record, error) {
	want := map[string]string{"B": "UniProt ID", "C": "species code", "D": "range used", "G": "best hit ", "H": "%ID", "I": "aln. Length", "J": "comments"}
	for c, v := range want {
		if rows[1][c] != v {
			return nil, fmt.Errorf("header %s=%q", c, rows[1][c])
		}
	}
	var out []record
	for row := 2; row <= 169; row++ {
		v := rows[row]
		r := record{Row: row, Model: strings.TrimSpace(v["A"]), UniProt: strings.TrimSpace(v["B"]), SpeciesCode: strings.TrimSpace(v["C"]), Range: strings.TrimSpace(v["D"]), SourceSpecies: strings.TrimSpace(v["E"]), Hit: strings.TrimSpace(v["G"]), Identity: strings.TrimSpace(v["H"]), Alignment: strings.TrimSpace(v["I"]), Comment: strings.TrimSpace(v["J"]), Symbol: strings.TrimSpace(v["K"])}
		raw := strings.TrimSpace(v["L"])
		if r.Symbol == "" || raw == "" {
			return nil, fmt.Errorf("row %d blank", row)
		}
		var b strings.Builder
		lower := false
		for _, x := range raw {
			if unicode.IsSpace(x) {
				continue
			}
			if unicode.IsLower(x) {
				lower = true
			}
			u := unicode.ToUpper(x)
			if (u < 'A' || u > 'Z') && x != '*' && x != '-' {
				return nil, fmt.Errorf("row %d residue %q", row, x)
			}
			b.WriteRune(x)
		}
		r.Sequence = b.String()
		if strings.HasSuffix(r.Sequence, "*") {
			r.Sequence = strings.TrimSuffix(r.Sequence, "*")
			r.Status = append(r.Status, "terminal-stop-removed")
		}
		if strings.Contains(r.Sequence, "*") {
			r.Status = append(r.Status, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Status = append(r.Status, "ambiguous-X-retained")
		}
		if lower {
			r.Status = append(r.Status, "source-lowercase-retained")
		}
		if len(r.Sequence) < 350 {
			r.Status = append(r.Status, "short-source-fragment")
		}
		lc := strings.ToLower(r.Comment + " " + r.Symbol)
		if strings.Contains(lc, "partial") || strings.Contains(lc, "missing") || strings.Contains(lc, "pseudo") || strings.Contains(lc, "error") {
			r.Status = append(r.Status, "source-partial-or-pseudogene")
		}
		r.ID = r.UniProt
		if r.ID == "" {
			r.ID = r.Model
		}
		if r.ID == "" {
			r.ID = r.Symbol
		}
		r.Key = fmt.Sprintf("lizard:sheet2:row-%04d:%s", row, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(r.ID), "-"), "-"))
		out = append(out, r)
	}
	if len(out) != 168 {
		return nil, fmt.Errorf("count=%d", len(out))
	}
	return out, nil
}
func writeCSV(rs []record, h string) error {
	p := filepath.Join("sources", "reviewed", "animals", "lizard.csv")
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	f, e := os.Create(p)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "model", "uniprot", "species_code", "range", "source_species", "best_hit", "identity", "alignment_length", "comment", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Lizard updated Sheet2 row %d; model=%s; UniProt=%s; source species=%s; best hit=%s; comment=%s", r.Row, r.Model, r.UniProt, r.SourceSpecies, r.Hit, r.Comment)
		_ = w.Write([]string{"animals", "Anolis carolinensis", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, sourceFile, h, sourceSheet, strconv.Itoa(r.Row), r.Model, r.UniProt, r.SpeciesCode, r.Range, r.SourceSpecies, r.Hit, r.Identity, r.Alignment, r.Comment, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(rs []record, wb *plantxlsx.Workbook, h string) error {
	c := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Status {
			c[s]++
		}
	}
	ns := make([]string, 0, len(c))
	for n := range c {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Anolis carolinensis (green anole lizard)\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Workbook sheets: `Sheet1`, `Sheet2`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Literal proteins: `%d`\n- Review status: `complete`\n\n", sourceFile, sourceURL, h, sourceSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\n`Sheet1` is the older 95-name list. `Sheet2` is the expanded relationship table with 168 source rows, explicit CYP names in K and literal proteins in L; it updates several older names and adds many records. Therefore only Sheet2 rows 2–169 are published, preserving its duplicate/alternate rows rather than appending the superseded name list. Metadata columns remain provenance. Only terminal stops are removed; internal stops, X/x, lowercase, gaps, fragments and pseudogenes remain literal. No sequence is repaired or filled externally.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	p := filepath.Join("docs", "animals", "lizard.md")
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0644)
}
