package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceHash = "8974bf1f2cd5b439dac6c800dea9791882142ef12ac775d7cb58fa979aface2a"

var config = planttable.Config{
	SourceFile: "plants-Cajanus.cajanifolius.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Cajanus.cajanifolius.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L309",
	Species:    "Cajanus cajanifolius",
	Slug:       "cajanus-cajanifolius",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 292, ExcludedStart: 293, ExcludedEnd: 309,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 291, ExpectedExcluded: 17,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The workbook contains one worksheet with `A1:L309` and no hidden rows or columns, merged cells, or native tables. Row 1 names H=`seq. ID`, I=`best hit`, J=`%ID`, and K=`CYP name`; L is the headerless literal sequence column. Rows 2-292 are the complete continuous assigned region and A exactly matches H for every accepted row. Rows 293-309 retain IDs and model text but have empty I/J/K assignment fields and are excluded. The 291 accepted rows preserve three gap-bearing sequences, literal X in three records, literal O in 11 records, 11 short sequences, one digit-P best-hit label, and two duplicate-sequence groups whose rows remain distinct. No accepted ID is duplicated, and no sequence is inferred or repaired.",
	RepresentativeRows: []int{2, 147, 292},
}

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range accepted {
		if r.GotohID != r.ID {
			return nil, nil, fmt.Errorf("Cajanus cajanifolius row %d A/H ID mismatch: %q != %q", r.Row, r.GotohID, r.ID)
		}
	}
	return accepted, excluded, nil
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Cajanus cajanifolius workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "cajanus-cajanifolius.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-cajanus-cajanifolius.md"), "review ledger")
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
		panic(fmt.Errorf("Cajanus cajanifolius source hash changed: %s", hash))
	}
	if err := planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err := planttable.WriteAudit(*audit, config, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Cajanus cajanifolius: %d assigned rows accepted, %d unassigned rows excluded\n", len(accepted), len(excluded))
}
