package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceHash = "cc1b3bc1a447e3e5d5e4961d019698503a4db80ff60ed960ba97098e7c72f59f"

var config = planttable.Config{
	SourceFile: "plants-Oryza.brachyantha.xlsx", SourceURL: "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Oryza.brachyantha.xlsx",
	Sheet: "sorted by CYP name", Dimension: "A1:L321", Species: "Oryza brachyantha", Slug: "oryza-brachyantha",
	HeaderRow: 1,
	Headers:   map[string]string{"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "", "H": "Gotoh ID", "I": "Best hit", "J": "%ID", "K": "CYP name", "L": ""},
	DataStart: 2, DataEnd: 317, ExcludedStart: 318, ExcludedEnd: 321,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField: "best-hit", AllowedResidues: "ACDEFGHIKLMNOPQRSTVWXY-",
	AllowTerminalStop:             true,
	AssignmentExceptionRows:       map[int]bool{20: true, 158: true, 159: true, 199: true, 245: true},
	ExcludedMayHaveBestHitPercent: true,
	ExpectedAccepted:              316, ExpectedExcluded: 4,
	ExcludedReason:     "out-of-frame translation with no assigned CYP name",
	Interpretation:     "The workbook has exactly one worksheet, `sorted by CYP name`, with used range `A1:L321` and no hidden rows or columns, merged cells, or native tables. Row 1 names H=`Gotoh ID`, I=`Best hit`, J=`%ID`, and K=`CYP name`; L is the headerless literal protein column. Rows 2-317 form the continuous assigned region and A (after its source FASTA marker is removed) equals H. Rows 318-321 explicitly say `out of frame translation overlaps ...`, have J=`0` and an empty K assignment, and are excluded even though L contains out-of-frame letter strings. Literal O/X and source truncation remain unchanged; no row is repaired or supplemented.",
	RepresentativeRows: []int{2, 160, 317},
}

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range accepted {
		if len(r.GotohID) < 2 || r.GotohID[0] != '>' || r.GotohID[1:] != r.ID {
			return nil, nil, fmt.Errorf("Oryza brachyantha row %d A/H ID mismatch: %q != %q", r.Row, r.GotohID, r.ID)
		}
	}
	want := map[int][3]string{20: {"out of frame translation novel P450", "76", "CYP71E"}, 158: {"out of frame translation REVISED SEQ", "79", "CYP81N2"}, 159: {"out of frame translation REVISED SEQ", "81", "CYP81N2"}, 199: {"out of frame translation, revised seq", "0", "CYP90D"}, 245: {"out of frame, novel seq", "revised seq", "CYP96D"}}
	for _, r := range accepted {
		if v, ok := want[r.Row]; ok && (r.BestHit != v[0] || r.PercentID != v[1] || r.Symbol != v[2]) {
			return nil, nil, fmt.Errorf("Oryza exception row %d changed", r.Row)
		}
	}
	return accepted, excluded, nil
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Oryza brachyantha workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "oryza-brachyantha.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-oryza-brachyantha.md"), "review ledger")
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
	if hash != sourceHash {
		panic(fmt.Errorf("Oryza brachyantha source hash changed: %s", hash))
	}
	if err := planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err := planttable.WriteAudit(*audit, config, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Oryza brachyantha: %d assigned rows accepted, %d out-of-frame rows excluded\n", len(accepted), len(excluded))
}
