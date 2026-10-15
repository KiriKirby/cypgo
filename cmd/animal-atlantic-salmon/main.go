package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const docFile = "animals-salmo.doc"
const textFile = "animals-salmo.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/salmo.doc"

var accRE = regexp.MustCompile(`\b(?:[A-Z]{1,4}_?\d{5,}(?:\.\d+)?|[A-Z]{2}\d{5,}(?:\.\d+)?)\b`)
var edgeNum = regexp.MustCompile(`^\d+|\d+$`)

type record struct {
	Block, Line                       int
	Header, Symbol, ID, Sequence, Key string
	Status                            []string
}

func main() {
	rs, lines, e := parse(filepath.Join("raw", textFile))
	if e != nil {
		panic(e)
	}
	dh, e := hash(filepath.Join("raw", docFile))
	if e != nil {
		panic(e)
	}
	th, e := hash(filepath.Join("raw", textFile))
	if e != nil {
		panic(e)
	}
	if e = writeCSV(rs, dh, th); e != nil {
		panic(e)
	}
	if e = writeAudit(rs, lines, dh, th); e != nil {
		panic(e)
	}
	n := 0
	for _, r := range rs {
		if r.Sequence != "" {
			n++
		}
	}
	fmt.Printf("Salmo salar: %d records, %d literal proteins\n", len(rs), n)
}
func parse(path string) ([]record, int, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return nil, 0, e
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	var starts []int
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), ">") {
			starts = append(starts, i)
		}
	}
	starts = append(starts, len(lines))
	var out []record
	for bi := 0; bi < len(starts)-1; bi++ {
		a, z := starts[bi], starts[bi+1]
		h := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[a]), ">"))
		fs := strings.Fields(h)
		symbol := ""
		for _, f := range fs {
			if strings.HasPrefix(strings.ToUpper(f), "CYP") {
				symbol = f
				break
			}
		}
		if symbol == "" {
			return nil, len(lines), fmt.Errorf("header %q", h)
		}
		var seq strings.Builder
		for _, line := range lines[a+1 : z] {
			q := strings.Join(strings.Fields(line), "")
			q = edgeNum.ReplaceAllString(q, "")
			if len(q) >= 10 && upperResidues(q) {
				seq.WriteString(q)
			}
		}
		s := seq.String()
		var st []string
		if s == "" {
			st = append(st, "source-sequence-missing")
		} else {
			if strings.HasSuffix(s, "*") {
				s = strings.TrimSuffix(s, "*")
				st = append(st, "terminal-stop-removed")
			}
			if strings.Contains(s, "*") {
				st = append(st, "internal-stop-retained")
			}
			if strings.Contains(s, "X") {
				st = append(st, "ambiguous-X-retained")
			}
			if len(s) < 350 {
				st = append(st, "short-source-fragment")
			}
		}
		lh := strings.ToLower(h)
		if strings.Contains(lh, "partial") || strings.Contains(lh, "fragment") || strings.Contains(lh, "missing") || strings.Contains(lh, "not found") {
			st = append(st, "source-partial-or-missing")
		}
		id := symbol
		if m := accRE.FindString(h); m != "" {
			id = m
		}
		key := fmt.Sprintf("atlantic-salmon:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(regexp.MustCompile(`[^A-Za-z0-9]+`).ReplaceAllString(id, "-")))
		out = append(out, record{bi + 1, a + 1, h, symbol, id, s, key, st})
	}
	if len(out) != 64 {
		return nil, len(lines), fmt.Errorf("count=%d want64", len(out))
	}
	return out, len(lines), nil
}
func upperResidues(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != '*' && r != '-' {
			return false
		}
	}
	return true
}
func hash(p string) (string, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}
func writeCSV(rs []record, dh, th string) error {
	p := filepath.Join("sources", "reviewed", "animals", "atlantic-salmon.csv")
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	f, e := os.Create(p)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "normalized_file", "normalized_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Atlantic salmon Word block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", "Salmo salar", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, docFile, dh, textFile, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(rs []record, lines int, dh, th string) error {
	c := map[string]int{}
	n := 0
	for _, r := range rs {
		if r.Sequence != "" {
			n++
		}
		for _, s := range r.Status {
			c[s]++
		}
	}
	ns := make([]string, 0, len(c))
	for x := range c {
		ns = append(ns, x)
	}
	sort.Strings(ns)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Salmo salar (Atlantic salmon)\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Text SHA-256: `%s`\n- Normalized lines: `%d`\n- Explicit CYP blocks: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `%d`\n- Review status: `complete`\n\n", docFile, sourceURL, dh, textFile, th, lines, len(rs), n, len(rs)-n)
	b.WriteString("## Resource-specific interpretation\n\nThe introductory 54-protein statement describes a 2008 class assembly, while the physical document contains 64 explicit CYP blocks, including later revisions, paralogs, alternates, short source fragments, one explicit `not found` entry and comparison/reference records. All explicit blocks remain source-distinct. Within a block, only all-uppercase literal residue lines are joined; author prose, coordinates and translator notes are excluded. This retains separated fragments and revised blocks without translating EST/DNA. Only terminal stops are removed; internal stops and X remain literal. The missing CYP27B record stays empty and nothing is fetched externally.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, x := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", x, c[x])
	}
	p := filepath.Join("docs", "animals", "atlantic-salmon.md")
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0644)
}
