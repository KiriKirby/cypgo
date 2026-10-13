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
	sourceFile = "plants-Cucumis.sativus.xlsx"
	sourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Cucumis.sativus.xlsx"
	sheetName  = "sorted by CYP name"
)

type record struct {
	Row                                      int
	GotohID, SeqID, Symbol, Sequence, Status string
}

var pseudogeneRE = regexp.MustCompile(`(?i)\dP$`)

func main() {
	input := flag.String("input", filepath.Join("raw", sourceFile), "downloaded Cucumis sativus workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "cucumis-sativus.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-cucumis-sativus.md"), "review ledger")
	flag.Parse()
	wb, err := plantxlsx.Read(*input, sheetName)
	if err != nil {
		panic(err)
	}
	accepted, err := reviewRows(wb)
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
	if err := writeAudit(*audit, accepted, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Cucumis sativus: %d rows accepted, %d literal sequences\n", len(accepted), len(accepted))
}

func reviewRows(wb *plantxlsx.Workbook) ([]record, error) {
	if wb.Dimension != "A1:J230" {
		return nil, fmt.Errorf("Cucumis sativus used range changed: %q", wb.Dimension)
	}
	if len(wb.HiddenRows) != 0 || len(wb.HiddenColumns) != 0 || len(wb.MergedCells) != 0 || wb.TableParts != 0 {
		return nil, fmt.Errorf("Cucumis sativus workbook structure changed")
	}
	h := wb.Rows[1]
	if h["A"] != "Gotoh's seq ID" || h["H"] != "seq ID" || h["I"] != "CYP name" || h["J"] != "sequence" {
		return nil, fmt.Errorf("Cucumis sativus header changed")
	}
	accepted := make([]record, 0, 229)
	for row := 2; row <= 230; row++ {
		v := wb.Rows[row]
		if v == nil {
			return nil, fmt.Errorf("missing Cucumis sativus row %d", row)
		}
		r := record{Row: row, GotohID: strings.TrimSpace(v["A"]), SeqID: strings.TrimSpace(v["H"]), Symbol: strings.TrimSpace(v["I"]), Sequence: strings.TrimSpace(v["J"])}
		if r.GotohID == "" || r.SeqID == "" || r.Symbol == "" || r.Sequence == "" {
			return nil, fmt.Errorf("Cucumis sativus row %d missing required value", row)
		}
		if !strings.HasPrefix(r.Symbol, "CYP") {
			return nil, fmt.Errorf("Cucumis sativus row %d non-CYP symbol %q", row, r.Symbol)
		}
		var status []string
		if pseudogeneRE.MatchString(r.Symbol) {
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
				return nil, fmt.Errorf("Cucumis sativus row %d unexpected residue %q", row, aa)
			}
		}
		r.Status = strings.Join(status, ";")
		accepted = append(accepted, r)
	}
	if len(accepted) != 229 {
		return nil, fmt.Errorf("Cucumis sativus boundary changed: accepted=%d", len(accepted))
	}
	return accepted, nil
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
		note := fmt.Sprintf("Cucumis sativus workbook row %d; Gotoh source ID=%s; seq ID=%s; sequence column=J", r.Row, r.GotohID, r.SeqID)
		_ = w.Write([]string{"plants", "Cucumis sativus", r.Symbol, r.SeqID, fmt.Sprintf("cucumis-sativus:row-%04d:%s", r.Row, r.SeqID), sourceURL, note, r.Sequence, sourceFile, hash, sheetName, fmt.Sprint(r.Row), r.Status})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, wb *plantxlsx.Workbook, hash string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	counts := map[string]int{}
	for _, r := range records {
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
	fmt.Fprintf(&b, "# Plant resource review: Cucumis sativus\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted rows: `%d`\n- Accepted rows with literal sequence: `%d`\n- Review status: `complete`\n\n", sourceFile, sourceURL, hash, sheetName, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(records), len(records))
	b.WriteString("## Resource-specific interpretation\n\nThis workbook contains one sheet named `sorted by CYP name` with `A1:J230`. The reviewed fields are A=`Gotoh's seq ID`, H=`seq ID`, I=`CYP name`, and J=`sequence`; rows 2-230 are all accepted because every row has a CYP symbol and a non-empty literal sequence. A is retained as source provenance while H is the output ID. The one leading cell-space in row 147 is removed as surrounding cell whitespace. Explicit terminal `*` markers are removed and audited; internal `*`, gaps, O, and X are retained exactly. Rows are not deduplicated and no sequence is repaired.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, name := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", name, counts[name])
	}
	b.WriteString("\n## Representative rows\n\n| Row | Gotoh ID | Seq ID | CYP name | Sequence | Status |\n|---:|---|---|---|---:|---|\n")
	for _, row := range []int{2, 116, 230} {
		for _, r := range records {
			if r.Row == row {
				fmt.Fprintf(&b, "| %d | %s | %s | %s | %d aa | %s |\n", r.Row, md(r.GotohID), md(r.SeqID), md(r.Symbol), len(r.Sequence), md(r.Status))
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
