// Resource-specific importer for David R. Nelson's historical biblioD P450
// resource. Each explicit CYP-at-line-start plus organism entry becomes a
// provenance-distinct record. Literal uppercase amino-acid runs are retained.
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
	"strings"
)

const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/biblioD.html"
const sourceSHA = "aac540d08652c2241b7c500f1b7e2c751ab94a8d5e98dd9e04a954cf48252c7b"

var (
	tagRE    = regexp.MustCompile(`(?s)<[^>]+>`)
	headerRE = regexp.MustCompile(`^\s*(CYP[0-9]{1,4}[A-Za-z0-9.\-?]*)\s{2,}(.+?)\s*$`)
	binomRE  = regexp.MustCompile(`^([A-Z][a-z]+)\s+([a-z][A-Za-z.-]+)\b`)
	hybridRE = regexp.MustCompile(`^([A-Z][a-z]+)\s+x\s+([a-z][A-Za-z.-]+)\b`)
	spRE     = regexp.MustCompile(`^([A-Z][a-z]+)\s+sp\.\s+([A-Za-z0-9.-]+)\b`)
	coordRE  = regexp.MustCompile(`^\d+\s+|\s+\d+$`)
	lowerRE  = regexp.MustCompile(`[a-z]`)
)

type entry struct {
	Line, Block int
	Name        string
	Species     string
	Category    string
	RawSpecies  string
	Text        string
	Sequence    string
	Status      string
}

func decodeWindows1252(b []byte) string {
	var out strings.Builder
	for _, c := range b {
		if c < 0x80 || c >= 0xa0 {
			out.WriteRune(rune(c))
			continue
		}
		const table = "€\u0081‚ƒ„…†‡ˆ‰Š‹Œ\u008dŽ\u008f\u0090‘’“”•–—˜™š›œ\u009džŸ"
		out.WriteRune([]rune(table)[int(c-0x80)])
	}
	return out.String()
}

func cleanLine(s string) string {
	return strings.TrimSpace(html.UnescapeString(tagRE.ReplaceAllString(s, "")))
}

func classifyOrganism(raw string) (string, string, bool) {
	s := strings.Join(strings.Fields(strings.ReplaceAll(cleanLine(raw), "_", " ")), " ")
	lower := strings.ToLower(s)
	for _, bad := range []string{"confidential", "discontinued,", "glyma02g31910.1", "os06g0570600", "this seq"} {
		if strings.HasPrefix(lower, bad) {
			return "", "", false
		}
	}
	for _, name := range []string{
		"Trypanosoma cruzi", "Leishmania donovani", "Leishmania infantum", "Leishmania major",
		"Monosiga ovata", "Euglena gracilis", "Ectocarpus siliculosis",
		"Phaeodactylum tricornutum", "Skeletonema costatum", "Thalassiosira pseudonana",
	} {
		if strings.HasPrefix(lower, strings.ToLower(name)) {
			return "protists", name, true
		}
	}
	aliases := []struct{ prefix, value string }{
		{"(gerbera hybrid)", "Gerbera hybrid"}, {"am455281.2 vitis vinifera", "Vitis vinifera"},
		{"fragaria x ananassa", "Fragaria x ananassa"}, {"malus x domestica x malus sieversii", "Malus x domestica x Malus sieversii"},
		{"malus x domestica", "Malus x domestica"}, {"citrus x paradisi x poncirus trifoliata", "Citrus x paradisi x Poncirus trifoliata"},
		{"citrus x paradisi", "Citrus x paradisi"}, {"mentha x gracilis", "Mentha x gracilis"},
		{"mentha x piperita", "Mentha x piperita"}, {"pelargonium x hortorum", "Pelargonium x hortorum"},
		{"petunia x hybrida", "Petunia x hybrida"}, {"populus x canescens", "Populus x canescens"},
		{"taxus x media", "Taxus x media"}, {"verbena x hybrida", "Verbena x hybrida"},
		{"pericallis cruenta", "Pericallis cruenta"}, {"potato", "potato"}, {"rice", "Oryza sativa"},
		{"tobacco", "Nicotiana tabacum"}, {"tomato", "tomato"}, {"oats", "oats"},
	}
	for _, a := range aliases {
		if strings.HasPrefix(lower, a.prefix) {
			return "plants", a.value, true
		}
	}
	if m := hybridRE.FindStringSubmatch(s); len(m) == 3 {
		return "plants", m[1] + " x " + m[2], true
	}
	if m := spRE.FindStringSubmatch(s); len(m) == 3 {
		return "plants", m[1] + " sp. " + m[2], true
	}
	if m := binomRE.FindStringSubmatch(s); len(m) == 3 {
		return "plants", m[1] + " " + m[2], true
	}
	for _, genus := range []string{"Amborella", "Arabidopsis", "Asparagus", "Citrus", "Petunia", "Selaginella", "Torenia", "Volvox"} {
		if strings.HasPrefix(s, genus) {
			return "plants", genus, true
		}
	}
	return "", "", false
}

