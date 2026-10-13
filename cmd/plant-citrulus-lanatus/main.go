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
	sourceFile = "plants-Citrulus.lanatus.xlsx"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Citrulus.lanatus.xlsx"
	sheetName  = "Sorted by CYP name"
)

type record struct {
	Row                                                          int
	GotohID, SeqID, BestHit, PercentID, Symbol, Sequence, Status string
}

var pseudogeneRE = regexp.MustCompile(`(?i)\dP$`)

func main() {
	input := flag.String("input", filepath.Join("raw", sourceFile), "downloaded Citrulus lanatus workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "citrulus-lanatus.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-citrulus-lanatus.md"), "review ledger")
	flag.Parse()
	wb, err := plantxlsx.Read(*input, sheetName)
	if err != nil {
		panic(err)
	}
	accepted, excluded, err := reviewRows(wb)
	if err != nil {
		panic(err)
	}
	hash, err := fileHash(*input)
	if err != nil {
		panic(err)
	}
	if err := writeCSV(*out, accepted, hash); err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Citrulus lanatus: %d named rows accepted, %d unnamed rows excluded\n", len(accepted), len(excluded))
}

func reviewRows(wb *plantxlsx.Workbook) ([]record, []record, error) {
	if wb.Dimension != "A1:L244" {
		return nil, nil, fmt.Errorf("Citrulus lanatus used range changed: %q", wb.Dimension)
	}
	if len(wb.HiddenRows) != 0 || len(wb.HiddenColumns) != 0 || len(wb.MergedCells) != 0 || wb.TableParts != 0 {
		return nil, nil, fmt.Errorf("Citrulus lanatus workbook structure changed")
	}
	var accepted, excluded []record
	for row := 1; row <= 244; row++ {
		v := wb.Rows[row]
		if v == nil {
			return nil, nil, fmt.Errorf("missing Citrulus lanatus row %d", row)
		}
		r := record{Row: row, GotohID: strings.TrimSpace(v["A"]), SeqID: strings.TrimSpace(v["H"]), BestHit: strings.TrimSpace(v["I"]), PercentID: strings.TrimSpace(v["J"]), Symbol: strings.TrimSpace(v["K"]), Sequence: strings.TrimSpace(v["L"])}
		if r.GotohID == "" || r.SeqID == "" || r.Sequence == "" {
			return nil, nil, fmt.Errorf("Citrulus lanatus row %d missing ID or sequence", row)
		}
		if row >= 234 {
			if r.BestHit != "" || r.PercentID != "" || r.Symbol != "" {
				return nil, nil, fmt.Errorf("Citrulus lanatus unnamed row %d gained assignment", row)
			}
			r.Status = "excluded: no best hit, percent identity, or assigned CYP name"
			excluded = append(excluded, r)
			continue
		}
		if r.BestHit == "" || r.PercentID == "" || !strings.HasPrefix(r.BestHit, "CYP") || !strings.HasPrefix(r.Symbol, "CYP") {
			return nil, nil, fmt.Errorf("Citrulus lanatus row %d missing reviewed CYP assignment", row)
		}
		var status []string
		if pseudogeneRE.MatchString(r.BestHit) {
			status = append(status, "source-best-hit-pseudogene-label")
		}
		if strings.Contains(r.Sequence, "*") {
			return nil, nil, fmt.Errorf("Citrulus lanatus row %d unexpected stop marker", row)
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
			if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXY-", aa) {
				return nil, nil, fmt.Errorf("Citrulus lanatus row %d unexpected residue %q", row, aa)
			}
		}
		r.Status = strings.Join(status, ";")
		accepted = append(accepted, r)
	}
	if len(accepted) != 233 || len(excluded) != 11 {
		return nil, nil, fmt.Errorf("Citrulus lanatus boundary changed: accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	return accepted, excluded, nil
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
		note := fmt.Sprintf("Citrulus lanatus workbook row %d; Gotoh source ID=%s; seq ID=%s; best hit=%s; %%ID=%s; assigned CYP name column=K; sequence column=L", r.Row, r.GotohID, r.SeqID, r.BestHit, r.PercentID)
		_ = w.Write([]string{"plants", "Citrulus lanatus", r.Symbol, r.SeqID, fmt.Sprintf("citrulus-lanatus:row-%04d:%s", r.Row, r.SeqID), sourceURL, note, r.Sequence, sourceFile, hash, sheetName, fmt.Sprint(r.Row), r.Status})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, accepted, excluded []record, wb *plantxlsx.Workbook, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, r := range accepted {
		for _, status := range strings.Split(r.Status, ";") {
			if status != "" {
				counts[status]++
			}
		}
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "# Plant resource review: Citrulus lanatus\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheet: `%s`\n- Used range: `%s`\n- Header row: `none; row 1 is data`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted named rows: `%d`\n- Accepted rows with literal sequence: `%d`\n- Unnamed rows excluded: `%d`\n- Review status: `complete`\n\n", sourceFile, sourceURL, hash, sheetName, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(accepted), len(accepted), len(excluded))
	b.WriteString("## Resource-specific interpretation\n\nThis workbook has no header row: row 1 is the first CYP51G1 record. Rows 1-233 are the complete named region. For those rows A is the Gotoh source ID, H is the output sequence ID, I is the best hit, J is percent identity, K is the source-assigned CYP name, and L is the literal protein sequence. The output symbol is K, not the more specific comparison value in I. Rows 234-244 retain IDs and sequence-like text but have empty I/J/K assignments; they are excluded as unnamed models. Gaps, O, and X in accepted L cells are retained. Duplicate assigned names and four duplicate-sequence groups remain row-distinct. No sequence is repaired or deduplicated.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, name := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", name, counts[name])
	}
	b.WriteString("\n## Representative and excluded rows\n\n| Row | Seq ID | Best hit | Assigned CYP | Sequence | Status |\n|---:|---|---|---|---:|---|\n")
	for _, row := range []int{1, 117, 233} {
		for _, r := range accepted {
			if r.Row == row {
				fmt.Fprintf(&b, "| %d | %s | %s | %s | %d aa | %s |\n", r.Row, md(r.SeqID), md(r.BestHit), md(r.Symbol), len(r.Sequence), md(r.Status))
			}
		}
	}
	for _, r := range excluded {
		fmt.Fprintf(&b, "| %d | %s |  |  | %d chars (excluded) | %s |\n", r.Row, md(r.SeqID), len(r.Sequence), md(r.Status))
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
