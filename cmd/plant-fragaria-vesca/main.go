package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var config = planttable.Config{
	SourceFile: "plants-Fragaria.vesca.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Fragaria.vesca.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L336",
	Species:    "Fragaria vesca",
	Slug:       "fragaria-vesca",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 332, ExcludedStart: 333, ExcludedEnd: 336,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY",
	ExpectedAccepted: 331, ExpectedExcluded: 4,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The workbook has one worksheet and no hidden rows or columns, merged cells, or native tables. Row 1 is the exact header. Rows 2-332 are the complete assigned region, with A as Gotoh source ID, H as sequence ID, I as best hit, J as percent identity, K as source-assigned CYP label, and L as literal protein sequence. Rows 333-336 retain IDs and sequence-like model text but have empty I/J/K assignment cells, so they are excluded. Literal O/X residues, eight short sequences, one digit-P best-hit label, and three duplicate-sequence pairs with distinct IDs are preserved. No sequence is repaired or deduplicated.",
	RepresentativeRows: []int{2, 168, 332},
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Fragaria vesca workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "fragaria-vesca.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-fragaria-vesca.md"), "review ledger")
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
	fmt.Printf("Fragaria vesca: %d named rows accepted, %d unnamed rows excluded\n", len(accepted), len(excluded))
}
