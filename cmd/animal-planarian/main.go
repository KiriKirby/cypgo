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

const docFile = "animals-Planarian.doc"
const textFile = "animals-Planarian.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Planarian.doc"

var planarianCoordinateRE = regexp.MustCompile(`\d+`)
var planarianParenRE = regexp.MustCompile(`\([^)]*\)`)
var planarianKeyRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

type record struct {
	Block, Line                                int
	Header, Species, Symbol, ID, Sequence, Key string
	Status                                     []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "planarian.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "planarian.md"), "")
	flag.Parse()
	records, lines, err := parse(*in)
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
	if err = writeCSV(*out, records, docHash, textHash); err != nil {
		panic(err)
	}
	if err = writeAudit(*audit, records, lines, docHash, textHash); err != nil {
		panic(err)
	}
	fmt.Printf("Planarian: %d records, all literal protein\n", len(records))
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
		id, symbol, species := planarianIdentity(header, block+1)
		var sequence strings.Builder
		for _, line := range lines[start+1 : end] {
			fragment := planarianProteinLine(line)
			sequence.WriteString(fragment)
			// In this Word resource an explicit stop at the end of a residue
			// line closes the physical protein reconstruction. Section labels
			// and prose may follow before the next > header and are not protein.
			if strings.HasSuffix(fragment, "*") {
				break
			}
		}
		seq := sequence.String()
		if seq == "" {
			return nil, len(lines), fmt.Errorf("block %d %q has no literal protein", block+1, header)
		}
		var status []string
		if strings.HasSuffix(seq, "*") {
			seq = strings.TrimSuffix(seq, "*")
			status = append(status, "terminal-stop-removed")
		}
		if strings.Contains(seq, "*") {
			status = append(status, "internal-stop-retained")
		}
		if strings.ContainsAny(seq, "Xx") {
			status = append(status, "ambiguous-X-or-x-retained")
		}
		if strings.Contains(seq, "-") {
			status = append(status, "source-gap-retained")
		}
		lower := strings.ToLower(header)
		for _, marker := range []struct{ name, status string }{
			{"pseudogene", "source-pseudogene"}, {"partial", "source-partial"}, {"allele", "source-allele"},
			{"duplicate", "source-duplicate"}, {"frameshift", "source-frameshift"}, {"runs off", "source-runs-off-end"},
		} {
			if strings.Contains(lower, marker.name) {
				status = append(status, marker.status)
			}
		}
		if len(seq) < 350 {
			status = append(status, "short-source-fragment")
		}
		key := fmt.Sprintf("planarian:block-%04d:line-%04d:%s", block+1, start+1, strings.ToLower(planarianKeyRE.ReplaceAllString(id, "-")))
		out = append(out, record{block + 1, start + 1, header, species, symbol, id, seq, key, status})
	}
	if len(out) != 69 {
		return nil, len(lines), fmt.Errorf("headers=%d want 69", len(out))
	}
	return out, len(lines), nil
}

func planarianIdentity(header string, block int) (id, symbol, species string) {
	species = "Schmidtea mediterranea"
	fields := strings.Fields(header)
	if len(fields) > 0 && (strings.HasPrefix(fields[0], "v31.") || strings.EqualFold(fields[0], "CYP20")) {
		return fields[0], fields[0], species
	}
	if strings.HasPrefix(strings.ToLower(header), "clone=") {
		id = "H.49.4f"
		return id, id, species
	}
	if strings.HasPrefix(header, "Dugesia ryukyuensis") {
		return "BW640569.1", "BW640569.1", "Dugesia ryukyuensis"
	}
	id = fmt.Sprintf("planarian-block-%04d", block)
	return id, strings.TrimSpace(header), species
}

// planarianProteinLine is intentionally specific to Planarian.doc. After its
// coordinate, exon-phase and prose parenthetical annotations are removed, the
// source protein is represented by uppercase residue tokens plus literal
// lowercase-x gap runs. Short residue tokens are retained because some exons
// are split into very short continuations; mixed-case prose and the lowercase
// genomic triplet "aga" in the CYP20 reconstruction are rejected.
func planarianProteinLine(line string) string {
	line = planarianParenRE.ReplaceAllString(line, " ")
	line = planarianCoordinateRE.ReplaceAllString(line, " ")
	line = strings.ReplaceAll(line, "&", " ")
	var b strings.Builder
	for _, token := range strings.Fields(line) {
		if token == "aga" { // literal genomic triplet embedded before CYP20 protein
			continue
		}
		if !planarianResidues(token) {
			return ""
		}
		b.WriteString(token)
	}
	return b.String()
}

func planarianResidues(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != 'x' && r != '*' && r != '-' {
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
		note := fmt.Sprintf("Planarian Word block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", r.Species, r.Symbol, r.ID, r.Key, sourceURL, note, r.Sequence, docFile, docHash, textFile, textHash, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, lines int, docHash, textHash string) error {
	counts := map[string]int{}
	species := map[string]int{}
	for _, r := range records {
		species[r.Species]++
		for _, s := range r.Status {
			counts[s]++
		}
	}
	statuses := make([]string, 0, len(counts))
	for s := range counts {
		statuses = append(statuses, s)
	}
	sort.Strings(statuses)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Planarian\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Text SHA-256: `%s`\n- Normalized lines: `%d`\n- Explicit source blocks: `%d`\n- Literal proteins: `%d`\n- Schmidtea mediterranea blocks: `%d`\n- Dugesia ryukyuensis blocks: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", docFile, sourceURL, docHash, textFile, textHash, lines, len(records), len(records), species["Schmidtea mediterranea"], species["Dugesia ryukyuensis"])
	b.WriteString("## Resource-specific interpretation\n\nThe 813-line Word document physically contains 69 explicit `>` blocks across its author-defined clades. They are retained in source order, including alleles, duplicate contigs, partials, pseudogenes and frameshift reconstructions. Sixty-eight blocks are Schmidtea mediterranea; the explicitly foreign Dugesia ryukyuensis EST is assigned to that species instead of being mislabeled. Coordinate and exon-phase annotations are removed, while literal uppercase protein runs, lowercase-x gap runs, X, internal stops and hyphens remain. Parenthesized prose, relationship lines, coordinates, the lowercase genomic `aga` token and the unidentified-sequence note are excluded. Only a terminal stop is removed. No DNA/EST is translated and no external sequence is substituted.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, s := range statuses {
		fmt.Fprintf(&b, "| %s | %d |\n", s, counts[s])
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
