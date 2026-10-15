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

const docFile = "animals-C.briggsae.doc"
const textFile = "animals-C.briggsae.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/C.briggsae.doc"

var coordinateRE = regexp.MustCompile(`\d+`)
var parentheticalRE = regexp.MustCompile(`\([^)]*\)`)
var safeKeyRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

type record struct {
	Block, Line                       int
	Header, Symbol, ID, Sequence, Key string
	Status                            []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "c-briggsae.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "c-briggsae.md"), "")
	flag.Parse()
	rs, lines, err := parse(*in)
	if err != nil {
		panic(err)
	}
	docHash, err := hash(*doc)
	if err != nil {
		panic(err)
	}
	textHash, err := hash(*in)
	if err != nil {
		panic(err)
	}
	if err = writeCSV(*out, rs, docHash, textHash); err != nil {
		panic(err)
	}
	if err = writeAudit(*audit, rs, lines, docHash, textHash); err != nil {
		panic(err)
	}
	fmt.Printf("Caenorhabditis briggsae: %d records, all literal protein\n", len(rs))
}

func parse(path string) ([]record, int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	var starts []int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			starts = append(starts, i)
		}
	}
	starts = append(starts, len(lines))
	var out []record
	for block := 0; block < len(starts)-1; block++ {
		start, end := starts[block], starts[block+1]
		header := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[start]), ">"))
		fields := strings.Fields(header)
		if len(fields) == 0 || !strings.HasPrefix(strings.ToUpper(fields[0]), "CYP") {
			return nil, len(lines), fmt.Errorf("block %d header %q", block+1, header)
		}
		symbol := fields[0]
		var sequence strings.Builder
		for _, line := range lines[start+1 : end] {
			sequence.WriteString(proteinOnCoordinateLine(line))
		}
		seq := sequence.String()
		if seq == "" {
			return nil, len(lines), fmt.Errorf("block %d has no literal protein", block+1)
		}
		var status []string
		if strings.HasSuffix(seq, "*") {
			seq = strings.TrimSuffix(seq, "*")
			status = append(status, "terminal-stop-removed")
		}
		if strings.Contains(seq, "*") {
			status = append(status, "internal-stop-retained")
		}
		if strings.Contains(seq, "X") {
			status = append(status, "ambiguous-X-retained")
		}
		lowerHeader := strings.ToLower(header)
		if strings.Contains(lowerHeader, "pseudogene") {
			status = append(status, "source-pseudogene")
		}
		if strings.Contains(lowerHeader, "possible pseudogene") {
			status = append(status, "source-possible-pseudogene")
		}
		if len(seq) < 350 {
			status = append(status, "short-source-fragment")
		}
		key := fmt.Sprintf("c-briggsae:block-%04d:line-%04d:%s", block+1, start+1, strings.ToLower(safeKeyRE.ReplaceAllString(symbol, "-")))
		out = append(out, record{block + 1, start + 1, header, symbol, symbol, seq, key, status})
	}
	if len(out) != 23 {
		return nil, len(lines), fmt.Errorf("headers=%d want 23", len(out))
	}
	return out, len(lines), nil
}

// proteinOnCoordinateLine implements the layout of C.briggsae.doc only. Each
// exon line contains genomic coordinates and/or a parenthesized phase note
// around one or more literal uppercase residue runs. Continuation lines omit
// coordinates. Removing those two annotation forms leaves only residue tokens;
// mixed-case prose such as Frameshift is removed with its parentheses.
func proteinOnCoordinateLine(line string) string {
	line = parentheticalRE.ReplaceAllString(line, " ")
	line = coordinateRE.ReplaceAllString(line, " ")
	var b strings.Builder
	for _, token := range strings.Fields(line) {
		if upperResidues(token) {
			b.WriteString(token)
		}
	}
	return b.String()
}

func upperResidues(s string) bool {
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

func hash(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func writeCSV(path string, records []record, docHash, textHash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "normalized_file", "normalized_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range records {
		note := fmt.Sprintf("C. briggsae Word reconstruction block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", "Caenorhabditis briggsae", r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, docFile, docHash, textFile, textHash, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, lines int, docHash, textHash string) error {
	counts := map[string]int{}
	for _, r := range records {
		for _, status := range r.Status {
			counts[status]++
		}
	}
	statuses := make([]string, 0, len(counts))
	for status := range counts {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Caenorhabditis briggsae\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Text SHA-256: `%s`\n- Normalized lines: `%d`\n- Explicit CYP reconstruction blocks: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", docFile, sourceURL, docHash, textFile, textHash, lines, len(records), len(records))
	b.WriteString("## Resource-specific interpretation\n\nThe November 2003 document is explicitly an incomplete, not-yet-formally-named genome assembly and contains 23 physical CYP-labelled reconstruction blocks. Every block remains distinct in source order. Protein is read only from that block's uppercase residue runs after removing genomic coordinates and parenthesized exon phase or frameshift annotations. This also accepts the document's intentional coordinate-free continuation lines and space-separated residue fragments, but never introductory prose. Source pseudogenes, short solo exons, X runs and internal stops are retained exactly; only one explicit terminal stop is removed. No DNA is translated and no external sequence is substituted.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, status := range statuses {
		fmt.Fprintf(&b, "| %s | %d |\n", status, counts[status])
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
