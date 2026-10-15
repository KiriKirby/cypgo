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

const (
	sourceFile = "bacteria-allbacteria.txt"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/allbacteria.doc"
	sourceSHA  = "ab82910e9f43b306de8946f723df050502eddb088e11c6920c41922c780f819b"
)

type record struct {
	Block, Line                                  int
	Symbol, Species, Header, Sequence, RecordKey string
	Status                                       []string
}

var binomial = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([a-z][a-z0-9.-]*)\b`)
var capped = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([A-Z][A-Za-z0-9-]+)\b`)

func species(h string) string {
	low := strings.ToLower(h)
	for _, x := range [][2]string{{"uncultured soil bacterium", "uncultured soil bacterium"}, {"uncultured haloarchaeon", "uncultured haloarchaeon FLAS10H9"}, {"seawater sequencing", "seawater environmental sample"}, {"marine metagenome", "marine metagenome"}} {
		if strings.Contains(low, x[0]) {
			return x[1]
		}
	}
	if strings.HasPrefix(h, "SRE9564") {
		return "Thermobifida fusca"
	}
	if strings.Contains(h, "SP0881") {
		return "Streptomyces peucetius"
	}
	if regexp.MustCompile(`\bsce\d+\b`).MatchString(h) {
		return "Sorangium cellulosum"
	}
	m := binomial.FindStringSubmatch(h)
	if len(m) == 3 {
		n := m[1] + " " + m[2]
		if m[2] == "sp." || m[2] == "sp" || m[2] == "spp" {
			fs := strings.Fields(h[midx(h, m[0])+len(m[0]):])
			if len(fs) > 0 && !strings.HasPrefix(fs[0], "CYP") {
				n += " " + strings.Trim(fs[0], ",;()[]")
			}
		}
		return n
	}
	m = capped.FindStringSubmatch(h)
	if len(m) == 3 {
		return m[1] + " sp. " + m[2]
	}
	return ""
}
func midx(s, x string) int { return strings.Index(s, x) }
func protein(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, "abcdefghijklmnopqrstuvwxyz") {
		return ""
	}
	out := ""
	for _, f := range strings.Fields(strings.TrimSuffix(s, `"`)) {
		if _, e := strconv.Atoi(f); e == nil {
			continue
		}
		if !strings.ContainsAny(f, `abcdefghijklmnopqrstuvwxyz+=>"`) {
			out += f
		} else {
			return ""
		}
	}
	if len(out) < 10 {
		return ""
	}
	return out
}
func parse(text string) ([]record, error) {
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	var hs []int
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), ">") {
			hs = append(hs, i)
		}
	}
	if len(hs) != 967 {
		return nil, fmt.Errorf("header count %d", len(hs))
	}
	out := make([]record, 0, len(hs))
	for n, s := range hs {
		e := len(lines)
		if n+1 < len(hs) {
			e = hs[n+1]
		}
		h := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[s]), ">"))
		fs := strings.Fields(h)
		r := record{Block: n + 1, Line: s + 1, Header: h, Symbol: fs[0], Species: species(h)}
		if r.Species == "" {
			r.Species = species(strings.Join(lines[s+1:min(e, s+5)], " "))
		}
		var runs []string
		cur := ""
		for _, l := range lines[s+1 : e] {
			p := protein(l)
			if p == "" {
				if len(cur) > 0 {
					runs = append(runs, cur)
					cur = ""
				}
			} else {
				cur += p
			}
		}
		if cur != "" {
			runs = append(runs, cur)
		}
		for _, q := range runs {
			if len(q) > len(r.Sequence) {
				r.Sequence = q
			}
		}
		r.Sequence = strings.TrimRight(r.Sequence, "*")
		r.RecordKey = fmt.Sprintf("bacteria-2009:block-%04d:line-%05d:%s", r.Block, r.Line, strings.ToLower(r.Symbol))
		r.Status = []string{"reviewed-bacteria-2009"}
		if r.Sequence != "" {
			r.Status = append(r.Status, "literal-protein")
		} else {
			r.Status = append(r.Status, "sequence-missing")
		}
		low := strings.ToLower(strings.Join(lines[s:e], " "))
		for _, m := range []string{"fragment", "partial", "pseudogene", "variant", "renamed", "duplicate", "deletion", "confidential", "frameshift"} {
			if strings.Contains(low, m) {
				r.Status = append(r.Status, "source-"+m)
			}
		}
		if strings.Contains(r.Sequence, "*") {
			r.Status = append(r.Status, "internal-stop")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Status = append(r.Status, "ambiguous-X-or-x")
		}
		if len(r.Sequence) > 0 && len(r.Sequence) < 350 {
			r.Status = append(r.Status, "short-sequence")
		}
		out = append(out, r)
	}
	return out, nil
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "bacteria", "bacteria-2009.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "bacteria", "bacteria-2009.md"), "")
	flag.Parse()
	b, e := os.ReadFile(*in)
	if e != nil {
		panic(e)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if h != sourceSHA {
		panic("source hash changed " + h)
	}
	rs, e := parse(string(b))
	if e != nil {
		panic(e)
	}
	os.MkdirAll(filepath.Dir(*out), 0755)
	f, _ := os.Create(*out)
	w := csv.NewWriter(f)
	w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("2009 allbacteria block %d at normalized line %d; header=%s", r.Block, r.Line, r.Header)
		w.Write([]string{"bacteria", r.Species, r.Symbol, r.Symbol, r.RecordKey, sourceURL, note, r.Sequence, sourceFile, sourceSHA, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	f.Close()
	os.MkdirAll(filepath.Dir(*audit), 0755)
	nseq := 0
	sp := map[string]bool{}
	cnt := map[string]int{}
	for _, r := range rs {
		if r.Sequence != "" {
			nseq++
		}
		sp[r.Species] = true
		for _, s := range r.Status {
			cnt[s]++
		}
	}
	names := make([]string, 0, len(cnt))
	for n := range cnt {
		names = append(names, n)
	}
	sort.Strings(names)
	var a strings.Builder
	fmt.Fprintf(&a, "# Bacteria resource review: allbacteria (2009)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Physical `>` blocks: `967`\n- Accepted source blocks: `%d`\n- Blocks with literal protein: `%d`\n- Missing-sequence blocks retained with provenance: `%d`\n- Distinct organism labels: `%d`\n- Review status: `complete`\n\nThe source title states 923 updated sequences (and an earlier 892-entry release), while the physical normalized file contains 967 CYP blocks. Every block is retained independently in source order, including variants, renamed/deleted records, duplicate names, pseudogenes, fragments and sequence-missing genome annotations. Protein comes only from the longest uninterrupted literal residue run inside that block; coordinates, accession lines, similarity prose and annotations are never concatenated. Only terminal `*` is removed; X/x, gaps, internal stops and short fragments remain literal.\n\n## Status counts\n\n| Status | Count |\n|---|---:|\n", sourceFile, sourceURL, sourceSHA, len(rs), nseq, len(rs)-nseq, len(sp))
	for _, n := range names {
		fmt.Fprintf(&a, "| %s | %d |\n", n, cnt[n])
	}
	os.WriteFile(*audit, []byte(a.String()), 0644)
	fmt.Printf("bacteria 2009: %d records, %d sequences, %d organism labels\n", len(rs), nseq, len(sp))
}
