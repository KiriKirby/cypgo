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

type profile struct {
	Slug, File, URL, Sheet, Species      string
	First, Last                          int
	SymbolCol, SeqCol, IDCol, CommentCol string
	Headers                              bool
	Expected                             int
}

var profiles = []profile{
	{"turkey", "animals-Meleagris.gallopavo.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Meleagris.gallopavo.P450s.xlsx", "Sheet1", "Meleagris gallopavo", 2, 61, "F", "G", "A", "", true, 60},
	{"zebrafinch", "animals-Zebrafinch.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Zebrafinch.P450s.xlsx", "Sheet1", "Taeniopygia guttata", 2, 56, "F", "H", "A", "", true, 55},
	{"mallard", "animals-Anas.platyrhynchos.duck.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Anas.platyrhynchos.duck.xlsx", "Sheet1", "Anas platyrhynchos", 2, 49, "L", "N", "A", "M", true, 48},
	{"flycatcher", "animals-Flycatcher.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Flycatcher.P450s.xlsx", "sorted by location", "Ficedula albicollis", 2, 45, "L", "N", "A", "M", true, 44},
	{"chimney-swift", "animals-Chaetura.pelagica.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Chaetura.pelagica.P450s.xlsx", "reduced set", "Chaetura pelagica", 1, 48, "H", "I", "B", "F", false, 48},
}

type record struct {
	Row                                int
	Symbol, ID, Comment, Sequence, Key string
	Status                             []string
}

var keyRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func main() {
	for _, p := range profiles {
		if e := run(p); e != nil {
			panic(fmt.Errorf("%s: %w", p.Slug, e))
		}
	}
}
func run(p profile) error {
	path := filepath.Join("raw", p.File)
	names, e := plantxlsx.SheetNames(path)
	if e != nil {
		return e
	}
	found := false
	for _, n := range names {
		if n == p.Sheet {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("sheet %q absent from %q", p.Sheet, names)
	}
	wb, e := plantxlsx.ReadSheet(path, p.Sheet, false)
	if e != nil {
		return e
	}
	rs, e := review(p, wb.Rows)
	if e != nil {
		return e
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if e = writeCSV(p, rs, h); e != nil {
		return e
	}
	if e = writeAudit(p, rs, wb, h, names); e != nil {
		return e
	}
	fmt.Printf("%s: %d literal proteins\n", p.Species, len(rs))
	return nil
}
func review(p profile, rows map[int]map[string]string) ([]record, error) {
	var out []record
	for row := p.First; row <= p.Last; row++ {
		v := rows[row]
		sym := strings.TrimSpace(v[p.SymbolCol])
		raw := strings.TrimSpace(v[p.SeqCol])
		if sym == "" || raw == "" {
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
		s := b.String()
		var st []string
		if strings.HasSuffix(s, "*") {
			s = strings.TrimSuffix(s, "*")
			st = append(st, "terminal-stop-removed")
		}
		if strings.Contains(s, "*") {
			st = append(st, "internal-stop-retained")
		}
		if strings.ContainsAny(s, "Xx") {
			st = append(st, "ambiguous-X-retained")
		}
		if lower {
			st = append(st, "source-lowercase-retained")
		}
		if len(s) < 350 {
			st = append(st, "short-source-fragment")
		}
		comment := strings.TrimSpace(v[p.CommentCol])
		lc := strings.ToLower(comment + " " + sym)
		if strings.Contains(lc, "partial") || strings.Contains(lc, "missing") || strings.Contains(lc, "pseudo") || strings.Contains(lc, "error") {
			st = append(st, "source-partial-or-pseudogene")
		}
		id := strings.TrimSpace(v[p.IDCol])
		if id == "" {
			id = sym
		}
		key := fmt.Sprintf("%s:sheet-%s:row-%04d:%s", p.Slug, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(p.Sheet), "-"), "-"), row, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(id), "-"), "-"))
		out = append(out, record{row, sym, id, comment, s, key, st})
	}
	if len(out) != p.Expected {
		return nil, fmt.Errorf("count=%d want %d", len(out), p.Expected)
	}
	return out, nil
}
func writeCSV(p profile, rs []record, h string) error {
	path := filepath.Join("sources", "reviewed", "animals", p.Slug+".csv")
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "source_comment", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("%s workbook authoritative sheet %q row %d; exact CYP name=%s; ID=%s; comment=%s", p.Species, p.Sheet, r.Row, r.Symbol, r.ID, r.Comment)
		_ = w.Write([]string{"animals", p.Species, r.Symbol, r.ID, r.Key, p.URL, note, r.Sequence, p.File, h, p.Sheet, strconv.Itoa(r.Row), r.Comment, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p profile, rs []record, wb *plantxlsx.Workbook, h string, names []string) error {
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
	fmt.Fprintf(&b, "# Animal resource review: %s\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Workbook sheets: `%s`\n- Authoritative sheet: `%s`\n- Exact accepted rows: `%d`–`%d`\n- Exact CYP / sequence columns: `%s` / `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Literal proteins: `%d`\n- Review status: `complete`\n\n", p.Species, p.File, p.URL, h, strings.Join(names, "`, `"), p.Sheet, p.First, p.Last, p.SymbolCol, p.SeqCol, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThis workbook uses the exact sheet, row interval and columns listed above; those coordinates are fixed for this resource and validated before output. Candidate/raw sheets or alternative orderings are not appended. Every accepted row has an explicit CYP name and literal source protein. IDs and comments remain provenance. Terminal stops are removed; internal stops, X/x, lowercase, gaps, partials, duplicates and pseudogenes remain literal and source-distinct. No translation, repair or external sequence lookup is performed.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	path := filepath.Join("docs", "animals", p.Slug+".md")
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
