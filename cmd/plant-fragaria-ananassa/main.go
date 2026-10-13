package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var config = planttable.Config{
	SourceFile: "plants-Fragaria.ananassa.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Fragaria.ananassa.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L225",
	Species:    "Fragaria ananassa",
	Slug:       "fragaria-ananassa",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "Gotoh ID", "H": "Blast output seq ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "sequence",
	},
	DataStart: 2, DataEnd: 207, ExcludedStart: 208, ExcludedEnd: 225,
	GotohCol: "A", IDCol: "A", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 206, ExpectedExcluded: 18,
	ExcludedReason:     "source marks the model as opposite strand with no hit and no CYP assignment",
	Interpretation:     "This workbook is not parsed with the ordinary H-column ID rule. Row 1 explicitly names A as `Gotoh ID`, and every A value in rows 2-225 begins with `>`. For accepted rows 2-206, H repeats A without `>`; row 207 is a real assigned CYP88A50/CYP88A record whose H cell instead says `opposite strand`. The reviewed output therefore uses A with exactly one leading `>` removed as the stable ID for every accepted row. Rows 208-225 also say `opposite strand`, but I and K are both `no hit` and J is empty, so those unassigned models are excluded. All 206 assigned rows keep their literal L-column sequences, including 16 gaps, literal X/O, 62 short fragments, and two digit-P best-hit labels. No sequence is inferred, repaired, or deduplicated.",
	RepresentativeRows: []int{2, 104, 207},
}

var digitP = regexp.MustCompile(`(?i)\dP$`)

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	if wb.Dimension != config.Dimension {
		return nil, nil, fmt.Errorf("used range changed: %q", wb.Dimension)
	}
	if len(wb.HiddenRows) != 0 || len(wb.HiddenColumns) != 0 || len(wb.MergedCells) != 0 || wb.TableParts != 0 {
		return nil, nil, fmt.Errorf("workbook structure changed")
	}
	for column, expected := range config.Headers {
		if wb.Rows[1][column] != expected {
			return nil, nil, fmt.Errorf("header %s changed: %q", column, wb.Rows[1][column])
		}
	}
	var accepted, excluded []planttable.Record
	for row := config.DataStart; row <= config.DataEnd; row++ {
		v := wb.Rows[row]
		rawID := strings.TrimSpace(v["A"])
		if !strings.HasPrefix(rawID, ">") || strings.HasPrefix(rawID, ">>") {
			return nil, nil, fmt.Errorf("row %d changed Gotoh ID form: %q", row, rawID)
		}
		id := strings.TrimPrefix(rawID, ">")
		h := strings.TrimSpace(v["H"])
		if row == 207 {
			if h != "opposite strand" {
				return nil, nil, fmt.Errorf("row 207 strand marker changed: %q", h)
			}
		} else if h != id {
			return nil, nil, fmt.Errorf("row %d A/H ID relationship changed: %q / %q", row, rawID, h)
		}
		r := planttable.Record{Row: row, GotohID: rawID, ID: id, BestHit: strings.TrimSpace(v["I"]), PercentID: strings.TrimSpace(v["J"]), Symbol: strings.TrimSpace(v["K"]), Sequence: strings.TrimSpace(v["L"])}
		if r.ID == "" || r.PercentID == "" || r.Sequence == "" || !strings.HasPrefix(r.BestHit, "CYP") || !strings.HasPrefix(r.Symbol, "CYP") {
			return nil, nil, fmt.Errorf("row %d missing reviewed CYP assignment, ID, or sequence", row)
		}
		var status []string
		if digitP.MatchString(r.BestHit) {
			status = append(status, "source-best-hit-pseudogene-label")
		}
		if strings.Contains(r.Sequence, "*") {
			return nil, nil, fmt.Errorf("row %d unexpected stop marker", row)
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
				return nil, nil, fmt.Errorf("row %d unexpected residue %q", row, aa)
			}
		}
		r.Status = strings.Join(status, ";")
		accepted = append(accepted, r)
	}
	for row := config.ExcludedStart; row <= config.ExcludedEnd; row++ {
		v := wb.Rows[row]
		rawID := strings.TrimSpace(v["A"])
		seq := strings.TrimSpace(v["L"])
		if !strings.HasPrefix(rawID, ">") || strings.TrimSpace(v["H"]) != "opposite strand" || strings.TrimSpace(v["I"]) != "no hit" || strings.TrimSpace(v["J"]) != "" || strings.TrimSpace(v["K"]) != "no hit" || seq == "" {
			return nil, nil, fmt.Errorf("excluded row %d layout changed", row)
		}
		excluded = append(excluded, planttable.Record{Row: row, GotohID: rawID, ID: strings.TrimPrefix(rawID, ">"), BestHit: "no hit", Symbol: "no hit", Sequence: seq, Status: "excluded: " + config.ExcludedReason})
	}
	if len(accepted) != config.ExpectedAccepted || len(excluded) != config.ExpectedExcluded {
		return nil, nil, fmt.Errorf("boundary changed: accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	return accepted, excluded, nil
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Fragaria ananassa workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "fragaria-ananassa.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-fragaria-ananassa.md"), "review ledger")
	flag.Parse()
	wb, err := plantxlsx.Read(*input, config.Sheet)
	if err != nil {
		panic(err)
	}
	accepted, excluded, err := review(wb)
	if err != nil {
		panic(err)
	}
	hash, err := planttable.FileHash(*input)
	if err != nil {
		panic(err)
	}
	if err := planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err := planttable.WriteAudit(*audit, config, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Fragaria ananassa: %d assigned rows accepted, %d no-hit rows excluded\n", len(accepted), len(excluded))
}
