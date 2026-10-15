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
	marmosetSourceFile = "animals-Marmoset.P450s.Dec8.2014.xlsx"
	marmosetSourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Marmoset.P450s.Dec8.2014.xlsx"
	marmosetSheet      = "Sorted by CYP name"
)

var marmosetKeyRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type marmosetRecord struct {
	Row                                                                           int
	EnsemblID, Location, Symbol, Species, SourceSequence, Sequence, ID, RecordKey string
	Statuses                                                                      []string
}

func main() {
	input := flag.String("input", filepath.Join("raw", marmosetSourceFile), "downloaded marmoset workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "marmoset.csv"), "reviewed structured CSV")
	audit := flag.String("audit", filepath.Join("docs", "animals", "marmoset.md"), "resource review ledger")
	flag.Parse()
	names, err := plantxlsx.SheetNames(*input)
	if err != nil {
		panic(err)
	}
	if strings.Join(names, "\x00") != strings.Join([]string{"Sorted by location", marmosetSheet}, "\x00") {
		panic(fmt.Errorf("marmoset workbook sheet layout changed: %q", names))
	}
	wb, err := plantxlsx.ReadSheet(*input, marmosetSheet, false)
	if err != nil {
		panic(err)
	}
	records, err := reviewMarmosetRows(wb.Rows)
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if err := writeMarmosetCSV(*out, records, hash); err != nil {
		panic(err)
	}
	if err := writeMarmosetAudit(*audit, records, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Callithrix jacchus: %d reviewed rows, %d literal protein sequences\n", len(records), len(records))
}

func reviewMarmosetRows(rows map[int]map[string]string) ([]marmosetRecord, error) {
	header := rows[1]
	expected := map[string]string{"A": "Ensembl ID", "B": "Location chr:position:strand", "C": "CYP name", "D": "species", "E": "sequence"}
	for col, want := range expected {
		if header[col] != want {
			return nil, fmt.Errorf("marmoset header changed at %s: got %q want %q", col, header[col], want)
		}
	}
	var out []marmosetRecord
	for row := 2; row <= 88; row++ {
		v := rows[row]
		r := marmosetRecord{Row: row, EnsemblID: strings.TrimSpace(v["A"]), Location: strings.TrimSpace(v["B"]), Symbol: strings.TrimSpace(v["C"]), Species: strings.TrimSpace(v["D"]), SourceSequence: strings.TrimSpace(v["E"])}
		if r.Symbol == "" || r.SourceSequence == "" || r.Species != "marmoset" {
			return nil, fmt.Errorf("marmoset row %d missing reviewed CYP/sequence/species", row)
		}
		r.Sequence, r.Statuses = normalizeMarmosetSequence(r.SourceSequence)
		for _, residue := range r.Sequence {
			upper := unicode.ToUpper(residue)
			if (upper < 'A' || upper > 'Z') && residue != '*' && residue != '-' {
				return nil, fmt.Errorf("marmoset row %d unexpected residue %q", row, residue)
			}
		}
		lower := strings.ToLower(r.Symbol)
		if strings.Contains(r.Sequence, "*") {
			r.Statuses = append(r.Statuses, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Statuses = append(r.Statuses, "ambiguous-X-retained")
		}
		if strings.Contains(r.Sequence, "-") {
			r.Statuses = append(r.Statuses, "source-gap-retained")
		}
		if len(r.Sequence) < 350 {
			r.Statuses = append(r.Statuses, "short-source-fragment")
		}
		if strings.Contains(lower, "fragment") || strings.Contains(lower, "partial") || strings.Contains(lower, "missing") || strings.Contains(lower, "exon") || strings.Contains(lower, "gap") {
			r.Statuses = append(r.Statuses, "source-partial-or-missing-region")
		}
		r.ID = firstMarmoset(r.EnsemblID, r.Location, r.Symbol)
		r.RecordKey = fmt.Sprintf("marmoset:sheet-sorted-by-cyp-name:row-%04d:%s", row, marmosetKeyPart(r.ID))
		out = append(out, r)
	}
	if len(out) != 87 {
		return nil, fmt.Errorf("marmoset count changed: got %d want 87", len(out))
	}
	return out, nil
}

func normalizeMarmosetSequence(source string) (string, []string) {
	var b strings.Builder
	lower := false
	whitespace := false
	for _, r := range source {
		if unicode.IsSpace(r) {
			whitespace = true
			continue
		}
		if unicode.IsLower(r) {
			lower = true
		}
		b.WriteRune(r)
	}
	seq := b.String()
	var status []string
	if whitespace {
		status = append(status, "source-whitespace-joined")
	}
	if lower {
		status = append(status, "source-lowercase-retained")
	}
	if strings.HasSuffix(seq, "*") {
		seq = strings.TrimSuffix(seq, "*")
		status = append(status, "terminal-stop-removed")
	}
	return seq, status
}

func firstMarmoset(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func marmosetKeyPart(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	return strings.Trim(marmosetKeyRE.ReplaceAllString(v, "-"), "-")
}

func writeMarmosetCSV(path string, records []marmosetRecord, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "ensembl_id", "location", "review_status"})
	for _, r := range records {
		note := fmt.Sprintf("Marmoset workbook final CYP-name sheet row %d; Ensembl ID=%s; source location=%s; exact CYP name=%s", r.Row, r.EnsemblID, r.Location, r.Symbol)
		_ = w.Write([]string{"animals", "Callithrix jacchus", r.Symbol, r.ID, r.RecordKey, marmosetSourceURL, note, r.Sequence, marmosetSourceFile, hash, marmosetSheet, strconv.Itoa(r.Row), r.EnsemblID, r.Location, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeMarmosetAudit(path string, records []marmosetRecord, wb *plantxlsx.Workbook, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, r := range records {
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
	fmt.Fprintf(&b, "# Animal resource review: Callithrix jacchus (marmoset)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheets: `Sorted by location`, `%s`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted records: `%d`\n- Records with literal protein sequence: `%d`\n- Missing-sequence records: `0`\n- Review status: `complete`\n\n", marmosetSourceFile, marmosetSourceURL, hash, marmosetSheet, marmosetSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(records), len(records))
	b.WriteString("## Resource-specific interpretation\n\nThe workbook has two alternate orderings of the same 87 source-distinct rows. Their CYP-name/sequence multisets are identical, so only `Sorted by CYP name` is published. Rows 2–88 each contain an explicit CYP name, the literal species label `marmoset`, and a sequence. The Ensembl ID is preferred as ID; 32 rows without one use the exact genomic location, and a CYP name is the final fallback. The footer notes that Ensembl regions may span gene clusters and that some nucleotide regions were edited to coding boundaries; these statements are preserved as provenance and do not authorize sequence repair.\n\nLowercase residues in six cells, one `-x` source gap, internal stops, short fragments, pseudogenes, incomplete genes and revised sequences are retained literally. Only explicit terminal `*` markers are removed. No case conversion, translation, repair, cross-sheet merge, or deduplication is performed.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	b.WriteString("\n## Representative rows\n\n| Source row | CYP name | ID | Sequence length | Notes |\n|---:|---|---|---:|---|\n")
	for _, i := range []int{0, 10, 40, len(records) / 2, len(records) - 1} {
		r := records[i]
		fmt.Fprintf(&b, "| %d | %s | %s | %d aa | %s |\n", r.Row, mdMarmoset(r.Symbol), mdMarmoset(r.ID), len(r.Sequence), mdMarmoset(strings.Join(r.Statuses, ";")))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func mdMarmoset(v string) string { return strings.ReplaceAll(strings.TrimSpace(v), "|", "\\|") }
