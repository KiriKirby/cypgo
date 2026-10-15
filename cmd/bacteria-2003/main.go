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
	"unicode/utf16"
)

const (
	sourceFile = "bacteria-bacteria.dbs.txt"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacteria.dbs.doc"
	sourceSHA  = "bc78753c1953f6cc328fdf608454a7669cf9baffc306cb3d53b0e01415c25eed"
)

type record struct {
	Block, Line, FirstSequenceLine, LastSequenceLine int
	Symbol, Species, Header, Sequence, RecordKey     string
	Status                                           []string
}

type excluded struct {
	Block, Line    int
	Header, Reason string
}

var (
	headerRE  = regexp.MustCompile(`^>\s*(.*)$`)
	wordRE    = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([a-z][a-z0-9.-]*)\b`)
	wordCapRE = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([A-Z][A-Za-z0-9-]+)\b`)
)

// These are the only source blocks whose first literal protein segment is
// printed on the header line itself. They were checked against the exact
// source text; no generic "sequence-looking" header scan is used.
var inlineHeaderProtein = map[int]string{
	27:  "MEERRVSISSITWRFPMLFAPVDDVTTIDDLTLDPYPIYRRMRV",
	233: "MQLHELMDPDYSDNPFPLYRKLHQQGPLIPAGDKIIISGSHAVVDALLNDRRVGKNYMESVRVRFGDDAAGLPLFQGISRMFLVLNPPDHNRL",
	323: "MATQQPALVLDPTGADHHTEHRTLREGGPATWVDVLGVQAWSVS",
}

func speciesFromHeader(header string) string {
	if strings.HasPrefix(header, "CYP102B1 ") {
		return "Streptomyces coelicolor"
	}
	if strings.Contains(header, "Streptoalloteichus hindustanus") {
		return "Streptoalloteichus hindustanus"
	}
	if m := wordRE.FindStringSubmatch(header); len(m) == 3 {
		name := m[1] + " " + m[2]
		// The source uses both "sp" and "sp." in headers. Keep the source's
		// strain token as part of the label where it is explicitly present.
		if m[2] == "sp" || m[2] == "sp." || m[2] == "spp" || m[2] == "spp." {
			name = m[1] + " " + strings.TrimSuffix(m[2], ".")
			rest := strings.TrimSpace(header[midx(header, m[0])+len(m[0]):])
			if fs := strings.Fields(rest); len(fs) > 0 {
				x := strings.Trim(fs[0], ",;()[]")
				if x != "" && !strings.HasPrefix(x, "CYP") && !strings.Contains(x, "%") {
					name += " " + x
				}
			}
		}
		return name
	}
	if m := wordCapRE.FindStringSubmatch(header); len(m) == 3 {
		return m[1] + " sp. " + m[2]
	}
	return ""
}

func midx(s, sub string) int { return strings.Index(s, sub) }

func proteinLine(raw string) (string, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false
	}
	if strings.HasSuffix(s, `"`) {
		s = strings.TrimSuffix(s, `"`)
		s = strings.TrimSpace(s)
	}
	fields := strings.Fields(s)
	var b strings.Builder
	for _, f := range fields {
		if _, err := strconv.Atoi(f); err == nil {
			continue
		}
		for i, r := range f {
			// Block 265 has exactly one source-literal lowercase prefix. This is
			// deliberately narrower than accepting arbitrary lowercase prose.
			if i < len("mrtyrtv") && strings.HasPrefix(f, "mrtyrtv") && rune("mrtyrtv"[i]) == r {
				continue
			}
			if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXYx*-", r) {
				return "", false
			}
		}
		b.WriteString(f)
	}
	if b.Len() == 0 {
		return "", false
	}
	return b.String(), true
}

