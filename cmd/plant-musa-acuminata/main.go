package main

import (
	"flag"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"strings"
)

const sourceHash = "e2941945d5e16e4ca5cf129560ab75319aa5e5f1c1ff6093464c267ce8eda3d2"

var config = planttable.Config{SourceFile: "plants-Musa.acuminata.xlsx", SourceURL: "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Musa.acuminata.xlsx", Sheet: "Sorted by CYP name", Dimension: "A1:M233", Species: "Musa acuminata", Slug: "musa-acuminata", DataStart: 1, DataEnd: 233, GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "L", SequenceCol: "M", PseudogeneField: "best-hit", AllowedResidues: "ACDEFGHIKLMNOPQRSTVWXY-", ExpectedAccepted: 233, AssignmentExceptionRows: map[int]bool{35: true, 66: true, 76: true, 120: true}, PercentExceptionRows: map[int]bool{35: true, 66: true, 76: true, 120: true}, Interpretation: "This workbook is headerless: row 1 is already the first data row. All 233 rows are assigned. A contains a FASTA-prefixed ID plus layout text, H the exact ID, I/J the best hit and identity, L the assigned CYP, and M the literal protein. Rows 35, 66, 76 and 120 explicitly carry broad `CYP... pseudo` labels and no percent identity; those literal labels are retained rather than rejected or made more specific. X, one gap, short proteins, a duplicate-sequence pair and all row distinctions are retained.", RepresentativeRows: []int{1, 117, 233}}

func review(w *plantxlsx.Workbook) ([]planttable.Record, error) {
	a, _, e := planttable.Review(config, w)
	if e != nil {
		return nil, e
	}
	for _, r := range a {
		baseID := strings.Fields(r.ID)[0]
		if !strings.HasPrefix(r.GotohID, ">"+baseID) {
			return nil, fmt.Errorf("Musa row %d ID mismatch", r.Row)
		}
	}
	want := map[int]string{35: "CYP71 pseudo", 66: "CYP74A pseudo", 76: "CYP75B pseudo", 120: "CYP86A pseudo"}
	for _, r := range a {
		if v, ok := want[r.Row]; ok && (r.BestHit != v || r.Symbol != v || r.PercentID != "") {
			return nil, fmt.Errorf("Musa pseudo row %d changed", r.Row)
		}
	}
	return a, nil
}
func main() {
	in := flag.String("input", filepath.Join("raw", config.SourceFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "musa-acuminata.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-musa-acuminata.md"), "")
	flag.Parse()
	w, e := plantxlsx.Read(*in, config.Sheet)
	if e != nil {
		panic(e)
	}
	a, e := review(w)
	if e != nil {
		panic(e)
	}
	h, e := planttable.FileHash(*in)
	if e != nil || h != sourceHash {
		panic(fmt.Errorf("hash %s %v", h, e))
	}
	if e = planttable.WriteCSV(*out, config, a, h); e != nil {
		panic(e)
	}
	if e = planttable.WriteAudit(*audit, config, a, nil, w, h); e != nil {
		panic(e)
	}
	fmt.Printf("Musa acuminata: %d rows accepted\n", len(a))
}
