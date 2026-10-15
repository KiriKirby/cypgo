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
)

const pigDoc = "animals-pig.P450s.doc"
const pigText = "animals-pig.P450s.txt"
const pigURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/pig.P450s.doc"

var pigAccRE = regexp.MustCompile(`\b(?:[A-Z]{1,4}_?\d{5,}(?:\.\d+)?|[A-Z]{2}\d{5,}(?:\.\d+)?)\b`)

type pigRecord struct {
	Block, Line                             int
	Header, Symbol, ID, Sequence, RecordKey string
	Statuses                                []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", pigText), "")
	doc := flag.String("doc", filepath.Join("raw", pigDoc), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "pig.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "pig.md"), "")
	flag.Parse()
	rs, lines, e := parsePig(*in)
	if e != nil {
		panic(e)
	}
	dh, e := fileHash(*doc)
	if e != nil {
		panic(e)
	}
	th, e := fileHash(*in)
	if e != nil {
		panic(e)
	}
	if e = writePigCSV(*out, rs, dh, th); e != nil {
		panic(e)
	}
	if e = writePigAudit(*audit, rs, lines, dh, th); e != nil {
		panic(e)
	}
	n := 0
	for _, r := range rs {
		if r.Sequence != "" {
			n++
		}
	}
	fmt.Printf("Sus scrofa: %d headers, %d literal proteins, %d missing\n", len(rs), n, len(rs)-n)
}
func parsePig(path string) ([]pigRecord, int, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, 0, e
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	var starts []int
	for i, x := range lines {
		if strings.HasPrefix(strings.TrimSpace(x), ">") {
			starts = append(starts, i)
		}
	}
	starts = append(starts, len(lines))
	var out []pigRecord
	for bi := 0; bi < len(starts)-1; bi++ {
		a, b := starts[bi], starts[bi+1]
		h := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[a]), ">"))
		sym := strings.Fields(h)[0]
		var runs []string
		var cur strings.Builder
		flush := func() {
			if cur.Len() > 0 {
				runs = append(runs, cur.String())
				cur.Reset()
			}
		}
		for _, ln := range lines[a+1 : b] {
			q := strings.Join(strings.Fields(ln), "")
			q = regexp.MustCompile(`^\d+`).ReplaceAllString(q, "")
			q = regexp.MustCompile(`\d+$`).ReplaceAllString(q, "")
			q = strings.TrimSuffix(q, "(1)")
			if q != "" && pigResidues(q) {
				cur.WriteString(q)
			} else {
				flush()
			}
		}
		flush()
		seq := ""
		for _, q := range runs {
			if len(q) > len(seq) {
				seq = q
			}
		}
		var st []string
		if strings.HasSuffix(seq, "*") {
			seq = strings.TrimSuffix(seq, "*")
			st = append(st, "terminal-stop-removed")
		}
		if strings.Contains(seq, "*") {
			st = append(st, "internal-stop-retained")
		}
		if strings.IndexFunc(seq, unicode.IsLower) >= 0 {
			st = append(st, "source-lowercase-retained")
		}
		if strings.ContainsAny(seq, "Xx") {
			st = append(st, "ambiguous-X-retained")
		}
		if seq != "" && len(seq) < 350 {
			st = append(st, "short-source-fragment")
		}
		lh := strings.ToLower(h)
		if strings.Contains(lh, "not found") {
			st = append(st, "source-sequence-missing")
		}
		id := sym
		if m := pigAccRE.FindString(h); m != "" {
			id = m
		}
		out = append(out, pigRecord{Block: bi + 1, Line: a + 1, Header: h, Symbol: sym, ID: id, Sequence: seq, RecordKey: fmt.Sprintf("pig:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(sym)), Statuses: st})
	}
	if len(out) != 87 {
		return nil, len(lines), fmt.Errorf("pig headers=%d want 87", len(out))
	}
	return out, len(lines), nil
}
func pigResidues(s string) bool {
	if len(s) < 10 {
		return false
	}
	for _, r := range s {
		u := unicode.ToUpper(r)
		if (u < 'A' || u > 'Z') && r != '*' && r != '-' {
			return false
		}
	}
	return true
}
func fileHash(p string) (string, error) {
	d, e := os.ReadFile(p)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(d)), nil
}
func writePigCSV(p string, rs []pigRecord, dh, th string) error {
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
		note := fmt.Sprintf("Pig Word document block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", "Sus scrofa", r.Symbol, r.ID, r.RecordKey, pigURL, note, r.Sequence, pigDoc, dh, pigText, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}
func writePigAudit(p string, rs []pigRecord, lines int, dh, th string) error {
	if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
		return e
	}
	c := map[string]int{}
	seq := 0
	for _, r := range rs {
		if r.Sequence != "" {
			seq++
		}
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
	fmt.Fprintf(&b, "# Animal resource review: Sus scrofa (pig)\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed Word-normalized text: `%s`\n- Text SHA-256: `%s`\n- Word layout: `746 paragraphs`, `13 pages`, `0 tables`\n- Normalized lines: `%d`\n- Explicit CYP headers: `%d`\n- Literal protein sequences: `%d`\n- Missing sequences: `%d`\n- Review status: `complete`\n\n", pigDoc, pigURL, dh, pigText, th, lines, len(rs), seq, len(rs)-seq)
	b.WriteString("## Resource-specific interpretation\n\nThe source says 59 different genes but physically contains 87 explicit `>` blocks, including variants, fragments and entries explicitly stating that no ortholog was found. All 87 source blocks remain records. Protein is taken only from complete residue lines inside the same header block; coordinate numbers and the one reviewed `(1)` phase marker in CYP2R1 are removed. CYP3A88 prose before its sequence is excluded. The longest uninterrupted literal residue run is used for this document because some blocks contain prose between independent fragments; CYP2R1's reviewed phase marker remains within one run. Lowercase cow-comparison residues, internal stops and fragments remain literal. Only terminal stops are removed. No missing record is filled externally.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	return os.WriteFile(p, []byte(b.String()), 0o644)
}
