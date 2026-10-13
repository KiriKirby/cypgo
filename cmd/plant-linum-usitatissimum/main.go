package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceHash = "02b59af9f3f01970bb0b27ac80af101c64c68e71e96761caa93bcc6cdd1f8f64"

var config = planttable.Config{
	SourceFile: "plants-Linum.usitatissimum.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Linum.usitatissimum.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L479",
	Species:    "Linum usitatissimum",
	Slug:       "linum-usitatissimum",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "", "H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 469, ExcludedStart: 470, ExcludedEnd: 479,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 468, ExpectedExcluded: 10,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The workbook contains one worksheet with `A1:L479` and no hidden rows or columns, merged cells, native tables, formulas, or comments. Row 1 names H=`seq ID`, I=`best hit`, J=`%ID`, and K=`CYP name`; L is the headerless literal sequence column. Rows 2-469 are the complete continuous assigned region and A exactly matches H for every accepted row. Rows 470-479 retain IDs and model-like text but have empty I/J/K assignment fields and are excluded. The 468 accepted rows preserve six gap-bearing sequences, literal X in 13 records, literal O in 25 records, 19 short sequences, four digit-P best-hit labels, and two duplicate-sequence groups whose rows remain distinct. No accepted ID is duplicated, and no sequence is inferred or repaired.",
	RepresentativeRows: []int{2, 236, 469},
}

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range accepted {
		if r.GotohID != r.ID {
			return nil, nil, fmt.Errorf("Linum usitatissimum row %d A/H ID mismatch: %q != %q", r.Row, r.GotohID, r.ID)
		}
	}
	return accepted, excluded, nil
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Linum usitatissimum workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "linum-usitatissimum.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-linum-usitatissimum.md"), "review ledger")
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
		panic(fmt.Errorf("Linum usitatissimum source hash changed: %s", hash))
	}
	if err := planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err := planttable.WriteAudit(*audit, config, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Linum usitatissimum: %d assigned rows accepted, %d unassigned rows excluded\n", len(accepted), len(excluded))
}
