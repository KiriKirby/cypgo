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
)

const docFile = "animals-Monarch.butterfly.doc"
const textFile = "animals-Monarch.butterfly.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Monarch.butterfly.doc"

var parRE = regexp.MustCompile(`\([^)]*\)`)
var numRE = regexp.MustCompile(`\d+`)
var keyRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

type record struct {
	Block, Line                       int
	Header, Symbol, ID, Sequence, Key string
	Status                            []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "monarch-butterfly.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "monarch-butterfly.md"), "")
	flag.Parse()
	rs, l, e := parse(*in)
	if e != nil {
		panic(e)
	}
	dh, e := hash(*doc)
	if e != nil {
		panic(e)
	}
	th, e := hash(*in)
	if e != nil {
		panic(e)
	}
	if e = writeCSV(*out, rs, dh, th); e != nil {
		panic(e)
	}
	if e = writeAudit(*audit, rs, l, dh, th); e != nil {
		panic(e)
	}
	fmt.Printf("Danaus plexippus: %d records, all literal protein\n", len(rs))
}
func parse(p string) ([]record, int, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return nil, 0, e
	}
	ls := strings.Split(strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n"), "\n")
	var st []int
	for i, l := range ls {
		if strings.HasPrefix(strings.TrimSpace(l), ">") {
			st = append(st, i)
		}
	}
	st = append(st, len(ls))
	var out []record
	for bi := 0; bi < len(st)-1; bi++ {
		a, z := st[bi], st[bi+1]
		h := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ls[a]), ">"))
		f := strings.Fields(h)
		if len(f) < 2 || !strings.HasPrefix(strings.ToUpper(f[0]), "CYP") || !strings.HasPrefix(f[1], "DPGLEAN") {
			return nil, len(ls), fmt.Errorf("bad header %d %q", bi+1, h)
		}
		sym, id := f[0], f[1]
		var sb strings.Builder
		for _, l := range ls[a+1 : z] {
			q := proteinLine(l)
			sb.WriteString(q)
			if strings.HasSuffix(q, "*") {
				break
			}
		}
		s := sb.String()
		if s == "" {
			return nil, len(ls), fmt.Errorf("no protein %s", id)
		}
		var status []string
		if strings.HasSuffix(s, "*") {
			s = strings.TrimSuffix(s, "*")
			status = append(status, "terminal-stop-removed")
		}
		if strings.Contains(s, "*") {
			status = append(status, "internal-stop-retained")
		}
		if strings.Contains(s, "X") {
			status = append(status, "ambiguous-X-retained")
		}
		lo := strings.ToLower(h)
		for _, m := range []struct{ k, v string }{{"join", "source-joined-model"}, {"partial", "source-partial"}, {"short", "source-short"}} {
			{
				if strings.Contains(lo, m.k) {
					status = append(status, m.v)
				}
			}
		}
		if len(s) < 350 {
			status = append(status, "short-source-fragment")
		}
		key := fmt.Sprintf("monarch:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(keyRE.ReplaceAllString(id, "-")))
		out = append(out, record{bi + 1, a + 1, h, sym, id, s, key, status})
	}
	if len(out) != 86 {
		return nil, len(ls), fmt.Errorf("count=%d want86", len(out))
	}
	return out, len(ls), nil
}
func proteinLine(l string) string {
	l = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(l), "."))
	l = parRE.ReplaceAllString(l, " ")
	l = numRE.ReplaceAllString(l, " ")
	var b strings.Builder
	for _, t := range strings.Fields(l) {
		if !res(t) {
			return ""
		}
		b.WriteString(t)
	}
	return b.String()
}
func res(s string) bool {
	if s == "" {
		return false
	}
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
func writeCSV(p string, rs []record, dh, th string) error {
	if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
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
		_ = w.Write([]string{"animals", "Danaus plexippus", r.Symbol, r.ID, r.Key, sourceURL, fmt.Sprintf("Monarch Word block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header), r.Sequence, docFile, dh, textFile, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p string, rs []record, l int, dh, th string) error {
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
	fmt.Fprintf(&b, "# Animal resource review: Danaus plexippus (monarch butterfly)\n\n- Source: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Text: `%s`\n- Text SHA-256: `%s`\n- Lines: `%d`\n- Explicit blocks: `%d`\n- Literal proteins: `%d`\n- Review status: `complete`\n\n## Interpretation\n\nThe document states that four source sequences were joined to leave 86 sequences, and it physically contains exactly 86 `>` blocks. Each DPGLEAN gene-model block remains distinct even when CYP family labels repeat. Coordinate and phase annotations are removed from literal uppercase protein lines. Parsing stops at a terminal stop so extra exon comparisons and text explicitly marked not-P450 cannot leak into the record. Only terminal stops are removed; internal stops and X remain. No DNA is translated and no external sequence is substituted.\n\n| Status | Records |\n|---|---:|\n", docFile, sourceURL, dh, textFile, th, l, len(rs), len(rs))
	for _, k := range ks {
		fmt.Fprintf(&b, "| %s | %d |\n", k, c[k])
	}
	if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0o644)
}