func proteinLine(line string) string {
	s := strings.TrimSpace(cleanLine(line))
	if s == "" || lowerRE.MatchString(s) {
		return ""
	}
	for {
		next := coordRE.ReplaceAllString(s, "")
		if next == s {
			break
		}
		s = strings.TrimSpace(next)
	}
	s = strings.Join(strings.Fields(s), "")
	if len(s) < 10 {
		return ""
	}
	for _, r := range s {
		if !strings.ContainsRune("ACDEFGHIKLMNPQRSTVWYBXZJUO*-", r) {
			return ""
		}
	}
	return s
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
	seq := strings.TrimRight(best.String(), "*")
	if len(seq) < 30 {
		return ""
	}
	// ACGT-only runs are nucleotide, not protein.
	if strings.Trim(seq, "ACGTNX") == "" {
		return ""
	}
	return seq
}

func parse(data []byte) []entry {
	text := strings.ReplaceAll(strings.ReplaceAll(decodeWindows1252(data), "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	type start struct {
		line             int
		name, rawSpecies string
	}
	var starts []start
	for i, line := range lines {
		m := headerRE.FindStringSubmatch(cleanLine(line))
		if len(m) == 3 {
			starts = append(starts, start{i, m[1], m[2]})
		}
	}
	var out []entry
	for i, st := range starts {
		category, species, ok := classifyOrganism(st.rawSpecies)
		if !ok {
			continue
		}
		end := len(lines)
		if i+1 < len(starts) {
			end = starts[i+1].line
		}
		var block []string
		for _, line := range lines[st.line:end] {
			if clean := cleanLine(line); clean != "" {
				block = append(block, strings.Join(strings.Fields(clean), " "))
			}
		}
		sequence := extractProtein(lines[st.line+1 : end])
		status := "reviewed-auxiliary-bibliod;sequence-missing"
		if sequence != "" {
			status = "reviewed-auxiliary-bibliod;literal-protein"
			lowerBlock := strings.ToLower(strings.Join(block, " "))
			if strings.Contains(lowerBlock, "fragment") || strings.Contains(lowerBlock, "partial") || strings.Contains(lowerBlock, " est") {
				status += ";fragment-or-est"
			}
		}
		out = append(out, entry{Line: st.line + 1, Block: i + 1, Name: st.name, Category: category, Species: species, RawSpecies: cleanLine(st.rawSpecies), Text: strings.Join(block, " / "), Sequence: sequence, Status: status})
	}
	return out
}

func main() {
	in := flag.String("in", "raw/plants-biblioD.html", "downloaded biblioD HTML")
	out := flag.String("out", "sources/reviewed/plants/bibliod.csv", "reviewed CSV")
	flag.Parse()
	b, err := os.ReadFile(*in)
	if err != nil {
		panic(err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(b))
	if got != sourceSHA {
		panic(fmt.Sprintf("source SHA-256 changed: got %s want %s", got, sourceSHA))
	}
	entries := parse(b)
	if len(entries) != 7134 {
		panic(fmt.Sprintf("unexpected accepted biblioD entry count: %d", len(entries)))
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "review_status"})
	for _, e := range entries {
		note := fmt.Sprintf("biblioD.html historical bibliography block %d at source line %d; source species text=%s; %s", e.Block, e.Line, e.RawSpecies, e.Text)
		_ = w.Write([]string{e.Category, e.Species, e.Name, e.Name, fmt.Sprintf("bibliod:block-%04d:line-%05d:%s", e.Block, e.Line, strings.ToLower(e.Name)), sourceURL, note, e.Sequence, filepath.Base(*in), sourceSHA, fmt.Sprintf("line-%d", e.Line), e.Status})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("biblioD: %d auxiliary bibliography records\n", len(entries))
}
