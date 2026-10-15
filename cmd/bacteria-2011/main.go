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

const (
	sourceFile = "bacteria-bacterial.P450s.2011.txt"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacterial.P450s.2011.doc"
	sourceSHA  = "884f9404487d8e42c98d51e5770f146a8a8075b40d678258aa4a69d6aa49a262"
)

var (
	headerRE        = regexp.MustCompile(`^>\s*(CYP\S+)\s*(.*)$`)
	spRE            = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+sp\.\s*([A-Za-z0-9-]+)?`)
	binomialRE      = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([a-z][a-z.-]*)\b`)
	strainRE        = regexp.MustCompile(`\b([A-Z][a-z]{2,})\s+([A-Z][A-Za-z0-9-]+)\b`)
	coordRE         = regexp.MustCompile(`^\d+$`)
	inlineProteinRE = regexp.MustCompile(`[ACDEFGHIKLMNPQRSTVWYBXZJUO*? -]{30,}`)
)

type record struct {
	Block, Line int
	Symbol      string
	Species     string
	Header      string
	Sequence    string
	Status      []string
}

func speciesFromHeader(header string) string {
	lower := strings.ToLower(header)
	if strings.HasPrefix(strings.TrimSpace(header), "SRE9564") {
		return "Thermobifida fusca"
	}
	switch {
	case strings.Contains(lower, "uncultured soil bacterium"):
		return "uncultured soil bacterium"
	case strings.Contains(lower, "uncultured haloarchaeon"):
		return "uncultured haloarchaeon FLAS10H9"
	case strings.Contains(lower, "seawater sequencing"):
		return "seawater environmental sample"
	case strings.Contains(lower, "marine metagenome"):
		return "marine metagenome"
	case strings.TrimSpace(header) == "":
		return ""
	case strings.Contains(header, "SP0881"):
		return "Streptomyces peucetius"
	case regexp.MustCompile(`\bsce\d+\b`).MatchString(header):
		return "Sorangium cellulosum"
	}
	if m := spRE.FindStringSubmatch(header); len(m) == 3 {
		name := m[1] + " sp."
		if m[2] != "" && len(m[2]) <= 20 {
			name += " " + m[2]
		}
		return normalizeSpecies(name)
	}
	m := binomialRE.FindStringSubmatch(header)
	if len(m) != 3 {
		if strain := strainRE.FindStringSubmatch(header); len(strain) == 3 {
			return strain[1] + " sp. " + strain[2]
		}
		return ""
	}
	return normalizeSpecies(m[1] + " " + strings.TrimRight(m[2], ".-"))
}

func normalizeSpecies(s string) string {
	replacer := strings.NewReplacer("Steptomyces ", "Streptomyces ", "Streptomcyes ", "Streptomyces ", "Micromosospora ", "Micromonospora ", "Citrabacter ", "Citrobacter ")
	return replacer.Replace(strings.TrimSpace(s))
}

func proteinLine(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" || strings.IndexFunc(s, func(r rune) bool { return r >= 'a' && r <= 'z' }) >= 0 {
		return ""
	}
	var b strings.Builder
	for _, field := range strings.Fields(s) {
		field = strings.Trim(field, `"`)
		if coordRE.MatchString(field) || field == "(0)" || field == "(1)" || field == "(2)" {
			continue
		}
		for _, r := range field {
			if !strings.ContainsRune("ACDEFGHIKLMNPQRSTVWYBXZJUO*-?", r) {
				return ""
			}
		}
		b.WriteString(field)
	}
	if b.Len() < 10 {
		return ""
	}
	return b.String()
}

func extractProtein(lines []string) string {
	var best, current strings.Builder
	flush := func() {
		if current.Len() > best.Len() {
			best.Reset()
			best.WriteString(current.String())
		}
		current.Reset()
	}
	for _, line := range lines {
		part := proteinLine(line)
		if part == "" {
			flush()
			continue
		}
		current.WriteString(part)
	}
	flush()
	return strings.TrimRight(best.String(), "*")
}

func inlineProtein(s string) string {
	best := ""
	for _, match := range inlineProteinRE.FindAllString(s, -1) {
		candidate := strings.ReplaceAll(match, " ", "")
		candidate = strings.TrimRight(candidate, "*")
		if len(candidate) > len(best) {
			best = candidate
		}
	}
	return best
}

