package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
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

const sourceFile = "animals-Kiwi.xlsx"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Kiwi.xlsx"
const sourceSheet = "Kiwi.blastout.txt"

var keyRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type record struct {
	Row                                                                                        int
	GI, Accession, Hit, Identity, Alignment, SourceSpecies, Symbol, Comment, Sequence, ID, Key string
	Status                                                                                     []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "kiwi.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "kiwi.md"), "")
	flag.Parse()
	names, e := plantxlsx.SheetNames(*in)
	if e != nil {
		panic(e)
	}
	if strings.Join(names, "|") != sourceSheet {
		panic(fmt.Errorf("sheets changed: %q", names))
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
	fmt.Printf("Apteryx australis mantelli: %d literal proteins\n", len(rs))
}
func review(rows map[int]map[string]string) ([]record, error) {
	if rows[1]["G"] != "CYP name" || rows[1]["H"] != "comments" {
		return nil, fmt.Errorf("headers changed")
	}
	var out []record
	for row := 2; row <= 66; row++ {
		v := rows[row]
		r := record{Row: row, GI: strings.TrimSpace(v["A"]), Accession: strings.TrimSpace(v["B"]), Hit: strings.TrimSpace(v["C"]), Identity: strings.TrimSpace(v["D"]), Alignment: strings.TrimSpace(v["E"]), SourceSpecies: strings.TrimSpace(v["F"]), Symbol: strings.TrimSpace(v["G"]), Comment: strings.TrimSpace(v["H"])}
		raw := strings.TrimSpace(v["I"])
		// Row 5 is the sole genomic-gap reconstruction and intentionally has
		// a blank species cell; all non-blank species cells use the kiwi label.
		if r.Symbol == "" || raw == "" || (r.SourceSpecies != "" && r.SourceSpecies != "Apteryx australis mantelli") {
			return nil, fmt.Errorf("row %d missing reviewed values", row)
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
		if strings.Contains(lc, "partial") || strings.Contains(lc, "missing") || strings.Contains(lc, "fragment") || strings.Contains(lc, "pseudo") {
			r.Status = append(r.Status, "source-partial-or-pseudogene")
		}
		r.ID = r.Accession
		if r.ID == "" {
			r.ID = r.GI
		}
		if r.ID == "" {
			r.ID = r.Symbol
		}
		r.Key = fmt.Sprintf("kiwi:sheet-kiwi-blastout:row-%04d:%s", row, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(r.ID), "-"), "-"))
		out = append(out, r)
	}
	if len(out) != 65 {
		return nil, fmt.Errorf("count=%d", len(out))
	}
	return out, nil
}
func writeCSV(p string, rs []record, h string) error {
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	f, e := os.Create(p)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "gi", "accession", "best_hit", "identity", "alignment_length", "source_species", "comment", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Kiwi workbook row %d; GI=%s; accession=%s; best hit=%s; identity=%s; alignment=%s; comment=%s", r.Row, r.GI, r.Accession, r.Hit, r.Identity, r.Alignment, r.Comment)
		_ = w.Write([]string{"animals", "Apteryx australis mantelli", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, sourceFile, h, sourceSheet, strconv.Itoa(r.Row), r.GI, r.Accession, r.Hit, r.Identity, r.Alignment, r.SourceSpecies, r.Comment, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p string, rs []record, wb *plantxlsx.Workbook, h string) error {
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
	fmt.Fprintf(&b, "# Animal resource review: Apteryx australis mantelli (kiwi)\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Authoritative worksheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", sourceFile, sourceURL, h, sourceSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nRows 2–66 of the sole worksheet form the curated kiwi table: explicit species, CYP name and literal protein in column I. Rows 67–71 are workbook notes/count summaries and are excluded. All 65 curated rows remain source-distinct, including duplicate names, partials, missing regions and pseudogenes. GI/accession, hit, identity, alignment and comment columns remain provenance. Only terminal stops are removed; internal stops, X/x, case and gaps remain literal. No external sequence is used.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0644)
}