func parse(text string) ([]record, []excluded, error) {
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	var starts []int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			starts = append(starts, i)
		}
	}
	if len(starts) != 555 {
		return nil, nil, fmt.Errorf("bacteria.dbs header count changed: got %d want 555", len(starts))
	}
	var out []record
	var ex []excluded
	for bi, start := range starts {
		end := len(lines)
		if bi+1 < len(starts) {
			end = starts[bi+1]
		}
		header := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[start]), ">"))
		fields := strings.Fields(header)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "CYP") {
			ex = append(ex, excluded{bi + 1, start + 1, header, "accession-only comparison block has no CYP model header"})
			continue
		}
		species := speciesFromHeader(header)
		if species == "" {
			return nil, nil, fmt.Errorf("block %d line %d has no reviewed organism: %s", bi+1, start+1, header)
		}
		r := record{Block: bi + 1, Line: start + 1, Header: header, Symbol: fields[0], Species: species}
		var parts []string
		if p, ok := inlineHeaderProtein[bi+1]; ok {
			parts = append(parts, p)
			r.FirstSequenceLine = start + 1
			r.LastSequenceLine = start + 1
			r.Status = append(r.Status, "source-inline-protein")
		}
		for i := start + 1; i < end; i++ {
			p, ok := proteinLine(lines[i])
			if !ok {
				continue
			}
			if r.FirstSequenceLine == 0 {
				r.FirstSequenceLine = i + 1
			}
			r.LastSequenceLine = i + 1
			parts = append(parts, p)
		}
		r.Sequence = strings.Join(parts, "")
		if r.Sequence == "" {
			return nil, nil, fmt.Errorf("block %d (%s) has no literal protein", bi+1, header)
		}
		if strings.HasSuffix(r.Sequence, "*") {
			r.Sequence = strings.TrimRight(r.Sequence, "*")
			r.Status = append(r.Status, "source-terminal-stop")
		}
		low := strings.ToLower(strings.Join(lines[start:end], " "))
		if strings.HasSuffix(strings.ToUpper(r.Symbol), "P") || strings.Contains(low, "pseudogene") {
			r.Status = append(r.Status, "source-pseudogene")
		}
		for _, marker := range []string{"partial", "fragment", "n-terminal", "c-terminal", "runs off", "missing"} {
			if strings.Contains(low, marker) {
				r.Status = append(r.Status, "source-fragment-or-partial")
				break
			}
		}
		if strings.Contains(low, "confidential") {
			r.Status = append(r.Status, "source-confidential")
		}
		if strings.IndexFunc(r.Sequence, func(c rune) bool { return c >= 'a' && c <= 'z' }) >= 0 {
			r.Status = append(r.Status, "source-lowercase-residue")
		}
		if strings.Contains(r.Sequence, "*") {
			r.Status = append(r.Status, "internal-stop")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Status = append(r.Status, "ambiguous-X-or-x")
		}
		if len(r.Sequence) < 350 {
			r.Status = append(r.Status, "short-sequence")
		}
		r.Status = append([]string{"reviewed-bacteria-2003", "literal-protein"}, r.Status...)
		r.RecordKey = fmt.Sprintf("bacteria-2003:block-%04d:line-%05d:%s", r.Block, r.Line, strings.ToLower(r.Symbol))
		for pos, c := range r.Sequence {
			if r.Block == 265 && pos < len("mrtyrtv") && rune("mrtyrtv"[pos]) == c {
				continue
			}
			if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXYx*-", c) {
				return nil, nil, fmt.Errorf("block %d unexpected residue %q", bi+1, c)
			}
		}
		out = append(out, r)
	}
	if len(out) != 553 || len(ex) != 2 {
		return nil, nil, fmt.Errorf("review counts changed: accepted=%d excluded=%d", len(out), len(ex))
	}
	return out, ex, nil
}

