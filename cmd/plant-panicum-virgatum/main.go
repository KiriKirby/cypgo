package main

import (
	"flag"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"strings"
)

const sourceHash = "138182c4d9a9f920d3830d34cc35c1b11bee035d49da7f8966492807113f91be"

var config = planttable.Config{SourceFile: "plants-Panicum.virgatum.xlsx", SourceURL: "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Panicum.virgatum.xlsx", Sheet: "sorted by CYP name", Dimension: "A1:G810", Species: "Panicum virgatum", Slug: "panicum-virgatum", HeaderRow: 1, Headers: map[string]string{"A": "best hit ", "B": "%ID", "C": "gene ID", "D": "length", "E": "gene ID", "F": "CYP name", "G": "sequence"}, DataStart: 2, DataEnd: 807, GotohCol: "C", IDCol: "E", BestHitCol: "A", PercentCol: "B", SymbolCol: "F", SequenceCol: "G", PseudogeneField: "best-hit", AllowedResidues: "ACDEFGHIKLMNOPQRSTVWXY-", ExpectedAccepted: 806, Interpretation: "The workbook contains two full data sheets: `Sorted by length` (`A1:F807`) and `sorted by CYP name` (`A1:G810`). They are alternate orderings, not two releases, so only the CYP-name sheet contributes records. Rows 2-807 contain 806 assigned literal G-column sequences; rows 808-809 are blank and row 810 is the legend `green <55% identical to a named P450`. Every D length equals the literal G length. Cross-sheet multiset comparison matches 802 rows exactly and isolates four source revisions/differences: Panivirg326850.2 rev, Panivirg327867.1, Panivirg356643.1rev, and Panivirg23744.8 (the name sheet assigns CYP76N1P while the length sheet says `not a P450`). The CYP-name sheet is the reviewed release authority. Ninety X-bearing and 189 short sequences remain literal.", RepresentativeRows: []int{2, 404, 807}}

func review(named, length *plantxlsx.Workbook) ([]planttable.Record, error) {
	a, _, e := planttable.Review(config, named)
	if e != nil {
		return nil, e
	}
	for _, r := range a {
		if fmt.Sprint(len(r.Sequence)) != strings.TrimSpace(named.Rows[r.Row]["D"]) {
			return nil, fmt.Errorf("Panicum row %d length mismatch", r.Row)
		}
	}
	lm := map[string]int{}
	nm := map[string]int{}
	for i := 2; i <= 807; i++ {
		r := length.Rows[i]
		lm[strings.Join([]string{strings.TrimSpace(r["A"]), strings.TrimSpace(r["B"]), strings.TrimSpace(r["C"]), strings.TrimSpace(r["D"]), strings.TrimSpace(r["E"]), strings.TrimSpace(r["F"])}, "\x1f")]++
		r = named.Rows[i]
		nm[strings.Join([]string{strings.TrimSpace(r["A"]), strings.TrimSpace(r["B"]), strings.TrimSpace(r["C"]), strings.TrimSpace(r["D"]), strings.TrimSpace(r["E"]), strings.TrimSpace(r["G"])}, "\x1f")]++
	}
	dl, dn := 0, 0
	for k, v := range lm {
		if nm[k] != v {
			dl++
		}
	}
	for k, v := range nm {
		if lm[k] != v {
			dn++
		}
	}
	if dl != 4 || dn != 4 {
		return nil, fmt.Errorf("Panicum cross-sheet differences=%d/%d", dl, dn)
	}
	return a, nil
}
func main() {
	in := flag.String("input", filepath.Join("raw", config.SourceFile), "downloaded Panicum workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "panicum-virgatum.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-panicum-virgatum.md"), "audit")
	flag.Parse()
	names, e := plantxlsx.SheetNames(*in)
	if e != nil || len(names) != 2 || names[0] != "Sorted by length" || names[1] != config.Sheet {
		panic(fmt.Errorf("Panicum sheets=%v: %v", names, e))
	}
	named, e := plantxlsx.ReadSheet(*in, config.Sheet, false)
	if e != nil {
		panic(e)
	}
	length, e := plantxlsx.ReadSheet(*in, "Sorted by length", false)
	if e != nil {
		panic(e)
	}
	if length.Dimension != "A1:F807" {
		panic("length sheet changed")
	}
	a, e := review(named, length)
	if e != nil {
		panic(e)
	}
	h, e := planttable.FileHash(*in)
	if e != nil {
		panic(e)
	}
	if h != sourceHash {
		panic(fmt.Errorf("Panicum hash=%s", h))
	}
	if e = planttable.WriteCSV(*out, config, a, h); e != nil {
		panic(e)
	}
	if e = planttable.WriteAudit(*audit, config, a, nil, named, h); e != nil {
		panic(e)
	}
	fmt.Printf("Panicum virgatum: %d CYP-name-sheet rows accepted; length sheet used only for cross-check\n", len(a))
}
