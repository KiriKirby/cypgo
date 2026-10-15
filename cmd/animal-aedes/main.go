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

const namedDocFile = "animals-AedesFASTA.June25.doc"
const namedTextFile = "animals-AedesFASTA.June25.txt"
const archiveDocFile = "animals-aedesaegypti.doc"
const archiveTextFile = "animals-aedesaegypti.txt"
const namedURL = "https://drnelson.uthsc.edu/AedesFasta.June25.htm"
const archiveURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/aedesaegypti.doc"

var aedesParenRE = regexp.MustCompile(`\([^)]*\)`)
var aedesNumberRE = regexp.MustCompile(`\d+`)
var aedesKeyRE = regexp.MustCompile(`[^A-Za-z0-9]+`)

type record struct {
	Block, Line                       int
	Header, Symbol, ID, Sequence, Key string
	Status                            []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", namedTextFile), "")
	doc := flag.String("doc", filepath.Join("raw", namedDocFile), "")
	archiveDoc := flag.String("archive-doc", filepath.Join("raw", archiveDocFile), "")
	archiveText := flag.String("archive-text", filepath.Join("raw", archiveTextFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "aedes-aegypti.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "aedes-aegypti.md"), "")
	flag.Parse()
	rs, lines, err := parse(*in)
	if err != nil {
		panic(err)
	}
	namedDocHash, err := hash(*doc)
	if err != nil {
		panic(err)
	}
	namedTextHash, err := hash(*in)
	if err != nil {
		panic(err)
	}
	archiveDocHash, err := hash(*archiveDoc)
	if err != nil {
		panic(err)
	}
	archiveTextHash, err := hash(*archiveText)
	if err != nil {
		panic(err)
	}
	if err = writeCSV(*out, rs, namedDocHash, namedTextHash); err != nil {
		panic(err)
	}
	if err = writeAudit(*audit, rs, lines, namedDocHash, namedTextHash, archiveDocHash, archiveTextHash); err != nil {
		panic(err)
	}
	fmt.Printf("Aedes aegypti: %d named records, all literal protein\n", len(rs))
}

func parse(path string) ([]record, int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
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
		header := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[a]), ">"))
		fields := strings.Fields(header)
		if len(fields) == 0 || !strings.HasPrefix(strings.ToUpper(fields[0]), "CYP") {
			return nil, len(lines), fmt.Errorf("block %d bad header %q", bi+1, header)
		}
		symbol := fields[0]
		var seq strings.Builder
		for _, line := range lines[a+1 : z] {
			fragment := aedesProteinLine(line)
			seq.WriteString(fragment)
			if strings.HasSuffix(fragment, "*") {
				break
			}
		}
		s := seq.String()
		if s == "" {
			return nil, len(lines), fmt.Errorf("block %d %s has no protein", bi+1, symbol)
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
		if strings.Contains(s, "-") {
			status = append(status, "source-gap-retained")
		}
		lower := strings.ToLower(header)
		for _, m := range []struct{ k, v string }{{"partial", "source-partial"}, {"pseudogene", "source-pseudogene"}, {"allele", "source-allele"}, {"missing", "source-missing-exon"}, {"frameshift", "source-frameshift"}} {
			{
				if strings.Contains(lower, m.k) {
					status = append(status, m.v)
				}
			}
		}
		if len(s) < 350 {
			status = append(status, "short-source-fragment")
		}
		key := fmt.Sprintf("aedes-aegypti:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(aedesKeyRE.ReplaceAllString(symbol, "-")))
		out = append(out, record{bi + 1, a + 1, header, symbol, symbol, s, key, status})
	}
	if len(out) != 167 {
		return nil, len(lines), fmt.Errorf("headers=%d want167", len(out))
	}
	return out, len(lines), nil
}

func aedesProteinLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ".")
	line = aedesParenRE.ReplaceAllString(line, " ")
	line = aedesNumberRE.ReplaceAllString(line, " ")
	var b strings.Builder
	for _, token := range strings.Fields(line) {
		if !aedesResidues(token) {
			return ""
		}
		b.WriteString(token)
	}
	return b.String()
}
func aedesResidues(s string) bool {
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
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func writeCSV(path string, rs []record, docHash, textHash string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
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
		note := fmt.Sprintf("Aedes named FASTA Word block %d at normalized line %d; exact header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"animals", "Aedes aegypti", r.Symbol, r.ID, r.Key, namedURL, note, r.Sequence, namedDocFile, docHash, namedTextFile, textHash, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, rs []record, lines int, ndh, nth, adh, ath string) error {
	counts := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Status {
			counts[s]++
		}
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Aedes aegypti\n\n- Authoritative named collection: `%s`\n- Historical URL (redirects to the named DOC): %s\n- Named DOC SHA-256: `%s`\n- Named normalized text: `%s`\n- Named text SHA-256: `%s`\n- Named text lines: `%d`\n- Archival reconstruction document: `%s`\n- Archival URL: %s\n- Archival DOC SHA-256: `%s`\n- Archival text SHA-256: `%s`\n- Explicit named blocks: `%d`\n- Literal proteins: `%d`\n- Missing sequences: `0`\n- Review status: `complete`\n\n", namedDocFile, namedURL, ndh, namedTextFile, nth, lines, archiveDocFile, archiveURL, adh, ath, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe archival document explicitly directs readers to the 25 June 2006 completed-and-named FASTA collection and describes itself as detailed archival material. The named collection is therefore the release input; the 206 unfinished archival candidates, BLAST Query/Sbjct alignments and assembly evidence are not added again. Although its title says 164 sequences (11 CYP2 + 84 CYP3 + 59 CYP4 + 10 mitochondrial), the named document physically contains 167 explicit `>` blocks. All physical named blocks remain source-distinct, including versions, alleles and partials, instead of silently discarding three records to match the prose total. Only literal uppercase protein lines are joined after removing coordinates and intron-phase annotations. Mixed prose/alignment lines are rejected. Terminal stops are removed; internal stops, X and source gaps remain. No DNA is translated and no external sequence is substituted.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "| %s | %d |\n", k, counts[k])
	}
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
