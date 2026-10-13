package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const sourceFile = "plants-Triticum.aestivum.xlsx"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Triticum.aestivum.xlsx"
const sourceHash = "ae064dab13fd31ea37999630a6704199a81838e481485749664a2de4b7c451e3"

type record struct {
	Row                                int
	ID, Hit, Percent, Sequence, Status string
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "downloaded Triticum workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "triticum-aestivum.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-triticum-aestivum.md"), "audit")
	flag.Parse()
	w, e := plantxlsx.Read(*in, "Sorted by CYP name")
	if e != nil {
		panic(e)
	}
	a, x, e := review(w)
	if e != nil {
		panic(e)
	}
	h, e := fileHash(*in)
	if e != nil {
		panic(e)
	}
	if h != sourceHash {
		panic(fmt.Errorf("Triticum hash=%s", h))
	}
	if e = writeCSV(*out, a, h); e != nil {
		panic(e)
	}
	if e = writeAudit(*audit, a, x, w, h); e != nil {
		panic(e)
	}
	fmt.Printf("Triticum aestivum: %d CYP-assigned rows accepted; %d unassigned/bacterial rows excluded\n", len(a), len(x))
}
func review(w *plantxlsx.Workbook) ([]record, []record, error) {
	if w.Dimension != "A1:J1701" || len(w.HiddenRows) > 0 || len(w.HiddenColumns) > 0 || len(w.MergedCells) > 0 || w.TableParts > 0 {
		return nil, nil, fmt.Errorf("Triticum layout changed")
	}
	want := map[string]string{"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "Seq. ID", "H": "best hit", "I": "%ID", "J": ""}
	for c, v := range want {
		if w.Rows[1][c] != v {
			return nil, nil, fmt.Errorf("header %s=%q", c, w.Rows[1][c])
		}
	}
	var a, x []record
	for i := 2; i <= 1701; i++ {
		v := w.Rows[i]
		g := strings.TrimSpace(v["G"])
		fs := strings.Fields(g)
		if len(fs) == 0 {
			return nil, nil, fmt.Errorf("row %d no ID", i)
		}
		r := record{Row: i, ID: fs[0], Hit: strings.TrimSpace(v["H"]), Percent: strings.TrimSpace(v["I"]), Sequence: strings.TrimSpace(v["J"])}
		if i <= 1476 {
			if !strings.HasPrefix(r.Hit, "CYP") || r.Percent == "" || r.Sequence == "" || !strings.HasPrefix(strings.TrimSpace(v["A"]), r.ID+" ") {
				return nil, nil, fmt.Errorf("assigned row %d changed", i)
			}
			st := []string{}
			if strings.HasSuffix(r.Hit, "P") {
				st = append(st, "source-best-hit-pseudogene-label")
			}
			if strings.ContainsAny(r.Sequence, "Xx") {
				st = append(st, "ambiguous-X-or-x")
			}
			if strings.Contains(r.Sequence, "O") {
				st = append(st, "nonstandard-O")
			}
			if strings.Contains(r.Sequence, "-") {
				st = append(st, "source-gap")
			}
			if len(r.Sequence) < 350 {
				st = append(st, "short-sequence")
			}
			for _, aa := range r.Sequence {
				if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXYx-", aa) {
					return nil, nil, fmt.Errorf("row %d residue %q", i, aa)
				}
			}
			r.Status = strings.Join(st, ";")
			a = append(a, r)
		} else {
			if i == 1701 {
				if !strings.Contains(strings.ToLower(g), "bacterial contamination") || r.Hit != "Bacillus cellulosilyticus CP002394.1" {
					return nil, nil, fmt.Errorf("bacterial row changed")
				}
				r.Status = "excluded: explicit bacterial contamination"
			} else {
				if r.Hit != "" || r.Percent != "" {
					return nil, nil, fmt.Errorf("unassigned row %d gained assignment", i)
				}
				r.Status = "excluded: no CYP best hit or percent identity"
			}
			x = append(x, r)
		}
	}
	if len(a) != 1475 || len(x) != 225 {
		return nil, nil, fmt.Errorf("counts %d/%d", len(a), len(x))
	}
	return a, x, nil
}
func writeCSV(path string, a []record, h string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "review_status"})
	for _, r := range a {
		note := fmt.Sprintf("Triticum workbook row %d; source H best hit=%s; source I %%ID=%s; literal sequence column=J", r.Row, r.Hit, r.Percent)
		_ = w.Write([]string{"plants", "Triticum aestivum", r.Hit, r.ID, fmt.Sprintf("triticum-aestivum:row-%04d:%s", r.Row, r.ID), sourceURL, note, r.Sequence, sourceFile, h, "Sorted by CYP name", strconv.Itoa(r.Row), r.Status})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(path string, a, x []record, w *plantxlsx.Workbook, h string) error {
	cs := map[string]int{}
	seqs := map[string]int{}
	for _, r := range a {
		seqs[r.Sequence]++
		for _, s := range strings.Split(r.Status, ";") {
			if s != "" {
				cs[s]++
			}
		}
	}
	names := []string{}
	for n := range cs {
		names = append(names, n)
	}
	sort.Strings(names)
	dg := 0
	for _, n := range seqs {
		if n > 1 {
			dg++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Plant resource review: Triticum aestivum\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Sheet: `Sorted by CYP name`\n- Used range: `%s`\n- Accepted CYP-assigned rows: `%d`\n- Excluded unassigned rows: `224`\n- Excluded bacterial rows: `1`\n- Duplicate sequence groups retained: `%d`\n- Review status: `complete`\n\nThis workbook has a distinct A:J layout: G is `Seq. ID`, H is `best hit`, I is `%%ID`, and headerless J is the literal protein. It has no separate assigned-name column, so the source H value is retained as the CYP symbol without inventing a different family. Rows 2-1476 are the continuous region with CYP best hits and percentages. Rows 1477-1700 have neither, and are excluded even where J contains an out-of-frame translation; row 1701 explicitly labels `Bacillus cellulosilyticus` bacterial contamination and is excluded. X/O, gaps, short proteins and duplicate sequences remain literal and row-distinct.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n", sourceFile, sourceURL, h, w.Dimension, len(a), dg)
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, cs[n])
	}
	b.WriteString("\n## Representative records\n\n| Row | ID | CYP best hit | Length |\n|---:|---|---|---:|\n")
	for _, i := range []int{0, len(a) / 2, len(a) - 1} {
		r := a[i]
		fmt.Fprintf(&b, "| %d | %s | %s | %d |\n", r.Row, r.ID, r.Hit, len(r.Sequence))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
func fileHash(p string) (string, error) {
	d, e := os.ReadFile(p)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(d)), nil
}
