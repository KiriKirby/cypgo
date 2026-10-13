package planttable

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

type Config struct {
	SourceFile, SourceURL, Sheet, Dimension string
	Species, Slug                           string
	HeaderRow                               int
	Headers                                 map[string]string
	DataStart, DataEnd                      int
	ExcludedStart, ExcludedEnd              int
	GotohCol, IDCol, BestHitCol             string
	PercentCol, SymbolCol, SequenceCol      string
	PseudogeneField                         string
	AllowTerminalStop, AllowInternalStop    bool
	AllowedResidues                         string
	ExpectedAccepted, ExpectedExcluded      int
	ExcludedReason                          string
	Interpretation                          string
	RepresentativeRows                      []int
	ExtraStatus                             func(Record) []string
	AssignmentExceptionRows                 map[int]bool
	PercentExceptionRows                    map[int]bool
	ExcludedMayHaveBestHitPercent           bool
}

type Record struct {
	Row                                                       int
	GotohID, ID, BestHit, PercentID, Symbol, Sequence, Status string
}

var pseudogeneRE = regexp.MustCompile(`(?i)\dP$`)

func Review(config Config, wb *plantxlsx.Workbook) ([]Record, []Record, error) {
	if wb.Dimension != config.Dimension {
		return nil, nil, fmt.Errorf("%s used range changed: %q", config.Species, wb.Dimension)
	}
	if len(wb.HiddenRows) != 0 || len(wb.HiddenColumns) != 0 || len(wb.MergedCells) != 0 || wb.TableParts != 0 {
		return nil, nil, fmt.Errorf("%s workbook structure changed", config.Species)
	}
	if config.HeaderRow > 0 {
		h := wb.Rows[config.HeaderRow]
		for column, expected := range config.Headers {
			if h[column] != expected {
				return nil, nil, fmt.Errorf("%s header %s changed: %q", config.Species, column, h[column])
			}
		}
	}
	var accepted, excluded []Record
	for row := config.DataStart; row <= config.DataEnd; row++ {
		r, err := readRecord(config, wb, row)
		if err != nil {
			return nil, nil, err
		}
		if r.BestHit == "" || (r.PercentID == "" && !config.PercentExceptionRows[row]) || (!strings.HasPrefix(r.BestHit, "CYP") && !config.AssignmentExceptionRows[row]) || !strings.HasPrefix(r.Symbol, "CYP") {
			return nil, nil, fmt.Errorf("%s row %d missing reviewed CYP assignment", config.Species, row)
		}
		var status []string
		pseudogeneValue := r.Symbol
		pseudogeneStatus := "source-pseudogene-label"
		if config.PseudogeneField == "best-hit" {
			pseudogeneValue = r.BestHit
			pseudogeneStatus = "source-best-hit-pseudogene-label"
		}
		if pseudogeneRE.MatchString(pseudogeneValue) {
			status = append(status, pseudogeneStatus)
		}
		if strings.HasSuffix(r.Sequence, "*") {
			if !config.AllowTerminalStop {
				return nil, nil, fmt.Errorf("%s row %d unexpected terminal stop", config.Species, row)
			}
			r.Sequence = strings.TrimRight(r.Sequence, "*")
			status = append(status, "source-terminal-stop")
		}
		if strings.Contains(r.Sequence, "*") {
			if !config.AllowInternalStop {
				return nil, nil, fmt.Errorf("%s row %d unexpected internal stop", config.Species, row)
			}
			status = append(status, "internal-stop")
		}
		if strings.Contains(r.Sequence, "-") {
			status = append(status, "source-gap")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			status = append(status, "ambiguous-X-or-x")
		}
		if strings.Contains(r.Sequence, "O") {
			status = append(status, "nonstandard-O")
		}
		if len(r.Sequence) < 350 {
			status = append(status, "short-sequence")
		}
		for _, aa := range r.Sequence {
			if !strings.ContainsRune(config.AllowedResidues, aa) {
				return nil, nil, fmt.Errorf("%s row %d unexpected residue %q", config.Species, row, aa)
			}
		}
		if config.ExtraStatus != nil {
			status = append(status, config.ExtraStatus(r)...)
		}
		r.Status = strings.Join(status, ";")
		accepted = append(accepted, r)
	}
	for row := config.ExcludedStart; row > 0 && row <= config.ExcludedEnd; row++ {
		r, err := readRecord(config, wb, row)
		if err != nil {
			return nil, nil, err
		}
		if ((!config.ExcludedMayHaveBestHitPercent) && (r.BestHit != "" || r.PercentID != "")) || r.Symbol != "" {
			return nil, nil, fmt.Errorf("%s excluded row %d gained assignment", config.Species, row)
		}
		r.Status = "excluded: " + config.ExcludedReason
		excluded = append(excluded, r)
	}
	if len(accepted) != config.ExpectedAccepted || len(excluded) != config.ExpectedExcluded {
		return nil, nil, fmt.Errorf("%s boundary changed: accepted=%d excluded=%d", config.Species, len(accepted), len(excluded))
	}
	return accepted, excluded, nil
}

