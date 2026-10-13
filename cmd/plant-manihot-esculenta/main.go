package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceHash = "b55399ee3930a89618d3272e9ec51d5934487d76e1e29e0a6ff378c53e0c333c"

var config = planttable.Config{
	SourceFile: "plants-Manihot.esculenta.seqs.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Manihot.esculenta.seqs.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L349",
	Species:    "Manihot esculenta",
	Slug:       "manihot-esculenta",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "", "H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 337, ExcludedStart: 338, ExcludedEnd: 349,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 336, ExpectedExcluded: 12,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The workbook contains one worksheet with `A1:L349` and no hidden rows or columns, merged cells, native tables, formulas, or comments. Row 1 names H=`seq ID`, I=`best hit`, J=`%ID`, and K=`CYP name`; L is the headerless literal sequence column. Rows 2-337 are the complete continuous assigned region and A exactly matches H for every accepted row. Rows 338-349 retain IDs and model-like text but have empty I/J/K assignment fields and are excluded. The 336 accepted rows preserve four gap-bearing sequences, literal X in 16 records, literal O in five records, 25 short sequences, five digit-P best-hit labels, and two duplicate-sequence groups whose rows remain distinct. No accepted ID is duplicated, and no sequence is inferred or repaired.",
	RepresentativeRows: []int{2, 170, 337},
}

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range accepted {
		if r.GotohID != r.ID {
			return nil, nil, fmt.Errorf("Manihot esculenta row %d A/H ID mismatch: %q != %q", r.Row, r.GotohID, r.ID)
		}
	}
	return accepted, excluded, nil
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Manihot esculenta workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "manihot-esculenta.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-manihot-esculenta.md"), "review ledger")
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
		panic(fmt.Errorf("Manihot esculenta source hash changed: %s", hash))
	}
	if err := planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err := planttable.WriteAudit(*audit, config, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Manihot esculenta: %d assigned rows accepted, %d unassigned rows excluded\n", len(accepted), len(excluded))
}
