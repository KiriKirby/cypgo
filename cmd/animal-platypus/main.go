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

const docFile = "animals-duckbill.doc"
const textFile = "animals-duckbill.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/duckbill.doc"

var accRE = regexp.MustCompile(`\b(?:[A-Z]{1,4}_?\d{5,}(?:\.\d+)?|[A-Z]{2}\d{5,}(?:\.\d+)?)\b`)
var phaseRE = regexp.MustCompile(`^\d+|\d+$`)

type record struct {
	Block, Line                       int
	Header, Symbol, ID, Sequence, Key string
	Status                            []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "platypus.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "platypus.md"), "")
	flag.Parse()
	rs, lines, e := parse(*in)
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
	if e = writeAudit(*audit, rs, lines, dh, th); e != nil {
		panic(e)
	}
	fmt.Printf("Ornithorhynchus anatinus: %d records, all literal protein\n", len(rs))
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
		if len(fs) == 0 || !strings.HasPrefix(strings.ToUpper(fs[0]), "CYP") {
			return nil, len(lines), fmt.Errorf("block %d header %q", bi+1, h)
		}
		sym := fs[0]
		var seq strings.Builder
		for _, line := range lines[a+1 : z] {
			q := strings.Join(strings.Fields(line), "")
			q = phaseRE.ReplaceAllString(q, "")
			if len(q) >= 10 && isUpperResidues(q) {
				seq.WriteString(q)
			}
		}
		s := seq.String()
		if s == "" {
			return nil, len(lines), fmt.Errorf("block %d no literal protein", bi+1)
		}
		var st []string
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
		if strings.Contains(s, "-") {
			st = append(st, "source-gap-retained")
		}
		if len(s) < 350 {
			st = append(st, "short-source-fragment")
		}
		lh := strings.ToLower(h)
		if strings.Contains(lh, "partial") || strings.Contains(lh, "fragment") {
			st = append(st, "source-partial")
		}
		id := sym
		if m := accRE.FindString(h); m != "" {
			id = m
		}
		key := fmt.Sprintf("platypus:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(regexp.MustCompile(`[^A-Za-z0-9]+`).ReplaceAllString(id, "-")))
		out = append(out, record{bi + 1, a + 1, h, sym, id, s, key, st})
	}
	if len(out) != 55 {
		return nil, len(lines), fmt.Errorf("headers=%d want 55", len(out))
	}
	return out, len(lines), nil
}
func isUpperResidues(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != '*' && r != '-' {
			return false
		}
	}
	return true
}
func hash(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}
func writeCSV(path string, rs []record, dh, th string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "normalized_file", "normalized_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Platypus Word block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", "Ornithorhynchus anatinus", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, docFile, dh, textFile, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(path string, rs []record, lines int, dh, th string) error {
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
	fmt.Fprintf(&b, "# Animal resource review: Ornithorhynchus anatinus (duck-billed platypus)\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Text SHA-256: `%s`\n- Normalized lines: `%d`\n- Explicit CYP blocks: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", docFile, sourceURL, dh, textFile, th, lines, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe document declares 55 Entrez-protein results, two false positives, and 53 matching P450 sequences, but physically provides 55 explicit CYP-labelled blocks representing the author's curated collection, including source-distinct genomic reconstructions and a combined CYP17A1 record. All 55 blocks are retained in source order. Within each block, only all-uppercase literal residue lines are joined; mixed-case prose, coordinates, `(gap)` markers, BLAST comments and nucleotide annotations are excluded. This resource-specific rule permits separated literal fragments in a single named block without accepting prose. Terminal stops are removed; internal stops, X and source fragments remain literal. No DNA is translated and no external sequence is substituted.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}