func decodeSource(b []byte) (string, error) {
	if len(b) < 2 || b[0] != 0xff || b[1] != 0xfe || len(b)%2 != 0 {
		return "", fmt.Errorf("bacteria.dbs text is no longer UTF-16LE with BOM")
	}
	u := make([]uint16, 0, (len(b)-2)/2)
	for i := 2; i < len(b); i += 2 {
		u = append(u, uint16(b[i])|uint16(b[i+1])<<8)
	}
	return string(utf16.Decode(u)), nil
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "Word-normalized bacteria.dbs source")
	out := flag.String("out", filepath.Join("sources", "reviewed", "bacteria", "bacteria-2003.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "bacteria", "bacteria-2003.md"), "resource audit")
	flag.Parse()
	b, err := os.ReadFile(*in)
	if err != nil {
		panic(err)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if h != sourceSHA {
		panic(fmt.Sprintf("source SHA-256 changed: %s", h))
	}
	text, err := decodeSource(b)
	if err != nil {
		panic(err)
	}
	rs, ex, err := parse(text)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("2003 bacteria.dbs block %d at normalized line %d; literal protein lines %d-%d; header=%s", r.Block, r.Line, r.FirstSequenceLine, r.LastSequenceLine, r.Header)
		_ = w.Write([]string{"bacteria", r.Species, r.Symbol, r.Symbol, r.RecordKey, sourceURL, note, r.Sequence, sourceFile, sourceSHA, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	_ = f.Close()
	if err := os.MkdirAll(filepath.Dir(*audit), 0o755); err != nil {
		panic(err)
	}
	counts := map[string]int{}
	species := map[string]bool{}
	lengths := make([]int, 0, len(rs))
	for _, r := range rs {
		species[r.Species] = true
		lengths = append(lengths, len(r.Sequence))
		for _, s := range r.Status {
			counts[s]++
		}
	}
	sort.Strings([]string{})
	var ab strings.Builder
	fmt.Fprintf(&ab, "# Bacteria resource review: bacteria.dbs (2003)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Normalized text: UTF-16 Word text export, 5,169 lines, no tables\n- Source title declaration: `430 bacterial cytochrome P450 sequences Nov. 5, 2003`\n- Physical `>` blocks: `555` (553 CYP blocks plus 2 accession-only comparison blocks)\n- Accepted CYP blocks: `%d`\n- Records with literal protein: `%d`\n- Excluded accession-only comparison blocks: `%d`\n- Distinct source organism labels: `%d`\n- Sequence length range: `%d-%d` aa\n- Review status: `complete`\n\n", sourceFile, sourceURL, sourceSHA, len(rs), len(rs), len(ex), len(species), minInt(lengths), maxInt(lengths))
	ab.WriteString("The document title says 556 sequences, but the downloaded normalized source contains 555 physical `>` blocks. Blocks 1-553 have CYP model headers and are retained in source order, including variants, pseudogenes, fragments, half proteins, short N-terminal/C-terminal pieces, duplicate names and confidential source text. Blocks 554-555 are accession-only comparison alignments (`L76374` and `AF071147`) and are excluded because they do not define a CYP record. Protein is read only from literal residue lines inside each block. Standalone numeric coordinate fields and Word line wrapping are removed; five trailing Word quote artifacts are removed only at line ends. Three reviewed blocks carry a literal protein segment on the header line and are handled by explicit block number. Lowercase residues, X/x, gaps and internal stops remain source literal; only terminal `*` markers are removed. No translation, repair, deduplication or external completion is performed.\n\n## Status counts\n\n| Status | Count |\n|---|---:|\n")
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(&ab, "| %s | %d |\n", n, counts[n])
	}
	ab.WriteString("\n## Excluded blocks\n\n| Block | Line | Header | Reason |\n|---:|---:|---|---|\n")
	for _, e := range ex {
		fmt.Fprintf(&ab, "| %d | %d | %s | %s |\n", e.Block, e.Line, md(e.Header), e.Reason)
	}
	ab.WriteString("\n## Representatives\n\n| Accepted index | Block | Line | Species | Symbol | Length | Status |\n|---:|---:|---:|---|---|---:|---|\n")
	for _, i := range []int{0, len(rs) / 2, len(rs) - 1} {
		r := rs[i]
		fmt.Fprintf(&ab, "| %d | %d | %d | %s | %s | %d | %s |\n", i+1, r.Block, r.Line, md(r.Species), r.Symbol, len(r.Sequence), strings.Join(r.Status, ";"))
	}
	if err := os.WriteFile(*audit, []byte(ab.String()), 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("bacteria 2003: %d records, %d sequences, %d organism labels\n", len(rs), len(rs), len(species))
}

func minInt(v []int) int {
	m := v[0]
	for _, x := range v[1:] {
		if x < m {
			m = x
		}
	}
	return m
}
func maxInt(v []int) int {
	m := v[0]
	for _, x := range v[1:] {
		if x > m {
			m = x
		}
	}
	return m
}
func md(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "|", "\\|"), "\n", " ")
}
