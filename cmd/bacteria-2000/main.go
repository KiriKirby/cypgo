package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/csv"
	"errors"
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
	sourceFile    = "bacteria-bacteria.2000.txt"
	sourceDocFile = "bacteria-bacteria.2000.doc"
	sourceURL     = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacteria.2000.doc"
	sourceSHA     = "0ef41b98f5a1bc779a186d9c604c290eb7b93cb9d71778cf040a2ebb5b308ade"
	sourceDocSHA  = "2051abd667c3612d7b1ff7dd3e0a1b087bef75225e2c09fae5435f38c853050b"
	wantBlocks    = 157
)

type record struct {
	Block, Start, End                     int
	Symbol, ID, Header, Species, Sequence string
	Status                                []string
}

var headerRE = regexp.MustCompile(`^>\s*(\S+)(?:\s+(.*))?$`)
var binomialRE = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+((?:[Ss]p\.)|(?:[a-z][a-z0-9-]+))`)
var accessionRE = regexp.MustCompile(`^[A-Z]{1,4}\d{4,}(?:\.\d+)?$`)
var sequenceAlphabet = "ACDEFGHIKLMNOPQRSTVWYBXZJUO*-"

func decodeUTF16LE(data []byte) (string, error) {
	if len(data) < 2 || data[0] != 0xff || data[1] != 0xfe {
		return "", errors.New("source is not UTF-16LE with BOM")
	}
	if (len(data)-2)%2 != 0 {
		return "", errors.New("odd UTF-16LE payload")
	}
	units := make([]uint16, (len(data)-2)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(data[2+i*2 : 4+i*2])
	}
	return string(utf16.Decode(units)), nil
}

func readSource(path string) (string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	got := fmt.Sprintf("%x", sha256.Sum256(b))
	if got != sourceSHA {
		return "", got, fmt.Errorf("source SHA-256 changed: got %s want %s", got, sourceSHA)
	}
	text, err := decodeUTF16LE(b)
	if err != nil {
		return "", got, err
	}
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), got, nil
}

func verifiedHash(path, want string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	got := fmt.Sprintf("%x", sha256.Sum256(b))
	if got != want {
		return got, fmt.Errorf("source SHA-256 changed for %s: got %s want %s", path, got, want)
	}
	return got, nil
}

func speciesFromHeader(header string) string {
	// This resource contains many accession-only headers. We deliberately do not
	// infer an organism from an accession or CYP family; unstated organisms stay
	// under an explicit reviewed label.
	m := binomialRE.FindStringSubmatch(header)
	if len(m) == 3 {
		if m[1] == "Sanger" && m[2] == "cosmid" {
			return "Bacterial collection (organism not stated)"
		}
		species := m[1] + " " + m[2]
		if strings.EqualFold(m[2], "sp.") {
			rest := strings.TrimSpace(header[strings.Index(header, m[0])+len(m[0]):])
			if fields := strings.Fields(rest); len(fields) > 0 {
				s := strings.Trim(fields[0], ",;()[]")
				if s != "" && !strings.ContainsAny(s, "%<>") && !strings.HasPrefix(s, "CYP") {
					species += " " + s
				}
			}
		}
		return species
	}
	return "Bacterial collection (organism not stated)"
}

func statusFor(blockText, seq string, missing bool, multiple bool) []string {
	status := []string{"reviewed-bacteria-2000"}
	low := strings.ToLower(blockText)
	for _, marker := range []struct{ token, label string }{
		{"fragment", "source-fragment"}, {"partial", "source-partial"}, {"pseudogene", "source-pseudogene"},
		{"confidential", "source-confidential"}, {"same gene", "same-gene-alternate"}, {"duplicate", "source-duplicate"},
		{"n-terminal", "n-terminal-fragment"}, {"c-term", "c-terminal-fragment"},
	} {
		if strings.Contains(low, marker.token) {
			status = append(status, marker.label)
		}
	}
	if missing {
		status = append(status, "sequence-missing")
	} else {
		status = append(status, "literal-protein")
	}
	if multiple {
		status = append(status, "multiple-sequence-runs-unmerged")
	}
	if strings.Contains(seq, "X") {
		status = append(status, "ambiguous-X")
	}
	if strings.Contains(seq, "-") {
		status = append(status, "source-gap")
	}
	if strings.Contains(seq, "*") {
		status = append(status, "internal-stop")
	}
	if len(seq) > 0 && len(seq) < 300 {
		status = append(status, "short-fragment")
	}
	return status
}

func residueLine(line string) (string, bool) {
	s := strings.TrimSpace(line)
	if s == "" {
		return "", false
	}
	// Word occasionally leaves a terminal quote after the sequence.
	s = strings.TrimSuffix(s, "\"")
	// The L76374 block is printed in reverse numbered chunks: 483 SEQ 304.
	fields := strings.Fields(s)
	if len(fields) >= 3 && isDigits(fields[0]) && isDigits(fields[len(fields)-1]) {
		fields = fields[1 : len(fields)-1]
		s = strings.Join(fields, "")
	} else {
		// CYP118 contains an annotation after a literal fragment line. Keep only
		// the residue prefix; the annotation starts at the explicit parenthesis.
		if idx := strings.Index(s, "("); idx >= 0 {
			s = strings.TrimSpace(s[:idx])
		}
		s = strings.Join(strings.Fields(s), "")
	}
	if s == "" {
		return "", false
	}
	for _, r := range s {
		if !strings.ContainsRune(sequenceAlphabet, r) {
			return "", false
		}
	}
	if len(s) < 3 {
		return "", false
	}
	return s, true
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func extractSequence(lines []string, symbol string) (string, bool, bool, error) {
	type run struct {
		seq        string
		start, end int
	}
	var runs []run
	current := strings.Builder{}
	start := -1
	end := -1
	flush := func() {
		if current.Len() > 0 {
			runs = append(runs, run{current.String(), start, end})
			current.Reset()
			start = -1
			end = -1
		}
	}
	for i, line := range lines {
		if part, ok := residueLine(line); ok {
			if start < 0 {
				start = i
			}
			end = i
			current.WriteString(part)
		} else {
			flush()
		}
	}
	flush()
	if len(runs) == 0 {
		return "", true, false, nil
	}
	// CYP118 explicitly contains two separately annotated fragments. Preserve
	// the first source fragment only; never concatenate across prose.
	if symbol == "CYP118" {
		return strings.TrimRight(runs[0].seq, "*"), false, len(runs) > 1, nil
	}
	if len(runs) != 1 {
		return "", false, true, fmt.Errorf("unexpected %d separate residue runs", len(runs))
	}
	seq := runs[0].seq
	seq = strings.TrimRight(seq, "*")
	for _, r := range seq {
		if !strings.ContainsRune(sequenceAlphabet, r) {
			return "", false, false, fmt.Errorf("unexpected residue %q", r)
		}
	}
	return seq, false, len(runs) > 1, nil
}

func parse(text string) ([]record, error) {
	lines := strings.Split(text, "\n")
	type start struct {
		idx            int
		symbol, header string
	}
	starts := make([]start, 0, wantBlocks)
	for i, line := range lines {
		if m := headerRE.FindStringSubmatch(strings.TrimSpace(line)); len(m) > 0 {
			starts = append(starts, start{i, m[1], strings.TrimPrefix(strings.TrimSpace(line), ">")})
		}
	}
	if len(starts) != wantBlocks {
		return nil, fmt.Errorf("header block count changed: got %d want %d", len(starts), wantBlocks)
	}
	out := make([]record, 0, len(starts))
	for i, st := range starts {
		end := len(lines)
		if i+1 < len(starts) {
			end = starts[i+1].idx
		}
		seq, missing, multiple, err := extractSequence(lines[st.idx+1:end], st.symbol)
		if err != nil {
			return nil, fmt.Errorf("block %d line %d: %w", i+1, st.idx+1, err)
		}
		// For an empty block, missing is true; nonempty blocks retain literal text.
		if seq == "" {
			missing = true
		}
		blockText := strings.Join(lines[st.idx:end], "\n")
		status := statusFor(blockText, seq, missing, multiple)
		id := st.symbol
		if fields := strings.Fields(st.header); len(fields) > 1 && accessionRE.MatchString(strings.Trim(fields[1], ",;")) {
			id = strings.Trim(fields[1], ",;")
		}
		out = append(out, record{Block: i + 1, Start: st.idx + 1, End: end, Symbol: st.symbol, ID: id, Header: st.header, Species: speciesFromHeader(st.header), Sequence: seq, Status: status})
	}
	return out, nil
}

func writeCSV(path string, records []record, docHash string) error {
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
		note := fmt.Sprintf("bacteria 2000 Word block %d at normalized line %d-%d; complete source header=%s; literal sequence runs are not joined across annotations", r.Block, r.Start, r.End, r.Header)
		key := fmt.Sprintf("bacteria-2000:block-%04d:line-%05d:%s", r.Block, r.Start, strings.ToLower(r.Symbol))
		_ = w.Write([]string{"bacteria", r.Species, r.Symbol, r.ID, key, sourceURL, note, r.Sequence, sourceDocFile, docHash, strconv.Itoa(r.Block), strconv.Itoa(r.Start), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, textHash, docHash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	withSeq := 0
	missing := 0
	species := map[string]bool{}
	counts := map[string]int{}
	for _, r := range records {
		species[r.Species] = true
		if r.Sequence != "" {
			withSeq++
		}
		for _, s := range r.Status {
			counts[s]++
		}
		if r.Sequence == "" {
			missing++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Bacteria resource review: over 130 bacterial cytochrome P450 sequences (July 27, 2000)\n\n- Source document: `%s`\n- URL: %s\n- DOC SHA-256: `%s`\n- Reviewed normalized text: `%s`\n- Normalized text SHA-256: `%s`\n- Container: UTF-16LE Word-normalized text, 1,563 logical lines\n- Explicit `>` blocks retained: `%d`\n- Blocks with literal protein sequence: `%d`\n- Missing-sequence blocks: `%d`\n- Distinct organism labels including the explicit unstated fallback: `%d`\n- Review status: `complete`\n\n", sourceDocFile, sourceURL, docHash, sourceFile, textHash, len(records), withSeq, missing, len(species))
	b.WriteString("## Resource-specific interpretation\n\nThe source is parsed as 157 complete logical blocks beginning at a literal `>` line. Every block is retained in source order, including accession-only records, repeated CYP symbols, aliases, fragments, partials, confidential entries, and the final CYP? record. The first whitespace token is the source symbol; accession tokens are retained as IDs when explicitly present. Organism labels are copied only from an explicit binomial or `sp.` phrase in the same header; accession/CYP inference is not used, and unstated organisms use the explicit `Bacterial collection (organism not stated)` label.\n\nProtein is accepted only from uppercase residue lines inside that exact block. Word wrapping and surrounding spaces are removed; terminal quote characters and explicit terminal `*` markers are removed, while internal `*`, X, gaps, and unusual residues are retained. The numbered reverse-orientation L76374 block removes only its printed end coordinates. CYP118 has two separately annotated fragments; only its first contiguous fragment is stored and the later fragment is not concatenated across prose. Empty blocks remain records with an empty sequence. No sequence is translated, repaired, deduplicated, or fetched externally.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	b.WriteString("\n## Representative and exceptional blocks\n\n| Block | Line | Symbol | Species | Length | Status |\n|---:|---:|---|---|---:|---|\n")
	wanted := map[int]bool{1: true, 31: true, 45: true, 69: true, 73: true, 85: true, 93: true, 118: true, 144: true, 152: true, 157: true}
	for _, r := range records {
		if wanted[r.Block] {
			fmt.Fprintf(&b, "| %d | %d | %s | %s | %d aa | %s |\n", r.Block, r.Start, strings.ReplaceAll(r.Symbol, "|", "\\|"), strings.ReplaceAll(r.Species, "|", "\\|"), len(r.Sequence), strings.Join(r.Status, ";"))
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "UTF-16LE normalized source")
	doc := flag.String("doc", filepath.Join("raw", sourceDocFile), "original Word source")
	out := flag.String("out", filepath.Join("sources", "reviewed", "bacteria", "bacteria-2000.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "resources", "bacteria-over-130-bacterial-cytochrome-p450-sequences-july-27-2000.md"), "resource audit")
	flag.Parse()
	text, hash, err := readSource(*in)
	if err != nil {
		panic(err)
	}
	records, err := parse(text)
	if err != nil {
		panic(err)
	}
	docHash, err := verifiedHash(*doc, sourceDocSHA)
	if err != nil {
		panic(err)
	}
	if err := writeCSV(*out, records, docHash); err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, records, hash, docHash); err != nil {
		panic(err)
	}
	with := 0
	for _, r := range records {
		if r.Sequence != "" {
			with++
		}
	}
	fmt.Printf("bacteria 2000: %d records, %d literal sequences\n", len(records), with)
}
