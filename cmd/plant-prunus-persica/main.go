package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var config = planttable.Config{
	SourceFile: "plants-Prunus.persica.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Prunus.persica.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L326",
	Species:    "Prunus persica",
	Slug:       "prunus-persica-updated",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 320, ExcludedStart: 321, ExcludedEnd: 326,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 319, ExpectedExcluded: 6,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name",
	Interpretation:     "The reviewed table uses row 1 as its exact header and rows 2-320 as the complete named region. A is the Gotoh source ID, H is the output sequence ID, I is the best hit, J is percent identity, K is the source-assigned CYP name, and L is the literal protein sequence. The output symbol is K rather than the more specific comparison value in I. Rows 321-326 have IDs and literal sequence cells but empty I/J/K assignments, so they are excluded as unnamed models. Gaps and O are retained. Four duplicate-sequence groups remain row-distinct. No sequence is repaired or deduplicated.",
	RepresentativeRows: []int{2, 161, 320},
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Prunus persica workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "prunus-persica-updated.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-prunus-persica-updated.md"), "review ledger")
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
	fmt.Printf("Prunus persica updated: %d named rows accepted, %d unnamed rows excluded\n", len(accepted), len(excluded))
}
