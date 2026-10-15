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

const (
	chimpSourceFile = "animals-Pan.trog.P450s.Dec8.2014.xlsx"
	chimpSourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Pan.trog.P450s.Dec8.2014.xlsx"
	chimpSheet      = "sorted by CYP name"
)

var coordinateIDRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type chimpRecord struct {
	Row                                     int
	GeneID, TranscriptID, Start, Stop       string
	Chromosome, Strand, HumanHit, HumanID   string
	GotohHit, GotohID, Comment, Symbol      string
	SourceSequence, Sequence, ID, RecordKey string
	Statuses                                []string
}

func main() {
	input := flag.String("input", filepath.Join("raw", chimpSourceFile), "downloaded chimp workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "chimp.csv"), "reviewed structured CSV")
	audit := flag.String("audit", filepath.Join("docs", "animals", "chimp.md"), "resource review ledger")
	flag.Parse()

	names, err := plantxlsx.SheetNames(*input)
	if err != nil {
		panic(err)
	}
	expected := []string{"Pan.troglodytes Gotoh", "Pan Troglodytes Ensembl", "combined by chr order", chimpSheet}
	if strings.Join(names, "\x00") != strings.Join(expected, "\x00") {
		panic(fmt.Errorf("chimp workbook sheet layout changed: got %q", names))
	}
	wb, err := plantxlsx.ReadSheet(*input, chimpSheet, false)
	if err != nil {
		panic(err)
	}
	records, err := reviewChimpRows(wb.Rows)
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if err := writeChimpCSV(*out, records, hash); err != nil {
		panic(err)
	}
	if err := writeChimpAudit(*audit, records, wb, names, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Pan troglodytes: %d reviewed rows, %d literal protein sequences\n", len(records), len(records))
}

func reviewChimpRows(rows map[int]map[string]string) ([]chimpRecord, error) {
	header := rows[1]
	expected := map[string]string{
		"A": "chimp Ensembl gene ID", "B": "transcript ID", "C": "start", "D": "stop ",
		"E": "chr", "F": "strand", "G": "Best hit human", "H": "%ID",
		"I": "Best hit Gotohs seqs", "J": "%ID", "K": "comment", "L": "CYP name", "M": "sequence",
	}
	for col, want := range expected {
		if header[col] != want {
			return nil, fmt.Errorf("chimp header layout changed at %s: got %q want %q", col, header[col], want)
		}
	}
	var records []chimpRecord
	for row := 2; row <= 122; row++ {
		v := rows[row]
		r := chimpRecord{
			Row: row, GeneID: strings.TrimSpace(v["A"]), TranscriptID: strings.TrimSpace(v["B"]),
			Start: strings.TrimSpace(v["C"]), Stop: strings.TrimSpace(v["D"]), Chromosome: strings.TrimSpace(v["E"]),
			Strand: strings.TrimSpace(v["F"]), HumanHit: strings.TrimSpace(v["G"]), HumanID: strings.TrimSpace(v["H"]),
			GotohHit: strings.TrimSpace(v["I"]), GotohID: strings.TrimSpace(v["J"]), Comment: strings.TrimSpace(v["K"]),
			Symbol: strings.TrimSpace(v["L"]), SourceSequence: strings.TrimSpace(v["M"]),
		}
		if r.Symbol == "" || r.SourceSequence == "" {
			return nil, fmt.Errorf("chimp final review row %d is missing CYP name or sequence", row)
		}
		r.Sequence, r.Statuses = normalizeChimpSequence(r.SourceSequence)
		if r.Sequence == "" {
			return nil, fmt.Errorf("chimp final review row %d has an empty normalized sequence", row)
		}
		for _, residue := range r.Sequence {
			upper := unicode.ToUpper(residue)
			if (upper < 'A' || upper > 'Z') && residue != '*' {
				return nil, fmt.Errorf("chimp row %d contains unexpected sequence character %q", row, residue)
			}
		}
		lowerSymbol := strings.ToLower(r.Symbol)
		lowerComment := strings.ToLower(r.Comment)
		if strings.Contains(r.Sequence, "*") {
			r.Statuses = append(r.Statuses, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Statuses = append(r.Statuses, "ambiguous-X-retained")
		}
		if len(r.Sequence) < 350 {
			r.Statuses = append(r.Statuses, "short-source-fragment")
		}
		if strings.Contains(lowerSymbol, "partial") || strings.Contains(lowerSymbol, "exon") || strings.Contains(lowerComment, "partial") || strings.Contains(lowerComment, "missing") {
			r.Statuses = append(r.Statuses, "source-partial-or-missing-region")
		}
		if r.Symbol == "CYP2B6 partial seq" {
			r.Statuses = append(r.Statuses, "duplicate-symbol-source-row")
		}
		r.ID = firstNonEmpty(r.GeneID, gotohBase(r.GotohHit), coordinateID(r), r.Symbol)
		r.RecordKey = fmt.Sprintf("chimp:sheet-sorted-by-cyp-name:row-%04d:%s", row, keyPart(r.ID))
		records = append(records, r)
	}
	if len(records) != 121 {
		return nil, fmt.Errorf("chimp reviewed row count changed: got %d want 121", len(records))
	}
	return records, nil
}

func normalizeChimpSequence(source string) (string, []string) {
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

func gotohBase(value string) string {
	value = strings.TrimSpace(value)
	if i := strings.Index(value, ","); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}

func coordinateID(r chimpRecord) string {
	if r.Chromosome == "" || r.Start == "" || r.Stop == "" {
		return ""
	}
	return fmt.Sprintf("chr%s:%s-%s:%s", r.Chromosome, r.Start, r.Stop, r.Strand)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func keyPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = coordinateIDRE.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func writeChimpCSV(path string, records []chimpRecord, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "gene_id", "transcript_id", "gotoh_hit", "chromosome", "start", "stop", "strand", "human_hit", "review_status"})
	for _, r := range records {
		note := fmt.Sprintf("Chimp workbook final curated sheet %q row %d; gene=%s; transcript=%s; coordinates=chr%s:%s-%s strand %s; best human hit=%s (%s%% ID); Gotoh hit=%s (%s%% ID); comment=%s", chimpSheet, r.Row, r.GeneID, r.TranscriptID, r.Chromosome, r.Start, r.Stop, r.Strand, r.HumanHit, r.HumanID, r.GotohHit, r.GotohID, r.Comment)
		_ = w.Write([]string{"animals", "Pan troglodytes", r.Symbol, r.ID, r.RecordKey, chimpSourceURL, note, r.Sequence, chimpSourceFile, hash, chimpSheet, strconv.Itoa(r.Row), r.GeneID, r.TranscriptID, r.GotohHit, r.Chromosome, r.Start, r.Stop, r.Strand, r.HumanHit, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeChimpAudit(path string, records []chimpRecord, wb *plantxlsx.Workbook, sheets []string, hash string) error {
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
	for name := range counts {
		statusNames = append(statusNames, name)
	}
	sort.Strings(statusNames)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Pan troglodytes (chimp)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheets: `%s`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted records: `%d`\n- Records with literal protein sequence: `%d`\n- Missing-sequence records: `0`\n- Review status: `complete`\n\n", chimpSourceFile, chimpSourceURL, hash, strings.Join(sheets, "`, `"), chimpSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(records), len(records))
	b.WriteString("## Resource-specific interpretation\n\nThe workbook contains two intermediate source sheets (`Pan.troglodytes Gotoh` and `Pan Troglodytes Ensembl`) and two curated combined views. The two combined views describe the same biological row set in different orders, but the final `sorted by CYP name` sheet contains five later nomenclature edits relative to `combined by chr order` (including CYP3A156P/CYP3A157P, CYP51A1, and the expanded CYP46A11P and CYP21A1 hybrid labels). Only the final sorted sheet is therefore published; intermediate and alternate-sort rows are not appended or deduplicated into extra records.\n\nRows 2–122 each contain an explicit CYP name in column L and literal protein text in column M. All 121 rows are retained, including complete genes, short exon fragments, partial sequences, pseudogene orthologs, two source-distinct rows both labeled `CYP2B6 partial seq`, sequences with internal stops, and manually revised/augmented sequence text. Column A Ensembl gene ID is the preferred searchable ID. When it is absent, the Gotoh identifier is used; if both are absent, the chromosome coordinate is used. Transcript, Gotoh hit, coordinates, human hit, percent identities, and comments remain row-level provenance.\n\nWhitespace inside the CYP2S1 source cell is spreadsheet layout and is joined without inserting residues. The mixed-case CYP2B6 sequence is retained in its literal source case because lowercase marks the workbook author's augmented segment. Only one explicit terminal `*` is removed. Every internal `*`, `X`, short fragment, duplicate symbol, and source comment remains intact. No sequence is translated, repaired, filled from another sheet/database, or deduplicated.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, name := range statusNames {
		fmt.Fprintf(&b, "| %s | %d |\n", name, counts[name])
	}
	b.WriteString("\n## Representative rows\n\n| Source row | CYP name | ID | Sequence length | Notes |\n|---:|---|---|---:|---|\n")
	for _, index := range []int{0, 8, 37, len(records) / 2, len(records) - 1} {
		r := records[index]
		fmt.Fprintf(&b, "| %d | %s | %s | %d aa | %s |\n", r.Row, md(r.Symbol), md(r.ID), len(r.Sequence), md(strings.Join(r.Statuses, ";")))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func md(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "|", "\\|")
}