func parse(text string) ([]record, error) {
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	type start struct {
		index                int
		symbol, tail, header string
	}
	var starts []start
	for i, line := range lines {
		if m := headerRE.FindStringSubmatch(strings.TrimSpace(line)); len(m) == 3 {
			starts = append(starts, start{i, m[1], m[2], strings.TrimPrefix(strings.TrimSpace(line), ">")})
		}
	}
	if len(starts) != 1152 {
		return nil, fmt.Errorf("2011 source header count changed: got %d want 1152", len(starts))
	}
	out := make([]record, 0, len(starts))
	for i, st := range starts {
		end := len(lines)
		if i+1 < len(starts) {
			end = starts[i+1].index
		}
		species := speciesFromHeader(st.tail)
		if species == "" {
			contextEnd := st.index + 4
			if contextEnd > end {
				contextEnd = end
			}
			species = speciesFromHeader(strings.Join(lines[st.index+1:contextEnd], " "))
		}
		if species == "" {
			return nil, fmt.Errorf("block %d line %d has no reviewed species: %s", i+1, st.index+1, st.header)
		}
		seq := extractProtein(lines[st.index+1 : end])
		if inline := inlineProtein(st.tail); len(inline) > len(seq) {
			seq = inline
		}
		status := []string{"reviewed-bacteria-2011"}
		if seq == "" {
			status = append(status, "sequence-missing")
		} else {
			status = append(status, "literal-protein")
		}
		low := strings.ToLower(strings.Join(lines[st.index:end], " "))
		for _, marker := range []string{"fragment", "partial", "pseudogene", "variant", "renamed", "duplicate"} {
			if strings.Contains(low, marker) {
				status = append(status, marker)
			}
		}
		out = append(out, record{i + 1, st.index + 1, st.symbol, species, st.header, seq, status})
	}
	return out, nil
}

func main() {
	in := flag.String("input", filepath.Join("raw", sourceFile), "Word-normalized 2011 bacterial P450 source")
	out := flag.String("out", filepath.Join("sources", "reviewed", "bacteria", "bacteria-2011.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "bacteria", "bacteria-2011.md"), "resource audit")
	flag.Parse()
	b, err := os.ReadFile(*in)
	if err != nil {
		panic(err)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if h != sourceSHA {
		panic(fmt.Sprintf("source SHA-256 changed: got %s want %s", h, sourceSHA))
	}
	records, err := parse(string(b))
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
	withSequence := 0
	species := map[string]bool{}
	for _, r := range records {
		if r.Sequence != "" {
			withSequence++
		}
		species[r.Species] = true
		note := fmt.Sprintf("2011 bacterial/archaeal Word release block %d at normalized line %d; complete header=%s", r.Block, r.Line, r.Header)
		_ = w.Write([]string{"bacteria", r.Species, r.Symbol, r.Symbol, fmt.Sprintf("bacteria-2011:block-%04d:line-%05d:%s", r.Block, r.Line, strings.ToLower(r.Symbol)), sourceURL, note, r.Sequence, "bacteria-bacterial.P450s.2011.doc", "de61b27e4183d7fb77db92d01c9c1a947066ab1d2df54829b110b4321e3565f1", strconv.Itoa(r.Block), strconv.Itoa(r.Line), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Dir(*audit), 0o755); err != nil {
		panic(err)
	}
	body := fmt.Sprintf("# Bacteria resource review: 2011 public bacterial and archaeal P450 collection\n\n- Source file: `bacteria-bacterial.P450s.2011.doc`\n- URL: %s\n- DOC SHA-256: `de61b27e4183d7fb77db92d01c9c1a947066ab1d2df54829b110b4321e3565f1`\n- Normalized text SHA-256: `%s`\n- Word layout: 10,262 paragraphs, 184 pages, no tables\n- Explicit CYP header blocks retained: `%d`\n- Blocks with literal protein: `%d`\n- Distinct source organism labels: `%d`\n- Review status: `complete`\n\nThe title records 1,200 bacterial sequences with 95 confidential entries removed and 27 archaeal entries included. The public document physically contains 1,152 explicit CYP-labelled blocks. They are retained in source order, including variants, renamed entries, pseudogenes, fragments and repeated names; the historical headline count is not used to discard source blocks. Organism labels come from each block header, with eleven explicitly reviewed exceptional headers (environmental samples, abbreviated Sorangium identifiers, and SP0881). Protein is accepted only from the longest uninterrupted uppercase residue run inside that exact block after removing Word wrapping, standalone residue numbering and phase markers. Annotation, similarity prose and accessions interrupt a run. Only terminal stop markers are removed; internal stops, X, gaps and question marks remain literal. No sequence is translated, repaired, deduplicated or fetched elsewhere.\n", sourceURL, sourceSHA, len(records), withSequence, len(species))
	if err := os.WriteFile(*audit, []byte(body), 0o644); err != nil {
		panic(err)
	}
	fmt.Printf("bacteria 2011: %d records, %d sequences, %d organism labels\n", len(records), withSequence, len(species))
}
