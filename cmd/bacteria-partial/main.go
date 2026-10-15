package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
)

const (
	sourceFile = "bacteria-bacterial.P450s.txt"
	docFile    = "bacteria-bacterial.P450s.doc"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacterial.P450s.doc"
	textSHA    = "fd803abf2ae94cef5667156fc154d7abf7c01bf2f5cc8fdf40768be572490b25"
	docSHA     = "2b174b909711d87a3bb30177cd7bdcb74ca4ee737ad8e3cc12fd2f61e380d0bf"
)

type spec struct {
	line, end, seqStart, seqEnd int
	symbol, id, species         string
	status                      []string
}

type record struct {
	Block    int
	Spec     spec
	Sequence string
}

var alphabet = "ACDEFGHIKLMNPQRSTVWYBXZJUO*-"

func verify(path, want string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	got := fmt.Sprintf("%x", sha256.Sum256(b))
	if got != want {
		return nil, fmt.Errorf("hash changed for %s: got %s want %s", path, got, want)
	}
	return b, nil
}

func linesOf(text string) []string {
	return strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
}

func decodeSource(b []byte) (string, error) {
	if len(b) < 2 || b[0] != 0xff || b[1] != 0xfe || (len(b)-2)%2 != 0 {
		return "", fmt.Errorf("expected UTF-16LE source")
	}
	u := make([]uint16, (len(b)-2)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[2+i*2 : 4+i*2])
	}
	return string(utf16.Decode(u)), nil
}

func sequence(lines []string, start, end int) (string, error) {
	if start < 1 || end > len(lines) || start > end {
		return "", fmt.Errorf("invalid sequence range %d-%d", start, end)
	}
	var b strings.Builder
	for i := start; i <= end; i++ {
		raw := strings.TrimSuffix(strings.TrimSpace(lines[i-1]), "\"")
		// The Word source places a similarity comment after the literal
		// residues on line 97 and after the first CYP118 fragment on line 322.
		// These exact inline annotations terminate, rather than extend, sequence.
		for _, marker := range []string{" 91%", " (THIS"} {
			if idx := strings.Index(raw, marker); idx >= 0 {
				raw = raw[:idx]
				break
			}
		}
		s := strings.Join(strings.Fields(raw), "")
		if s == "" {
			continue
		}
		for _, aa := range s {
			if !strings.ContainsRune(alphabet, aa) {
				return "", fmt.Errorf("line %d unexpected residue/annotation %q", i, aa)
			}
		}
		b.WriteString(s)
	}
	return strings.TrimRight(b.String(), "*"), nil
}

func specs() []spec {
	return []spec{
		{7, 16, 9, 16, "CYP127", "AE000101", "Rhizobium sp. NGR234", nil},
		{18, 22, 20, 22, "AF071144", "AF071144", "Streptomyces glaucescens", []string{"short-fragment"}},
		{24, 28, 26, 28, "AF071146", "AF071146", "Micromonospora inyoensis", []string{"short-fragment"}},
		{30, 34, 32, 34, "AF071147", "AF071147", "Streptoalloteichus hindustanus", []string{"short-fragment"}},
		{36, 40, 38, 40, "AF071148", "AF071148", "Amycolata autotrophica", []string{"short-fragment"}},
		{42, 57, 50, 57, "CYP116", "U17130", "Rhodococcus erythropolis", nil},
		{59, 68, 61, 68, "CYP109B1", "AB000735", "Nocardioides sp.", nil},
		{70, 80, 73, 80, "CYP103A2", "AF034769", "Agrobacterium tumefaciens", nil},
		{82, 90, 85, 90, "CYP165C2", "U84350", "Amycolatopsis orientalis", nil},
		{92, 103, 95, 103, "CYP165C3-gene1", "AJ223998-gene1", "Amycolatopsis orientalis", []string{"n-terminal-fragment"}},
		{92, 112, 105, 112, "CYP107H2-gene2", "AJ223998-gene2", "Amycolatopsis orientalis", nil},
		{114, 125, 118, 124, "CYP165A2-gene1", "AJ223999-gene1", "Amycolatopsis orientalis", nil},
		{114, 135, 128, 135, "CYP165B2-gene2", "AJ223999-gene2", "Amycolatopsis orientalis", nil},
		{114, 143, 140, 142, "CYP165C3-gene3", "AJ223999-gene3", "Amycolatopsis orientalis", []string{"n-terminal-fragment"}},
		{144, 166, 147, 156, "CYP107Q1-gene1", "AF040571-gene1", "Amycolatopsis mediterranei", nil},
		{144, 166, 159, 165, "CYP105C1-gene2", "AF040571-gene2", "Amycolatopsis mediterranei", []string{"short-fragment"}},
		{178, 190, 182, 189, "CYP105C1-gene1", "AF040570-gene1", "Amycolatopsis mediterranei", nil},
		{178, 201, 194, 201, "CYP105D1-gene2", "AF040570-gene2", "Amycolatopsis mediterranei", nil},
		{178, 213, 205, 212, "CYP107Q2-gene3", "AF040570-gene3", "Amycolatopsis mediterranei", nil},
		{214, 223, 216, 223, "CYP107M1", "D87924", "Actinomadura hibisca", nil},
		{225, 234, 227, 234, "CYP140", "Z97193", "Mycobacterium tuberculosis", nil},
		{236, 245, 238, 245, "CYP144", "Z97345", "Mycobacterium tuberculosis", nil},
		{247, 256, 249, 255, "CYP143", "AL022021", "Mycobacterium tuberculosis", nil},
		{257, 266, 259, 266, "CYP142", "AL022022", "Mycobacterium tuberculosis", nil},
		{268, 278, 270, 278, "CYP135B1", "AL021942", "Mycobacterium tuberculosis", nil},
		{280, 288, 282, 287, "CYP?", "AL008609", "Mycobacterium leprae", nil},
		{289, 301, 293, 301, "CYP136", "Z83866", "Mycobacterium tuberculosis", nil},
		{303, 315, 308, 315, "CYP?", "Z95150", "Mycobacterium tuberculosis", nil},
		{319, 326, 321, 322, "CYP118", "CYP118-fragment-1", "Bacillus subtilis", []string{"source-fragment", "ambiguous-X", "multiple-sequence-runs-unmerged", "short-fragment"}},
		{332, 342, 333, 341, "CYP102A2", "CYP102A2", "Bacillus subtilis", []string{"n-terminal-fragment"}},
	}
}

