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
	"strings"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const (
	sourceFile = "plants-Lotus.japonicus.xlsx"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Lotus.japonicus.xlsx"
	sheetName  = "Sorted by CYP name"
)

type record struct {
	Row                                int
	GotohID, SeqID, BestHit, PercentID string
	Symbol, Sequence, Status           string
}

var digitP = regexp.MustCompile(`(?i)\dP$`)

func main() {
	input := flag.String("input", filepath.Join("raw", sourceFile), "downloaded Lotus japonicus workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "lotus-japonicus.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-lotus-japonicus.md"), "review ledger")
	flag.Parse()
	wb, err := plantxlsx.Read(*input, sheetName)
	if err != nil {
		panic(err)
	}
	records, err := reviewRows(wb)
	if err != nil {
		panic(err)
	}
	hash, err := fileHash(*input)
	if err != nil {
		panic(err)
	}
	if hash != "d8afd69e5393c6e88f20f179a482af815be4c5de2e3debb69dcec4106bb58a50" {
		panic(fmt.Errorf("Lotus japonicus source hash changed: %s", hash))
	}
	if err := writeCSV(*out, records, hash); err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, records, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Lotus japonicus: %d assigned rows accepted with literal sequence; 39 unassigned rows excluded\n", len(records))
}

func reviewRows(wb *plantxlsx.Workbook) ([]record, error) {
	if wb.Dimension != "A1:L286" {
		return nil, fmt.Errorf("Lotus japonicus used range changed: %q", wb.Dimension)
	}
	if len(wb.HiddenRows) != 0 || len(wb.HiddenColumns) != 0 || len(wb.MergedCells) != 0 || wb.TableParts != 0 {
		return nil, fmt.Errorf("Lotus japonicus workbook structure changed")
	}
	h := wb.Rows[1]
	for _, column := range []string{"A", "B", "C", "D", "E", "F", "G", "L"} {
		if strings.TrimSpace(h[column]) != "" {
			return nil, fmt.Errorf("Lotus japonicus unexpected header in %s1: %q", column, h[column])
		}
	}
	if h["H"] != "seq. ID" || h["I"] != "best hit" || h["J"] != "%ID" || h["K"] != "CYP name" {
		return nil, fmt.Errorf("Lotus japonicus header changed")
	}

	records := make([]record, 0, 246)
	for row := 2; row <= 247; row++ {
		v := wb.Rows[row]
		if v == nil {
			return nil, fmt.Errorf("missing Lotus japonicus assigned row %d", row)
		}
		r := record{
			Row:       row,
			GotohID:   strings.TrimSpace(v["A"]),
			SeqID:     strings.TrimSpace(v["H"]),
			BestHit:   strings.TrimSpace(v["I"]),
			PercentID: strings.TrimSpace(v["J"]),
			Symbol:    strings.TrimSpace(v["K"]),
			Sequence:  strings.TrimSpace(v["L"]),
		}
		if r.GotohID == "" || r.SeqID == "" || r.BestHit == "" || r.PercentID == "" || r.Symbol == "" || r.Sequence == "" {
			return nil, fmt.Errorf("Lotus japonicus row %d missing assigned value", row)
		}
		if r.GotohID != r.SeqID {
			return nil, fmt.Errorf("Lotus japonicus row %d A/H ID mismatch: %q != %q", row, r.GotohID, r.SeqID)
		}
		if !strings.HasPrefix(r.BestHit, "CYP") || !strings.HasPrefix(r.Symbol, "CYP") {
			return nil, fmt.Errorf("Lotus japonicus row %d non-CYP assignment: best=%q assigned=%q", row, r.BestHit, r.Symbol)
		}
		var status []string
		if digitP.MatchString(r.BestHit) {
			status = append(status, "source-best-hit-pseudogene-label")
		}
		if digitP.MatchString(r.Symbol) {
			status = append(status, "source-pseudogene-label")
		}
		if strings.HasSuffix(r.Sequence, "*") {
			r.Sequence = strings.TrimRight(r.Sequence, "*")
			status = append(status, "source-terminal-stop")
		}
		if strings.Contains(r.Sequence, "*") {
			status = append(status, "internal-stop")
		}
		if strings.Contains(r.Sequence, "-") {
			status = append(status, "source-gap")
		}
		if strings.Contains(r.Sequence, "X") {
			status = append(status, "ambiguous-X")
		}
		if strings.Contains(r.Sequence, "O") {
			status = append(status, "nonstandard-O")
		}
		if len(r.Sequence) < 350 {
			status = append(status, "short-sequence")
		}
		for _, aa := range r.Sequence {
			if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXY*-", aa) {
				return nil, fmt.Errorf("Lotus japonicus row %d unexpected residue %q", row, aa)
			}
		}
		r.Status = strings.Join(status, ";")
		records = append(records, r)
	}
	for row := 248; row <= 286; row++ {
		v := wb.Rows[row]
		if v == nil || strings.TrimSpace(v["H"]) == "" || strings.TrimSpace(v["L"]) == "" {
			return nil, fmt.Errorf("Lotus japonicus unassigned row %d layout changed", row)
		}
		if strings.TrimSpace(v["I"]) != "" || strings.TrimSpace(v["J"]) != "" || strings.TrimSpace(v["K"]) != "" {
			return nil, fmt.Errorf("Lotus japonicus row %d is no longer unassigned", row)
		}
	}
	if len(records) != 246 {
		return nil, fmt.Errorf("Lotus japonicus boundary changed: accepted=%d", len(records))
	}
	return records, nil
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
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "review_status"})
	for _, r := range records {
		note := fmt.Sprintf("Lotus japonicus workbook row %d; Gotoh source ID=%s; seq ID=%s; best hit=%s; percent identity=%s; assigned CYP=%s; literal sequence column=L", r.Row, r.GotohID, r.SeqID, r.BestHit, r.PercentID, r.Symbol)
		_ = w.Write([]string{"plants", "Lotus japonicus", r.Symbol, r.SeqID, fmt.Sprintf("lotus-japonicus:row-%04d:%s", r.Row, r.SeqID), sourceURL, note, r.Sequence, sourceFile, hash, sheetName, fmt.Sprint(r.Row), r.Status})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, wb *plantxlsx.Workbook, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	sequences := map[string]int{}
	for _, r := range records {
		sequences[r.Sequence]++
		for _, status := range strings.Split(r.Status, ";") {
			if status != "" {
				counts[status]++
			}
		}
	}
	duplicateSequences := 0
	for _, count := range sequences {
		if count > 1 {
			duplicateSequences++
		}
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "# Plant resource review: Lotus japonicus\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted assigned rows: %d\n- Accepted rows with literal sequence: %d\n- Excluded unassigned rows: 39\n- Duplicate literal-sequence groups retained: %d\n- Review status: `complete`\n\n", sourceFile, sourceURL, hash, sheetName, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(records), len(records), duplicateSequences)
	b.WriteString("## Resource-specific interpretation\n\nThe workbook contains one worksheet with `A1:L286`. Row 1 names H=`seq. ID`, I=`best hit`, J=`%ID`, and K=`CYP name`; L has no header but is the literal sequence field for every data row. Rows 2-247 are the complete continuous assigned region: A and H contain the same stable ID, and I/J/K are all populated. Rows 248-286 retain IDs and model text in L but have empty I/J/K assignment fields and are excluded. The 246 accepted rows preserve 11 gap-bearing sequences, literal X in 11 records, literal O in 14 records, 43 short sequences, three digit-P best-hit labels, one digit-P assigned label, and five duplicate-sequence groups with distinct IDs. No sequence is inferred, repaired, or deduplicated.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, name := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", name, counts[name])
	}
	b.WriteString("\n## Representative rows\n\n| Row | Seq ID | Best hit | Percent identity | Assigned CYP | Sequence | Status |\n|---:|---|---|---:|---|---:|---|\n")
	for _, row := range []int{2, 124, 247} {
		for _, r := range records {
			if r.Row == row {
				fmt.Fprintf(&b, "| %d | %s | %s | %s | %s | %d aa | %s |\n", r.Row, md(r.SeqID), md(r.BestHit), md(r.PercentID), md(r.Symbol), len(r.Sequence), md(r.Status))
			}
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func md(value string) string { return strings.ReplaceAll(strings.TrimSpace(value), "|", "\\|") }
