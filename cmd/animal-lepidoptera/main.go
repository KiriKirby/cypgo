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

const docFile = "animals-lepidoptera.doc"
const textFile = "animals-lepidoptera.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/lepidoptera.doc"

var lp = regexp.MustCompile(`\([^)]*\)`)
var ln = regexp.MustCompile(`\d+`)
var lk = regexp.MustCompile(`[^A-Za-z0-9]+`)
var acc = regexp.MustCompile(`\b[A-Z]{1,4}\d{5,}(?:\.\d+)?\b`)

type record struct {
	Block, Line                                int
	Header, Species, Symbol, ID, Sequence, Key string
	Status                                     []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "lepidoptera.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "lepidoptera.md"), "")
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
	fmt.Printf("Lepidoptera public collection: %d records, all literal protein\n", len(rs))
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
		sym := f[0]
		id := acc.FindString(h)
		if id == "" {
			id = sym
		}
		species := headerSpecies(h)
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
			return nil, len(ls), fmt.Errorf("block %d %q no protein", bi+1, h)
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
		if species == "Lepidoptera public collection" {
			status = append(status, "source-species-unresolved")
		}
		lo := strings.ToLower(h)
		if strings.Contains(lo, "est") {
			status = append(status, "source-est-protein")
		}
		if len(s) < 350 {
			status = append(status, "short-source-fragment")
		}
		key := fmt.Sprintf("lepidoptera:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(lk.ReplaceAllString(id, "-")))
		out = append(out, record{bi + 1, a + 1, h, species, sym, id, s, key, status})
	}
	if len(out) != 94 {
		return nil, len(ls), fmt.Errorf("count=%d", len(out))
	}
	return out, len(ls), nil
}
func headerSpecies(h string) string {
	for _, x := range []string{"Manduca sexta", "Helicoverpa armigera", "Helicoverpa zea", "Mamestra brassicae", "Papilio polyxenes", "Papilio glaucus", "Papilio canadensis", "Heliothis virescens", "Depressaria pastinacella", "Spodoptera litura", "Plutella xylostella", "Bombyx mori", "Drosophila melanogaster"} {
		if strings.Contains(strings.ToLower(h), strings.ToLower(x)) {
			return x
		}
	}
	return "Lepidoptera public collection"
}
func proteinLine(l string) string {
	l = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(l), "."))
	l = lp.ReplaceAllString(l, " ")
	l = ln.ReplaceAllString(l, " ")
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
		_ = w.Write([]string{"animals", r.Species, r.Symbol, r.ID, r.Key, sourceURL, fmt.Sprintf("Lepidoptera Word block %d at line %d; exact header=%s", r.Block, r.Line, r.Header), r.Sequence, docFile, dh, textFile, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p string, rs []record, l int, dh, th string) error {
	c := map[string]int{}
	sp := map[string]int{}
	for _, r := range rs {
		sp[r.Species]++
		for _, s := range r.Status {
			c[s]++
		}
	}
	ks := make([]string, 0, len(c))
	for k := range c {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	ss := make([]string, 0, len(sp))
	for s := range sp {
		ss = append(ss, s)
	}
	sort.Strings(ss)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: public Lepidoptera collection\n\n- Source: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Text SHA-256: `%s`\n- Lines: `%d`\n- Explicit blocks: `%d`\n- Literal proteins: `%d`\n- Review: `complete`\n\n## Interpretation\n\nAll 94 physical records are retained, including EST fragments, variants and a public Drosophila comparison record. Species is assigned only when the exact header states a reviewed binomial; otherwise the record stays under `Lepidoptera public collection`, avoiding inference from neighboring blocks. Protein is limited to literal uppercase source lines; coordinate/phase notes are removed and parsing stops at terminal stop. No DNA is translated and no external sequence is substituted.\n\n## Species counts\n\n", docFile, sourceURL, dh, th, l, len(rs), len(rs))
	for _, s := range ss {
		fmt.Fprintf(&b, "- %s: %d\n", s, sp[s])
	}
	b.WriteString("\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, k := range ks {
		fmt.Fprintf(&b, "| %s | %d |\n", k, c[k])
	}
	if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0o644)
}