func parse(text string) ([]record, error) {
	lines := linesOf(text)
	out := make([]record, 0, len(specs()))
	for i, sp := range specs() {
		seq, err := sequence(lines, sp.seqStart, sp.seqEnd)
		if err != nil {
			return nil, fmt.Errorf("%s lines %d-%d: %w", sp.symbol, sp.seqStart, sp.seqEnd, err)
		}
		out = append(out, record{i + 1, sp, seq})
	}
	return out, nil
}

func writeCSV(path string, records []record, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range records {
		sp := r.Spec
		note := fmt.Sprintf("bacterial partial collection lines %d-%d; explicit source section; sequence lines %d-%d; BLAST Query/Sbjct and annotations excluded", sp.line, sp.end, sp.seqStart, sp.seqEnd)
		key := fmt.Sprintf("bacteria-partial:block-%02d:line-%04d:%s", r.Block, sp.line, strings.ToLower(sp.id))
		_ = w.Write([]string{"bacteria", sp.species, sp.symbol, sp.id, key, sourceURL, note, r.Sequence, docFile, hash, strconv.Itoa(r.Block), strconv.Itoa(sp.line), fmt.Sprintf("lines %d-%d", sp.line, sp.end), strings.Join(append([]string{"reviewed-bacteria-partial", "literal-protein"}, sp.status...), ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, textHash, docHash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lengths := make([]int, len(records))
	for i, r := range records {
		lengths[i] = len(r.Sequence)
	}
	sort.Ints(lengths)
	var b strings.Builder
	fmt.Fprintf(&b, "# Bacteria resource review: Bacterial Cytochrome P450s (partial collection)\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Normalized text SHA-256: `%s`\n- Accepted independent source sections: `%d`\n- Accepted literal protein sequences: `%d`\n- Excluded BLAST Query/Sbjct alignment lines: 18 Query + 25 Sbjct lines\n- Excluded foreign fungal section: `X82490 Fusarium oxysporum`\n- Excluded BLAST-only `C. crescentus at TIGR` section: 4 HSPs\n- Review status: `complete`\n\n", docFile, sourceURL, docHash, sourceFile, textHash, len(records), len(records))
	b.WriteString("## Resource-specific interpretation\n\nThis UTF-16LE Word-normalized partial collection is not a FASTA file. Each accepted record is a manually delimited source section with an accession/header, optional gene annotation, and a contiguous literal protein range. The three-gene AJ223999, two-gene AJ223998, two-gene AF040571, and three-gene AF040570 sections are emitted as independent records with separate source IDs and line ranges. AJ223012 is a BLAST Sbjct alignment only and is excluded. The C. crescentus section contains only BLAST Query/Sbjct HSPs and is excluded. X82490 is explicitly a fungal P450 and is excluded from the bacterial category. CYP118 has separately annotated fragment pieces; only the first contiguous fragment is retained, with X markers preserved. Terminal quotes/stops are removed only at the sequence boundary; no alignment, translation, repair, deduplication, or external lookup is used.\n\n| Record | Symbol | Species | Length | Source lines | Status |\n|---:|---|---|---:|---|---|\n")
	for _, r := range records {
		fmt.Fprintf(&b, "| %d | %s | %s | %d aa | %d-%d | %s |\n", r.Block, r.Spec.symbol, r.Spec.species, len(r.Sequence), r.Spec.seqStart, r.Spec.seqEnd, strings.Join(r.Spec.status, ";"))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "normalized partial collection")
	doc := flag.String("doc", filepath.Join("raw", docFile), "original Word source")
	out := flag.String("out", filepath.Join("sources", "reviewed", "bacteria", "bacteria-partial.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "resources", "bacteria-cytochrome-p450s-partial-collection.md"), "resource audit")
	flag.Parse()
	b, err := verify(*in, textSHA)
	if err != nil {
		panic(err)
	}
	if _, err := verify(*doc, docSHA); err != nil {
		panic(err)
	}
	text, err := decodeSource(b)
	if err != nil {
		panic(err)
	}
	records, err := parse(text)
	if err != nil {
		panic(err)
	}
	if err := writeCSV(*out, records, docSHA); err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, records, textSHA, docSHA); err != nil {
		panic(err)
	}
	fmt.Printf("bacteria partial: %d records, %d literal sequences\n", len(records), len(records))
}
