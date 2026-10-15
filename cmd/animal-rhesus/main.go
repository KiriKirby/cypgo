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
	"unicode"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var rhesusKeyPartRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

const (
	rhesusSourceFile = "animals-Rhesus.P450s.Dec8.2014.xlsx"
	rhesusSourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Rhesus.P450s.Dec8.2014.xlsx"
	rhesusSheet      = "sorted by CYP name"
)

type rhesusRecord struct {
	Row                                     int
	GotohID, GeneID, TranscriptID           string
	Chromosome, Start, Stop, Strand         string
	HomepageHit, PercentID, Symbol          string
	SourceSequence, Sequence, ID, RecordKey string
	Statuses                                []string
}

func main() {
	input := flag.String("input", filepath.Join("raw", rhesusSourceFile), "downloaded rhesus workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "rhesus-monkey.csv"), "reviewed structured CSV")
	audit := flag.String("audit", filepath.Join("docs", "animals", "rhesus-monkey.md"), "resource review ledger")
	flag.Parse()

	names, err := plantxlsx.SheetNames(*input)
	if err != nil {
		panic(err)
	}
	expected := []string{"Rhesus.csv.txt", rhesusSheet, "Sorted by location"}
	if strings.Join(names, "\x00") != strings.Join(expected, "\x00") {
		panic(fmt.Errorf("rhesus workbook sheet layout changed: got %q", names))
	}
	wb, err := plantxlsx.ReadSheet(*input, rhesusSheet, false)
	if err != nil {
		panic(err)
	}
	records, err := reviewRhesusRows(wb.Rows)
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if err := writeRhesusCSV(*out, records, hash); err != nil {
		panic(err)
	}
	if err := writeRhesusAudit(*audit, records, wb, names, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Macaca mulatta: %d reviewed rows, %d literal protein sequences\n", len(records), len(records))
}

func reviewRhesusRows(rows map[int]map[string]string) ([]rhesusRecord, error) {
	header := rows[1]
	expected := map[string]string{
		"A": "Gotoh's ID", "B": "Ensembl gene ID", "C": "Ensembl transcript ID", "D": "chromosome",
		"E": "start", "F": "stop", "G": "strand", "H": "best hit in homepage list (49 seqs)",
		"I": "%ID", "J": "comment/CYP name", "K": "sequence", "L": "",
	}
	for col, want := range expected {
		if header[col] != want {
			return nil, fmt.Errorf("rhesus header layout changed at %s: got %q want %q", col, header[col], want)
		}
	}
	var records []rhesusRecord
	for row := 2; row <= 111; row++ {
		v := rows[row]
		r := rhesusRecord{
			Row: row, GotohID: strings.TrimSpace(v["A"]), GeneID: strings.TrimSpace(v["B"]), TranscriptID: strings.TrimSpace(v["C"]),
			Chromosome: strings.TrimSpace(v["D"]), Start: strings.TrimSpace(v["E"]), Stop: strings.TrimSpace(v["F"]),
			Strand: strings.TrimSpace(v["G"]), HomepageHit: strings.TrimSpace(v["H"]), PercentID: strings.TrimSpace(v["I"]),
			Symbol: strings.TrimSpace(v["J"]), SourceSequence: strings.TrimSpace(v["K"]),
		}
		if r.Symbol == "" || !strings.HasPrefix(strings.ToUpper(r.Symbol), "CYP") || r.SourceSequence == "" {
			return nil, fmt.Errorf("rhesus final review row %d is missing a CYP name or sequence", row)
		}
		r.Sequence, r.Statuses = normalizeRhesusSequence(r.SourceSequence)
		for _, residue := range r.Sequence {
			upper := unicode.ToUpper(residue)
			if (upper < 'A' || upper > 'Z') && residue != '*' {
				return nil, fmt.Errorf("rhesus row %d contains unexpected sequence character %q", row, residue)
			}
		}
		lower := strings.ToLower(r.Symbol)
		if strings.Contains(r.Sequence, "*") {
			r.Statuses = append(r.Statuses, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Statuses = append(r.Statuses, "ambiguous-X-retained")
		}
		if len(r.Sequence) < 350 {
			r.Statuses = append(r.Statuses, "short-source-fragment")
		}
		if strings.Contains(lower, "partial") || strings.Contains(lower, "fragment") || strings.Contains(lower, "exon") || strings.Contains(lower, "missing") || strings.Contains(lower, "incomplete") {
			r.Statuses = append(r.Statuses, "source-partial-or-missing-region")
		}
		r.ID = rhesusPreferredID(r)
		r.RecordKey = fmt.Sprintf("rhesus-monkey:sheet-sorted-by-cyp-name:row-%04d:%s", row, keyPart(r.ID))
		records = append(records, r)
	}
	if len(records) != 110 {
		return nil, fmt.Errorf("rhesus reviewed row count changed: got %d want 110", len(records))
	}
	return records, nil
}

func normalizeRhesusSequence(source string) (string, []string) {
	var b strings.Builder
	hadWhitespace, hadLower := false, false
	for _, r := range source {
		if unicode.IsSpace(r) {
			hadWhitespace = true
			continue
		}
		if unicode.IsLower(r) {
			hadLower = true
		}
		b.WriteRune(r)
	}
	sequence := b.String()
	var statuses []string
	if hadWhitespace {
		statuses = append(statuses, "source-whitespace-joined")
	}
	if hadLower {
		statuses = append(statuses, "source-lowercase-retained")
	}
	if strings.HasSuffix(sequence, "*") {
		sequence = strings.TrimSuffix(sequence, "*")
		statuses = append(statuses, "terminal-stop-removed")
	}
	return sequence, statuses
}

func rhesusPreferredID(r rhesusRecord) string {
	if strings.HasPrefix(r.GeneID, "ENSMMUG") {
		return r.GeneID
	}
	if r.GotohID != "" {
		return r.GotohID
	}
	if strings.HasPrefix(r.TranscriptID, "ENSMMUT") {
		return r.TranscriptID
	}
	if r.Chromosome != "" && r.Start != "" && r.Stop != "" {
		return fmt.Sprintf("chr%s:%s-%s:%s", r.Chromosome, r.Start, r.Stop, r.Strand)
	}
	return r.Symbol
}

func keyPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = rhesusKeyPartRE.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func writeRhesusCSV(path string, records []rhesusRecord, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "gotoh_id", "gene_id", "transcript_id", "chromosome", "start", "stop", "strand", "homepage_hit", "percent_identity", "review_status"})
	for _, r := range records {
		note := fmt.Sprintf("Rhesus workbook final CYP-name sheet row %d; exact source CYP/comment=%s; Gotoh ID=%s; Ensembl gene=%s; transcript=%s; coordinates=chr%s:%s-%s strand %s; homepage-list hit=%s (%s%% ID)", r.Row, r.Symbol, r.GotohID, r.GeneID, r.TranscriptID, r.Chromosome, r.Start, r.Stop, r.Strand, r.HomepageHit, r.PercentID)
		_ = w.Write([]string{"animals", "Macaca mulatta", r.Symbol, r.ID, r.RecordKey, rhesusSourceURL, note, r.Sequence, rhesusSourceFile, hash, rhesusSheet, strconv.Itoa(r.Row), r.GotohID, r.GeneID, r.TranscriptID, r.Chromosome, r.Start, r.Stop, r.Strand, r.HomepageHit, r.PercentID, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeRhesusAudit(path string, records []rhesusRecord, wb *plantxlsx.Workbook, sheets []string, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, r := range records {
		for _, status := range r.Statuses {
			counts[status]++
		}
	}
	statusNames := make([]string, 0, len(counts))
	for status := range counts {
		statusNames = append(statusNames, status)
	}
	sort.Strings(statusNames)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Macaca mulatta (rhesus monkey)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheets: `%s`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted records: `%d`\n- Records with literal protein sequence: `%d`\n- Missing-sequence records: `0`\n- Review status: `complete`\n\n", rhesusSourceFile, rhesusSourceURL, hash, strings.Join(sheets, "`, `"), rhesusSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(records), len(records))
	b.WriteString("## Resource-specific interpretation\n\nThe workbook's first sheet is the 117-row Ensembl BioMart/PF00067 starting candidate set, not the final nomenclature. The other two sheets are manually curated views. `sorted by CYP name` has 110 sequence-bearing CYP rows and is authoritative: it contains the CYP4F77P row absent from `Sorted by location`, calls the hybrid row CYP21A2 rather than CYP21, and uses the later CYP3A143P wording. Neither the BioMart candidates nor the alternate location ordering is appended as additional records.\n\nThe footer says `100 sequences are named. 35 are pseudogenes and 65 are genes`, while rows 2–111 actually contain 110 non-empty CYP-name/sequence pairs. All 110 literal named rows are retained and the discrepancy is documented rather than deleting ten records to force the footer total. The exact column-J CYP/comment text remains the displayed symbol because it carries the source's completeness, orthology, old-name, and assembly qualifications. A real Ensembl gene ID is preferred as record ID, followed by Gotoh ID, Ensembl transcript, coordinates, and finally the source CYP/comment when no identifier or complete coordinate exists. Placeholder text such as `not in BIOMART set` and `no gene` is never used as an ID.\n\nThe two lowercase `x` residues are retained literally because they occur in the reviewed sequence cells. Only an explicit terminal `*` is removed. Internal stops, pseudogenes, fragments, manually revised sequences, absent BioMart identifiers, and the complete CYP4A22 sequence linked by the source comment to AC203029.4 remain source-distinct. No sequence is filled from the candidate or location sheet, translated, repaired, or deduplicated.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, status := range statusNames {
		fmt.Fprintf(&b, "| %s | %d |\n", status, counts[status])
	}
	b.WriteString("\n## Representative rows\n\n| Source row | CYP name/comment | ID | Sequence length | Notes |\n|---:|---|---|---:|---|\n")
	for _, index := range []int{0, 11, 24, 59, 79, len(records) - 1} {
		r := records[index]
		fmt.Fprintf(&b, "| %d | %s | %s | %d aa | %s |\n", r.Row, md(r.Symbol), md(r.ID), len(r.Sequence), md(strings.Join(r.Statuses, ";")))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func md(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "|", "\\|")
}
