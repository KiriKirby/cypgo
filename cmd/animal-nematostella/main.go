package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const docFile = "animals-nematostella.doc"
const textFile = "animals-nematostella.txt"
const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/nematostella.doc"

var headRE = regexp.MustCompile(`^\s*(\d+):\s*([A-Z]{1,4}_\d+(?:\.\d+)?)`)

type record struct {
	Block, Line               int
	Header, ID, Sequence, Key string
	Status                    []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "")
	doc := flag.String("doc", filepath.Join("raw", docFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "nematostella.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "nematostella.md"), "")
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
	fmt.Printf("Nematostella vectensis: %d literal proteins\n", len(rs))
}
func parse(p string) ([]record, int, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return nil, 0, e
	}
	ls := strings.Split(strings.ReplaceAll(strings.ReplaceAll(string(b), "\r\n", "\n"), "\r", "\n"), "\n")
	var st []int
	for i, l := range ls {
		if headRE.MatchString(l) {
			st = append(st, i)
		}
	}
	st = append(st, len(ls))
	var out []record
	for bi := 0; bi < len(st)-1; bi++ {
		a, z := st[bi], st[bi+1]
		h := strings.TrimSpace(ls[a])
		m := headRE.FindStringSubmatch(h)
		id := m[2]
		var sb strings.Builder
		for _, l := range ls[a+1 : z] {
			q := strings.Join(strings.Fields(l), "")
			if q == "" {
				continue
			}
			if !res(q) {
				continue
			}
			sb.WriteString(q)
		}
		s := sb.String()
		if s == "" {
			return nil, len(ls), fmt.Errorf("block %d no protein", bi+1)
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
		if len(s) < 350 {
			status = append(status, "short-source-fragment")
		}
		out = append(out, record{bi + 1, a + 1, h, id, s, fmt.Sprintf("nematostella:block-%04d:line-%04d:%s", bi+1, a+1, strings.ToLower(id)), status})
	}
	if len(out) != 91 {
		return nil, len(ls), fmt.Errorf("count=%d", len(out))
	}
	return out, len(ls), nil
}
func res(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != '*' && r != '-' {
			return false
		}
	}
	return s != ""
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
		_ = w.Write([]string{"animals", "Nematostella vectensis", r.ID, r.ID, r.Key, sourceURL, fmt.Sprintf("Nematostella numbered block %d line %d; exact header=%s", r.Block, r.Line, r.Header), r.Sequence, docFile, dh, textFile, th, strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(p string, rs []record, l int, dh, th string) error {
	body := fmt.Sprintf("# Animal resource review: Nematostella vectensis\n\n- File: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Text SHA-256: `%s`\n- Lines: `%d`\n- Numbered blocks: `%d`\n- Literal proteins: `%d`\n- Missing: `0`\n- Review: `complete`\n\nThis resource has no FASTA `>` headers. Its exact physical boundary is the 91 author-numbered `N: accession` blocks. Protein is joined only from all-uppercase residue lines inside each numbered block. The Word-to-text output concatenates each accession with `Nematostella`; the fixed resource profile extracts only the explicit accession grammar (`XM_` plus digits/version) and preserves the full physical header in provenance. Nothing is resolved externally. Only terminal stops are removed.\n", docFile, sourceURL, dh, th, l, len(rs), len(rs))
	if e := os.MkdirAll(filepath.Dir(p), 0o755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(body), 0o644)
}
