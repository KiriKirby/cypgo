package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceHash = "7ab761fa617a91f6e7ddac0ba059b13cca8b051b1125974710cccf468078bd75"

var config = planttable.Config{
	SourceFile: "plants-Cicer.arietinum.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Cicer.arietinum.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L231",
	Species:    "Cicer arietinum",
	Slug:       "cicer-arietinum",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 212, ExcludedStart: 213, ExcludedEnd: 231,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 211, ExpectedExcluded: 19,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The workbook contains one worksheet with `A1:L231` and no hidden rows or columns, merged cells, or native tables. Row 1 names H=`seq. ID`, I=`best hit`, J=`%ID`, and K=`CYP name`; L is the headerless literal sequence column. Rows 2-212 are the complete continuous assigned region and A exactly matches H for every accepted row. Rows 213-231 retain IDs and model text but have empty I/J/K assignment fields and are excluded. The 211 accepted rows preserve six gap-bearing sequences, literal X in five records, literal O in eight records, 22 short sequences, and three digit-P best-hit labels. No accepted ID or literal sequence is duplicated, and no sequence is inferred or repaired.",
	RepresentativeRows: []int{2, 107, 212},
}

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		return nil, nil, err
	}
	for _, r := range accepted {
		if r.GotohID != r.ID {
			return nil, nil, fmt.Errorf("Cicer arietinum row %d A/H ID mismatch: %q != %q", r.Row, r.GotohID, r.ID)
		}
	}
	return accepted, excluded, nil
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Cicer arietinum workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "cicer-arietinum.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-cicer-arietinum.md"), "review ledger")
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
		panic(fmt.Errorf("Cicer arietinum source hash changed: %s", hash))
	}
	if err := planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err := planttable.WriteAudit(*audit, config, accepted, excluded, wb, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Cicer arietinum: %d assigned rows accepted, %d unassigned rows excluded\n", len(accepted), len(excluded))
}
