package main

import (
	"flag"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"strings"
)

const sourceHash = "2fae1256d5cfcbbE9bdc9b1fa41c6ba10b7632447dd40b51d637ee7d6c3673c6"

var config = planttable.Config{SourceFile: "plants-Setaria.italica.xlsx", SourceURL: "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Setaria.italica.xlsx", Sheet: "Sorted by CYP name", Dimension: "A1:K419", Species: "Setaria italica", Slug: "setaria-italica", HeaderRow: 1, Headers: map[string]string{"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "SEQ ID", "H": "Best hit", "I": "%ID", "J": "CYP name", "K": ""}, DataStart: 2, DataEnd: 414, ExcludedStart: 415, ExcludedEnd: 419, GotohCol: "A", IDCol: "G", BestHitCol: "H", PercentCol: "I", SymbolCol: "J", SequenceCol: "K", PseudogeneField: "best-hit", AllowedResidues: "ACDEFGHIKLMNOPQRSTVWXY-", ExpectedAccepted: 413, ExpectedExcluded: 5, ExcludedReason: "no best hit, percent identity, or assigned CYP name; K is an out-of-frame translation", Interpretation: "The sole `Sorted by CYP name` sheet has exact used range `A1:K419`, no hidden rows or columns, merged cells, or native tables. G/J/K are the literal sequence ID, assigned CYP, and headerless protein columns. Rows 2-414 are the continuous assigned region. Rows 415-419 carry model text in A/G and out-of-frame translations containing O/X in K, but H/I/J are empty; they are excluded. The extra locus/strand text in A is checked to start with G rather than being mistaken for the ID itself. Literal O/X/gaps, short sequences, pseudogene hits and three duplicate-sequence groups remain row-distinct.", RepresentativeRows: []int{2, 208, 414}}

func review(w *plantxlsx.Workbook) ([]planttable.Record, []planttable.Record, error) {
	a, x, e := planttable.Review(config, w)
	if e != nil {
		return nil, nil, e
	}
	for _, r := range a {
		if !strings.HasPrefix(strings.TrimSpace(r.GotohID), r.ID+" ") {
			return nil, nil, fmt.Errorf("Setaria row %d A/G mismatch", r.Row)
		}
	}
	return a, x, nil
}
func main() {
	in := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Setaria workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "setaria-italica.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-setaria-italica.md"), "audit")
	flag.Parse()
	w, e := plantxlsx.Read(*in, config.Sheet)
	if e != nil {
		panic(e)
	}
	a, x, e := review(w)
	if e != nil {
		panic(e)
	}
	h, e := planttable.FileHash(*in)
	if e != nil {
		panic(e)
	}
	if !strings.EqualFold(h, sourceHash) {
		panic(fmt.Errorf("Setaria hash=%s", h))
	}
	if e = planttable.WriteCSV(*out, config, a, h); e != nil {
		panic(e)
	}
	if e = planttable.WriteAudit(*audit, config, a, x, w, h); e != nil {
		panic(e)
	}
	fmt.Printf("Setaria italica: %d assigned rows accepted, %d out-of-frame rows excluded\n", len(a), len(x))
}
