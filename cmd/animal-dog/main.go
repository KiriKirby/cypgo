package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const dogDoc = "animals-dogP450s.doc"
const dogFASTA = "animals-dog.dbs.20110817.html"
const dogDocURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/dogP450s.doc"
const dogFASTAURL = "http://drnelson.uthsc.edu/BLAST/dog.dbs.html"
const dogArchiveURL = "https://web.archive.org/web/20110817181055id_/http://drnelson.uthsc.edu/BLAST/dog.dbs.html"

var dogAccRE = regexp.MustCompile(`\b(?:[A-Z]{1,4}_?\d{5,}(?:\.\d+)?|[A-Z]{2}\d{5,}(?:\.\d+)?)\b`)
var dogSymbolRE = regexp.MustCompile(`(?i)^\s*(CYP[^\s]+)`)

type dogRecord struct {
	Block, Line                             int
	Header, Symbol, ID, Sequence, RecordKey string
	Statuses                                []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", dogFASTA), "")
	doc := flag.String("doc", filepath.Join("raw", dogDoc), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "dog.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "dog.md"), "")
	flag.Parse()
	rs, lines, err := parseDog(*in)
	if err != nil {
		panic(err)
	}
	dh, err := fileHash(*doc)
	if err != nil {
		panic(err)
	}
	fh, err := fileHash(*in)
	if err != nil {
		panic(err)
	}
	if err = writeDogCSV(*out, rs, dh, fh); err != nil {
		panic(err)
	}
	if err = writeDogAudit(*audit, rs, lines, dh, fh); err != nil {
		panic(err)
	}
	fmt.Printf("Canis familiaris: %d reviewed FASTA records, all with literal protein\n", len(rs))
}

func parseDog(path string) ([]dogRecord, int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	text := strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n")
	text = html.UnescapeString(text)
	lines := strings.Split(text, "\n")
	var starts []int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			starts = append(starts, i)
		}
	}
	starts = append(starts, len(lines))
	var out []dogRecord
	for bi := 0; bi < len(starts)-1; bi++ {
		a, z := starts[bi], starts[bi+1]
		header := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[a]), ">"))
		m := dogSymbolRE.FindStringSubmatch(header)
		if len(m) != 2 {
			return nil, len(lines), fmt.Errorf("block %d has no CYP symbol: %q", bi+1, header)
		}
		symbol := m[1]
		var seq strings.Builder
		for _, line := range lines[a+1 : z] {
			q := strings.TrimSpace(strings.TrimSuffix(line, "</pre>"))
			if q == "" {
				continue
			}
			for _, r := range q {
				if unicode.IsLetter(r) || r == '*' || r == '-' {
					seq.WriteRune(r)
				}
			}
		}
		s := seq.String()
		if s == "" {
			return nil, len(lines), fmt.Errorf("block %d has no literal protein", bi+1)
		}
		var statuses []string
		if strings.HasSuffix(s, "*") {
			s = strings.TrimSuffix(s, "*")
			statuses = append(statuses, "terminal-stop-removed")
		}
		if strings.Contains(s, "*") {
			statuses = append(statuses, "internal-stop-retained")
		}
		if strings.IndexFunc(s, unicode.IsLower) >= 0 {
			statuses = append(statuses, "source-lowercase-retained")
		}
		if strings.ContainsAny(s, "Xx") {
			statuses = append(statuses, "ambiguous-X-retained")
		}
		if len(s) < 350 {
			statuses = append(statuses, "short-source-fragment")
		}
		id := symbol
		if acc := dogAccRE.FindString(header); acc != "" {
			id = acc
		}
		keySymbol := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(symbol, "-"))
		out = append(out, dogRecord{Block: bi + 1, Line: a + 1, Header: header, Symbol: symbol, ID: id, Sequence: s,
			RecordKey: fmt.Sprintf("dog:archived-fasta:block-%04d:line-%04d:%s", bi+1, a+1, keySymbol), Statuses: statuses})
	}
	if len(out) != 56 {
		return nil, len(lines), fmt.Errorf("dog FASTA records=%d want 56", len(out))
	}
	return out, len(lines), nil
}

func fileHash(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func writeDogCSV(path string, rs []dogRecord, docHash, fastaHash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "companion_file", "companion_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Dog newest-collection FASTA block %d at archived source line %d; exact header=%s; live historical FASTA URL=%s; naming/annotation companion=%s", r.Block, r.Line, r.Header, dogFASTAURL, dogDocURL)
		_ = w.Write([]string{"animals", "Canis familiaris", r.Symbol, r.ID, r.RecordKey, dogArchiveURL, note, r.Sequence, dogFASTA, fastaHash, dogDoc, docHash, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeDogAudit(path string, rs []dogRecord, lines int, docHash, fastaHash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Statuses {
			counts[s]++
		}
	}
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Canis familiaris (dog)\n\n- Naming/annotation document: `%s`\n- Document URL: %s\n- DOC SHA-256: `%s`\n- Newest-collection FASTA live URL recorded in the Word hyperlink: %s\n- Archived capture used because the live path now returns 404: %s\n- Reviewed archived file: `%s`\n- FASTA SHA-256: `%s`\n- Archived source lines: `%d`\n- Explicit FASTA records: `%d`\n- Literal protein sequences: `%d`\n- Review status: `complete`\n\n", dogDoc, dogDocURL, docHash, dogFASTAURL, dogArchiveURL, dogFASTA, fastaHash, lines, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe Word document explicitly labels itself modified January 8, 2010 and links `Dog FASTA file` to `/BLAST/dog.dbs.html` as its newest collection. That link is now dead, but the Internet Archive preserves the exact 2011 capture. The archived FASTA is therefore the authoritative publishable row set. Its 56 explicit `>` blocks are parsed independently and kept in source order, including pseudogenes, internal-exon fragments, deletion fragments, internal stops, lowercase residues and ambiguous X/x. Only an explicit terminal `*` is removed. The older Word body was inspected as annotation/candidate material (101 headers, including family aggregation and multi-run blocks) and is retained as companion provenance; it is not appended as duplicate/intermediate records over the source-designated newest FASTA. No external sequence database is queried or used.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
