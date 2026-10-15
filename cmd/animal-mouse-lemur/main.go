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

const mouseLemurFile = "animals-Mouse.lemur.P450s.UCSC.xlsx"
const mouseLemurURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Mouse.lemur.P450s.UCSC.xlsx"
const mouseLemurSheet = "Sorted by CYP family"

var mouseLemurKeyRE = regexp.MustCompile(`[^A-Za-z0-9._|-]+`)

type mouseLemurRecord struct {
	Row                                                                                                      int
	UCSCTag, Scaffold, Range, GenbankHit, Status, Comment, GotohID, Ensembl, Symbol, Sequence, ID, RecordKey string
	Statuses                                                                                                 []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", mouseLemurFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "mouse-lemur.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "mouse-lemur.md"), "")
	flag.Parse()
	names, e := plantxlsx.SheetNames(*in)
	if e != nil {
		panic(e)
	}
	if strings.Join(names, "\x00") != strings.Join([]string{"Sorted by location, then joined", mouseLemurSheet}, "\x00") {
		panic(fmt.Errorf("mouse lemur sheets changed: %q", names))
	}
	wb, e := plantxlsx.ReadSheet(*in, mouseLemurSheet, false)
	if e != nil {
		panic(e)
	}
	rs, e := reviewMouseLemurRows(wb.Rows)
	if e != nil {
		panic(e)
	}
	d, e := os.ReadFile(*in)
	if e != nil {
		panic(e)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(d))
	if e = writeMouseLemurCSV(*out, rs, h); e != nil {
		panic(e)
	}
	if e = writeMouseLemurAudit(*audit, rs, wb, h); e != nil {
		panic(e)
	}
	fmt.Printf("Microcebus murinus: %d reviewed literal protein records\n", len(rs))
}

