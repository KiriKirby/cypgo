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
	"strconv"
	"strings"
	"unicode"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceFile = "animals-Tasmanian.devil.xlsx"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Tasmanian.devil.xlsx"
const sourceSheet = "Tasmanian.blastout.txt"

var keyRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type record struct {
	Row                                                                    int
	SourceID, Hit, Identity, Alignment, Comment, Symbol, Sequence, ID, Key string
	Status                                                                 []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "tasmanian-devil.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "tasmanian-devil.md"), "")
	flag.Parse()
	names, e := plantxlsx.SheetNames(*in)
	if e != nil {
		panic(e)
	}
	if strings.Join(names, "|") != sourceSheet {
		panic(fmt.Errorf("sheet layout changed: %q", names))
	}
	wb, e := plantxlsx.ReadSheet(*in, sourceSheet, false)
	if e != nil {
		panic(e)
	}
	rs, e := review(wb.Rows)
	if e != nil {
		panic(e)
	}
	b, e := os.ReadFile(*in)
	if e != nil {
		panic(e)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if e = writeCSV(*out, rs, h); e != nil {
		panic(e)
	}
	if e = writeAudit(*audit, rs, wb, h); e != nil {
		panic(e)
	}
	fmt.Printf("Sarcophilus harrisii: %d literal proteins\n", len(rs))
}

func review(rows map[int]map[string]string) ([]record, error) {
	want := map[string]string{"A": "seq ID (Gotoh, or accession)", "B": "best hit in mammals", "C": "%ID", "D": "Aln.length", "E": "comment", "F": "CYP name"}
	for c, v := range want {
		if rows[1][c] != v {
			return nil, fmt.Errorf("header %s=%q want %q", c, rows[1][c], v)
		}
	}
	var out []record
	for row := 2; row <= 92; row++ {
		v := rows[row]
		r := record{Row: row, SourceID: strings.TrimSpace(v["A"]), Hit: strings.TrimSpace(v["B"]), Identity: strings.TrimSpace(v["C"]), Alignment: strings.TrimSpace(v["D"]), Comment: strings.TrimSpace(v["E"]), Symbol: strings.TrimSpace(v["F"])}
		raw := strings.TrimSpace(v["G"])
		if r.Symbol == "" || raw == "" {
			return nil, fmt.Errorf("row %d missing CYP/sequence", row)
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
		if strings.Contains(lc, "partial") || strings.Contains(lc, "missing") || strings.Contains(lc, "fragment") || strings.Contains(lc, "pseudogene") {
			r.Status = append(r.Status, "source-partial-or-pseudogene")
		}
		r.ID = r.SourceID
		if r.ID == "" {
			r.ID = r.Symbol
		}
		r.Key = fmt.Sprintf("tasmanian-devil:sheet-tasmanian-blastout:row-%04d:%s", row, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(r.ID), "-"), "-"))
		out = append(out, r)
	}
	if len(out) != 91 {
		return nil, fmt.Errorf("count=%d want 91", len(out))
	}
	return out, nil
}

func writeCSV(path string, rs []record, hash string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "source_id", "best_hit", "identity", "alignment_length", "comment", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Tasmanian devil workbook row %d; source ID=%s; best mammal hit=%s; identity=%s; alignment=%s; comment=%s", r.Row, r.SourceID, r.Hit, r.Identity, r.Alignment, r.Comment)
		_ = w.Write([]string{"animals", "Sarcophilus harrisii", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, sourceFile, hash, sourceSheet, strconv.Itoa(r.Row), r.SourceID, r.Hit, r.Identity, r.Alignment, r.Comment, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, rs []record, wb *plantxlsx.Workbook, hash string) error {
	counts := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Status {
			counts[s]++
		}
	}
	ns := make([]string, 0, len(counts))
	for n := range counts {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Sarcophilus harrisii (Tasmanian devil)\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Workbook sheets: `%s`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", sourceFile, sourceURL, hash, sourceSheet, sourceSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe only worksheet is a final seven-column relationship table. Rows 2–92 each have an explicit CYP name and literal protein sequence in column G, so all 91 source rows are retained in worksheet order. Column A is used as the ID and the CYP name is the fallback. Best-hit, identity, alignment length and comments remain provenance only. Terminal stops are removed; internal stops, X/x, lowercase, gaps, fragments, missing regions and pseudogene annotations remain literal. No sequence is translated, repaired, inferred or fetched externally.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
