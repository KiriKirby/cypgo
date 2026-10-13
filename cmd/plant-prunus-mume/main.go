package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var config = planttable.Config{
	SourceFile: "plants-Prunus.mume.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Prunus.mume.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L288",
	Species:    "Prunus mume",
	Slug:       "prunus-mume",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "H": "seq. ID", "I": "best hit ", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 283, ExcludedStart: 284, ExcludedEnd: 288,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 282, ExpectedExcluded: 5,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The reviewed workbook has one worksheet and no hidden rows or columns, merged cells, or native tables. Row 1 is the exact header, including the trailing space in the I-column heading `best hit `. Rows 2-283 are the complete assigned region: A is the Gotoh source ID, H is the sequence ID, I is the best hit, J is percent identity, K is the source-assigned CYP label, and L is the literal protein sequence. Rows 281-283 deliberately retain the broad source label `CYP` because the workbook supplies that assignment together with a specific best hit and percent identity. Rows 284-288 have IDs and sequence-like text but empty I/J/K assignments and extensive non-protein O/X content, so they are excluded as unnamed models. One gap, literal O/X residues, eleven short sequences, two digit-P best-hit labels, and four duplicate-sequence pairs remain exactly as supplied. No sequence is repaired or deduplicated.",
	RepresentativeRows: []int{2, 140, 283},
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Prunus mume workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "prunus-mume.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-prunus-mume.md"), "review ledger")
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
	fmt.Printf("Prunus mume: %d named rows accepted, %d unnamed rows excluded\n", len(accepted), len(excluded))
}
