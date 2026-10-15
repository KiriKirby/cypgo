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

const docFile = "animals-whitefly.doc"
const textFile = "animals-whitefly.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/whitefly.doc"

var whiteflyKeyRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

type record struct {
	Block, Line                       int
	Header, Symbol, ID, Sequence, Key string
	Status                            []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "whitefly.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "whitefly.md"), "")
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
	fmt.Printf("Trialeurodes vaporariorum: %d records, all literal protein\n", len(rs))
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
		fields := strings.Fields(h)
		if len(fields) == 0 {
			return nil, len(lines), fmt.Errorf("empty header")
		}
		sym := strings.TrimPrefix(fields[0], "*")
		if strings.EqualFold(sym, "whitefly") {
			for _, f := range fields {
				if strings.HasPrefix(strings.ToLower(f), "seq") {
					sym = strings.TrimSuffix(f, "_")
					break
				}
			}
		}
		sym = strings.TrimSuffix(sym, "_")
		var seq strings.Builder
		for _, l := range lines[a+1 : z] {
			q := whiteflyProteinLine(l)
			seq.WriteString(q)
		}
		s := seq.String()
		if s == "" {
			return nil, len(lines), fmt.Errorf("block %d %s no protein", bi+1, sym)
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
		lower := strings.ToLower(h)
		for _, m := range []struct{ k, v string }{{"truncated", "source-truncated"}, {"partial", "source-partial"}, {"missing", "source-missing-region"}, {"may be part", "source-possible-same-gene"}} {
			{
				if strings.Contains(lower, m.k) {
					st = append(st, m.v)
				}
			}
		}
		if !strings.HasPrefix(strings.ToUpper(sym), "CYP") {
			st = append(st, "source-unnamed")
		}
		if len(s) < 350 {
			st = append(st, "short-source-fragment")
		}
		key := fmt.Sprintf("whitefly:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(whiteflyKeyRE.ReplaceAllString(sym, "-")))
		out = append(out, record{bi + 1, a + 1, h, sym, sym, s, key, st})
	}
	if len(out) != 78 {
		return nil, len(lines), fmt.Errorf("headers=%d want78", len(out))
	}
	return out, len(lines), nil
}
func whiteflyProteinLine(line string) string {
	var b strings.Builder
	for _, t := range strings.Fields(strings.TrimSpace(line)) {
		if !whiteflyResidues(t) {
			return ""
		}
		b.WriteString(t)
	}
	return b.String()
}
func whiteflyResidues(s string) bool {
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
		note := fmt.Sprintf("Whitefly Word block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", "Trialeurodes vaporariorum", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, docFile, dh, textFile, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p string, rs []record, lines int, dh, th string) error {
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
	fmt.Fprintf(&b, "# Animal resource review: Trialeurodes vaporariorum (whitefly)\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Text SHA-256: `%s`\n- Normalized lines: `%d`\n- Explicit blocks: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", docFile, sourceURL, dh, textFile, th, lines, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nAll 78 physical `>` blocks are retained in source order. The document mixes named CYPs with four explicitly too-short-to-name sequence blocks; those retain stable source seq identifiers instead of being discarded or guessed. v1/v2/v3 records, truncations and possible same-gene fragments remain distinct because the author presents distinct physical source sequences. Only lines made entirely of uppercase protein tokens are joined; similarity prose is excluded. Terminal stops are removed while internal stops and X remain. No fragments are joined across headers, no DNA is translated, and no external sequence is substituted.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, k := range ks {
		fmt.Fprintf(&b, "| %s | %d |\n", k, c[k])
	}
	if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0o644)
}