func readRecord(config Config, wb *plantxlsx.Workbook, row int) (Record, error) {
	v := wb.Rows[row]
	if v == nil {
		return Record{}, fmt.Errorf("missing %s row %d", config.Species, row)
	}
	r := Record{Row: row, GotohID: strings.TrimSpace(v[config.GotohCol]), ID: strings.TrimSpace(v[config.IDCol]), BestHit: strings.TrimSpace(v[config.BestHitCol]), PercentID: strings.TrimSpace(v[config.PercentCol]), Symbol: strings.TrimSpace(v[config.SymbolCol]), Sequence: strings.TrimSpace(v[config.SequenceCol])}
	if r.GotohID == "" || r.ID == "" || r.Sequence == "" {
		return Record{}, fmt.Errorf("%s row %d missing ID or sequence", config.Species, row)
	}
	return r, nil
}

func WriteCSV(path string, config Config, records []Record, hash string) error {
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
		note := fmt.Sprintf("%s workbook row %d; Gotoh source ID=%s; seq ID=%s; best hit=%s; %%ID=%s; assigned CYP column=%s; sequence column=%s", config.Species, r.Row, r.GotohID, r.ID, r.BestHit, r.PercentID, config.SymbolCol, config.SequenceCol)
		_ = w.Write([]string{"plants", config.Species, r.Symbol, r.ID, fmt.Sprintf("%s:row-%04d:%s", config.Slug, r.Row, r.ID), config.SourceURL, note, r.Sequence, config.SourceFile, hash, config.Sheet, fmt.Sprint(r.Row), r.Status})
	}
	w.Flush()
	return w.Error()
}

func WriteAudit(path string, config Config, accepted, excluded []Record, wb *plantxlsx.Workbook, hash string) error {
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
	fmt.Fprintf(&b, "# Plant resource review: %s\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Native table parts: `%d`\n- Accepted named rows: `%d`\n- Accepted rows with literal sequence: `%d`\n- Excluded rows: `%d`\n- Review status: `complete`\n\n", config.Species, config.SourceFile, config.SourceURL, hash, config.Sheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, wb.TableParts, len(accepted), len(accepted), len(excluded))
	b.WriteString("## Resource-specific interpretation\n\n" + config.Interpretation + "\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, name := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", name, counts[name])
	}
	b.WriteString("\n## Representative and excluded rows\n\n| Row | Seq ID | Best hit | Assigned CYP | Sequence | Status |\n|---:|---|---|---|---:|---|\n")
	for _, wanted := range config.RepresentativeRows {
		for _, r := range accepted {
			if r.Row == wanted {
				fmt.Fprintf(&b, "| %d | %s | %s | %s | %d aa | %s |\n", r.Row, md(r.ID), md(r.BestHit), md(r.Symbol), len(r.Sequence), md(r.Status))
			}
		}
	}
	for _, r := range excluded {
		fmt.Fprintf(&b, "| %d | %s |  |  | %d chars (excluded) | %s |\n", r.Row, md(r.ID), len(r.Sequence), md(r.Status))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func FileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func md(value string) string { return strings.ReplaceAll(strings.TrimSpace(value), "|", "\\|") }
