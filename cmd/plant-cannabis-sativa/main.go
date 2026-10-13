package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

var config = planttable.Config{
	SourceFile: "plants-Cannabis.sativa.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Cannabis.sativa.xlsx",
	Sheet:      "Sorted by CYP name",
	Dimension:  "A1:L423",
	Species:    "Cannabis sativa",
	Slug:       "cannabis-sativa",
	HeaderRow:  1,
	Headers: map[string]string{
		"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "",
	},
	DataStart: 2, DataEnd: 357, ExcludedStart: 358, ExcludedEnd: 421,
	GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L",
	PseudogeneField:  "best-hit",
	AllowedResidues:  "ACDEFGHIKLMNOPQRSTVWXY-",
	ExpectedAccepted: 356, ExpectedExcluded: 64,
	ExcludedReason:     "no best hit, percent identity, or assigned CYP name; source legend identifies 64 no-hit out-of-frame sequences",
	Interpretation:     "The workbook contains one worksheet with no hidden rows or columns, merged cells, or native tables. Rows 2-357 are the complete assigned region, using H as sequence ID, I as best hit, J as percent identity, K as source-assigned CYP label, and L as literal protein sequence. Rows 358-421 are exactly the 64 unassigned models described by the source legend in E423 as `green name = no hits in blast (64 seqs. out of frame)`; their I/J/K cells are empty, so they are excluded. Row 422 is blank. The 356 accepted rows retain 25 gaps, literal X/O, 100 short sequences, three digit-P best-hit labels, and one duplicate-sequence pair with distinct IDs. No sequence is repaired or deduplicated.",
	RepresentativeRows: []int{2, 179, 357},
}

func review(wb *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	if len(wb.Rows[422]) != 0 || wb.Rows[423]["E"] != "green name = no hits in blast (64 seqs. out of frame)" {
		return nil, nil, fmt.Errorf("trailing blank row or source legend changed")
	}
	return planttable.Review(config, wb)
}

func main() {
	input := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Cannabis sativa workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "cannabis-sativa.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-cannabis-sativa.md"), "review ledger")
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
	fmt.Printf("Cannabis sativa: %d named rows accepted, %d no-hit rows excluded\n", len(accepted), len(excluded))
}