func reviewMouseLemurRows(rows map[int]map[string]string) ([]mouseLemurRecord, error) {
	h := rows[2]
	want := map[string]string{"A": "UCSC tag", "B": "scaffold", "C": "range", "D": "Genbank hit", "K": "status", "L": "comments", "M": "Gotoh ID", "N": "Ensembl sequences", "O": "seq"}
	for c, v := range want {
		if h[c] != v {
			return nil, fmt.Errorf("mouse lemur header %s got %q want %q", c, h[c], v)
		}
	}
	var out []mouseLemurRecord
	for row := 3; row <= 180; row++ {
		v := rows[row]
		seq := strings.TrimSpace(v["O"])
		if seq == "" {
			continue
		}
		r := mouseLemurRecord{Row: row, UCSCTag: strings.TrimSpace(v["A"]), Scaffold: strings.TrimSpace(v["B"]), Range: strings.TrimSpace(v["C"]), GenbankHit: strings.TrimSpace(v["D"]), Status: strings.TrimSpace(v["K"]), Comment: strings.TrimSpace(v["L"]), GotohID: strings.TrimSpace(v["M"]), Ensembl: strings.TrimSpace(v["N"]), Sequence: seq}
		r.Symbol = strings.TrimSpace(r.UCSCTag)
		if r.Symbol == "" {
			for _, candidate := range []string{r.Comment, r.Status} {
				if strings.HasPrefix(strings.ToUpper(candidate), "CYP") {
					r.Symbol = candidate
					break
				}
			}
		}
		r.Sequence, r.Statuses = normalizeMouseLemurSequence(r.Sequence)
		for _, x := range r.Sequence {
			u := unicode.ToUpper(x)
			if (u < 'A' || u > 'Z') && x != '*' && x != '-' {
				return nil, fmt.Errorf("mouse lemur row %d unexpected %q", row, x)
			}
		}
		joined := strings.ToLower(r.Status + " " + r.Comment)
		if strings.Contains(r.Sequence, "*") {
			r.Statuses = append(r.Statuses, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Statuses = append(r.Statuses, "ambiguous-X-retained")
		}
		if strings.Contains(r.Sequence, "-") {
			r.Statuses = append(r.Statuses, "source-gap-retained")
		}
		if len(r.Sequence) < 350 {
			r.Statuses = append(r.Statuses, "short-source-fragment")
		}
		if strings.Contains(joined, "fragment") || strings.Contains(joined, "partial") || strings.Contains(joined, "missing") || strings.Contains(joined, "exon") {
			r.Statuses = append(r.Statuses, "source-partial-or-missing-region")
		}
		if r.Symbol == "" {
			r.Statuses = append(r.Statuses, "source-cyp-name-missing")
		}
		r.ID = firstMouseLemur(r.Ensembl, r.GotohID, mouseLemurLocation(r), r.Symbol)
		r.RecordKey = fmt.Sprintf("mouse-lemur:sheet-sorted-by-cyp-family:row-%04d:%s", row, mouseLemurKey(r.ID))
		out = append(out, r)
	}
	if len(out) != 177 {
		return nil, fmt.Errorf("mouse lemur records=%d want 177", len(out))
	}
	return out, nil
}
func normalizeMouseLemurSequence(s string) (string, []string) {
	var b strings.Builder
	lo, ws := false, false
	for _, r := range s {
		if unicode.IsSpace(r) {
			ws = true
			continue
		}
		if unicode.IsLower(r) {
			lo = true
		}
		b.WriteRune(r)
	}
	s = b.String()
	var st []string
	if ws {
		st = append(st, "source-whitespace-joined")
	}
	if lo {
		st = append(st, "source-lowercase-retained")
	}
	if strings.HasSuffix(s, "*") {
		s = strings.TrimSuffix(s, "*")
		st = append(st, "terminal-stop-removed")
	}
	return s, st
}
func mouseLemurLocation(r mouseLemurRecord) string {
	if r.Scaffold != "" && r.Range != "" {
		return "scaffold" + r.Scaffold + ":" + r.Range
	}
	return ""
}
func firstMouseLemur(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func mouseLemurKey(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	return strings.Trim(mouseLemurKeyRE.ReplaceAllString(v, "-"), "-")
}
func writeMouseLemurCSV(path string, rs []mouseLemurRecord, h string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "ucsc_tag", "scaffold", "range", "genbank_hit", "source_status", "source_comment", "gotoh_id", "ensembl", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Mouse lemur final CYP-family sheet row %d; UCSC tag=%s; scaffold=%s; range=%s; GenBank hit=%s; status=%s; comment=%s; Gotoh ID=%s; Ensembl=%s", r.Row, r.UCSCTag, r.Scaffold, r.Range, r.GenbankHit, r.Status, r.Comment, r.GotohID, r.Ensembl)
		_ = w.Write([]string{"animals", "Microcebus murinus", r.Symbol, r.ID, r.RecordKey, mouseLemurURL, note, r.Sequence, mouseLemurFile, h, mouseLemurSheet, strconv.Itoa(r.Row), r.UCSCTag, r.Scaffold, r.Range, r.GenbankHit, r.Status, r.Comment, r.GotohID, r.Ensembl, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeMouseLemurAudit(path string, rs []mouseLemurRecord, wb *plantxlsx.Workbook, h string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	c := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Statuses {
			c[s]++
		}
	}
	ns := make([]string, 0, len(c))
	for n := range c {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Microcebus murinus (grey mouse-lemur)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Literal protein sequences: `%d`\n- Review status: `complete`\n\n", mouseLemurFile, mouseLemurURL, h, mouseLemurSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe first worksheet is the UCSC/Gotoh search and joining workspace. Its footer says more than 2,100 browser hits were reduced to 169 and compared with 52 Gotoh sequences, with many still needing assembly. The second worksheet is explicitly the improved CYP-family view and is the only published table. Rows 3–180 contain 177 literal sequence cells. Seven have an empty UCSC-tag/CYP-name cell. A same-row comment/status beginning with CYP is used only when explicit; otherwise the symbol remains empty and Ensembl, Gotoh, or location supplies the ID. No CYP name is inferred from sequence similarity.\n\nAll source-distinct fragments and repeated labels remain separate. Lowercase x, gaps, internal stops, Ensembl-inferred/hybrid warnings, incomplete assemblies and short exons remain literal. Only terminal `*` is removed. No external sequence is used and no cross-row assembly is attempted.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
