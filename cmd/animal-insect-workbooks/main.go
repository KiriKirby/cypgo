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
	Expected                             int
}

var profiles = []profile{
	{"apis-florea", "animals-Apis.florea.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Apis.florea.P450s.xlsx", "Apis.florea.csv.txt", "Apis florea", 2, 45, "G", "H", "A", "F", 44},
	{"bombus-terrestris", "animals-Bombus.terrestris.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Bombus.terrestris.P450s.xlsx", "reduced set (dups removed)", "Bombus terrestris", 2, 56, "G", "H", "A", "F", 55},
	{"rhodnius-prolixus", "animals-Rhodnius.prolixus.P450s.xlsx", "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Rhodnius.prolixus.P450s.xlsx", "Sheet1", "Rhodnius prolixus", 2, 128, "C", "M", "B", "", 127},
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
		return fmt.Errorf("sheet absent: %v", names)
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
		comment := strings.TrimSpace(v[p.CommentCol])
		lc := strings.ToLower(comment + " " + sym)
		if strings.Contains(lc, "partial") || strings.Contains(lc, "pseudo") || strings.Contains(lc, "missing") || strings.Contains(lc, "frameshift") {
			st = append(st, "source-partial-or-pseudogene")
		}
		if len(s) < 350 {
			st = append(st, "short-source-fragment")
		}
		id := strings.TrimSpace(v[p.IDCol])
		if id == "" {
			id = sym
		}
		key := fmt.Sprintf("%s:sheet-%s:row-%04d:%s", p.Slug, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(p.Sheet), "-"), "-"), row, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(id), "-"), "-"))
		out = append(out, record{row, sym, id, comment, s, key, st})
	}
	if len(out) != p.Expected {
		return nil, fmt.Errorf("count=%d want%d", len(out), p.Expected)
	}
	return out, nil
}
func writeCSV(p profile, rs []record, h string) error {
	path := filepath.Join("sources", "reviewed", "animals", p.Slug+".csv")
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
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
		_ = w.Write([]string{"animals", p.Species, r.Symbol, r.ID, r.Key, p.URL, fmt.Sprintf("%s authoritative sheet %q row %d; CYP=%s; ID=%s", p.Species, p.Sheet, r.Row, r.Symbol, r.ID), r.Sequence, p.File, h, p.Sheet, strconv.Itoa(r.Row), r.Comment, strings.Join(r.Status, ";")})
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
	ks := make([]string, 0, len(c))
	for k := range c {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: %s\n\n- File: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Sheets: `%s`\n- Authoritative sheet: `%s`\n- Rows: `%d`–`%d`\n- CYP / protein / ID columns: `%s` / `%s` / `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Records/proteins: `%d/%d`\n- Review: `complete`\n\nCandidate, comparison, self-BLAST, raw and alternate-order sheets are inspected but not stacked onto the fixed authoritative profile. Literal source protein is retained, terminal stop alone is removed, and no external sequence is substituted.\n\n| Status | Records |\n|---|---:|\n", p.Species, p.File, p.URL, h, strings.Join(names, "`, `"), p.Sheet, p.First, p.Last, p.SymbolCol, p.SeqCol, p.IDCol, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	for _, k := range ks {
		fmt.Fprintf(&b, "| %s | %d |\n", k, c[k])
	}
	path := filepath.Join("docs", "animals", p.Slug+".md")
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
