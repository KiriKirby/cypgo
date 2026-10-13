package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var config = planttable.Config{
	SourceFile: "plants-Malus.domestica.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Malus.domestica.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L349",
	Species:    "Malus domestica",
	Slug:       "malus-domestica",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 332, ExcludedStart: 333, ExcludedEnd: 349,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 331, ExpectedExcluded: 17,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The workbook contains one worksheet with no hidden rows or columns, merged cells, or native tables. Row 1 is the exact header. Rows 2-332 form the complete assigned region: A is the Gotoh source ID, H is the sequence ID, I is the best hit, J is percent identity, K is the source-assigned CYP label, and L is the literal protein sequence. Rows 333-349 retain IDs and sequence-like model text but have empty I/J/K assignment fields, so they are excluded. The accepted source contains extensive literal X and O residues, two gaps, eleven short sequences, four digit-P best-hit labels, one duplicated ID with two different sequences, and two duplicate-sequence pairs with distinct IDs. Broad source labels such as `CYP` are retained. No sequence is repaired or deduplicated.",
	RepresentativeRows: []int{2, 167, 332},
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Malus domestica workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "malus-domestica.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-malus-domestica.md"), "review ledger")
	flag.Parse()
	wb, err := plantxlsx.Read(*input, config.Sheet)
	if err != nil {
		panic(err)
	}
	accepted, excluded, err := planttable.Review(config, wb)
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
	fmt.Printf("Malus domestica: %d named rows accepted, %d unnamed rows excluded\n", len(accepted), len(excluded))
}
