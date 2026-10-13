package main

import (
	"flag"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
)

const sourceHash = "85d4d51a03b061e4bf314eda7341b81c3731855afa459ceb4cc0f422a45dd5c0"

var config = planttable.Config{SourceFile: "plants-Phoenix.dactylifera.xlsx", SourceURL: "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Phoenix.dactylifera.xlsx", Sheet: "Soreted by CYP name", Dimension: "A1:L232", Species: "Phoenix dactylifera", Slug: "phoenix-dactylifera", HeaderRow: 1, Headers: map[string]string{"A": "", "B": "", "C": "", "D": "", "E": "", "F": "", "G": "", "H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": ""}, DataStart: 2, DataEnd: 210, ExcludedStart: 211, ExcludedEnd: 232, GotohCol: "A", IDCol: "H", BestHitCol: "I", PercentCol: "J", SymbolCol: "K", SequenceCol: "L", PseudogeneField: "best-hit", AllowedResidues: "ACDEFGHIKLMNOPQRSTVWXY-", ExpectedAccepted: 209, ExpectedExcluded: 22, ExcludedReason: "no best hit, percent identity, or assigned CYP name; L is an out-of-frame translation", Interpretation: "The single sheet name is literally misspelled `Soreted by CYP name`. Rows 2-210 contain 209 assigned L-column proteins. Rows 211-232 retain IDs and out-of-frame translations but have empty I/J/K and are excluded. Literal O/X, gaps, short sequences and one duplicate-sequence pair remain unchanged.", RepresentativeRows: []int{2, 106, 210}}

func main() {
	in := flag.String("input", filepath.Join("raw", config.SourceFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "phoenix-dactylifera.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-phoenix-dactylifera.md"), "")
	flag.Parse()
	w, e := plantxlsx.Read(*in, config.Sheet)
	if e != nil {
		panic(e)
	}
	a, x, e := planttable.Review(config, w)
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
	if e = planttable.WriteAudit(*audit, config, a, x, w, h); e != nil {
		panic(e)
	}
	fmt.Printf("Phoenix dactylifera: %d accepted, %d excluded\n", len(a), len(x))
}
