package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const sourceHash = "c5ac3e9051338757b5e4d002a7e4cca594ab6e30937226f2db976f0488581a78"

var config = planttable.Config{
	SourceFile: "plants-Amborella.2014.pub.xlsx",
	SourceURL:  "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/2021/04/Amborella.2014.pub.xlsx",
	Sheet:      "Sheet1", Dimension: "A1:F231", Species: "Amborella trichopoda", Slug: "amborella-trichopoda",
	HeaderRow: 1, Headers: map[string]string{"A": "Scaffold", "B": "Location", "C": "best hit", "D": "%ID", "E": "CYP name", "F": "sequence"},
	DataStart: 2, DataEnd: 231, GotohCol: "A", IDCol: "A", BestHitCol: "C", PercentCol: "D", SymbolCol: "E", SequenceCol: "F",
	PseudogeneField: "symbol", AllowInternalStop: true, AllowedResidues: "ACDEFGHIKLMNOPQRSTVWXY*", ExpectedAccepted: 230,
	AssignmentExceptionRows: map[int]bool{29: true},
	Interpretation:          "The index Word document links three parts of the Amborella release. The old annotated document contains 195 pre-name blocks (its preface describes 164 genome sequences followed by GenBank ESTs) and is retained only as historical annotation, not imported beside the newer named set. The 2014 workbook has one visible, unmerged sheet with 230 assigned rows. Its A/C/D/E/F columns contain scaffold ID, best hit, percent identity, final CYP name and literal protein respectively. The separately linked named FASTA contains exactly the same 230 ID/name/sequence triples in exactly the same order, so it is an independent byte-for-byte biological cross-check rather than a second record source. Internal stops, O, X, fragments, pseudogenes, short sequences and duplicate sequences remain literal; no sequence is repaired, translated or deduplicated.",
	RepresentativeRows:      []int{2, 116, 231},
	ExtraStatus: func(r planttable.Record) []string {
		var out []string
		if strings.Contains(strings.ToLower(r.Symbol), "fragment") {
			out = append(out, "source-fragment-label")
		}
		return out
	},
}

func review(w *plantxlsx.Workbook) ([]planttable.Record, error) {
	accepted, _, err := planttable.Review(config, w)
	if err != nil {
		return nil, err
	}
	return accepted, nil
}

func main() {
	in := flag.String("input", filepath.Join("raw", config.SourceFile), "reviewed Amborella workbook")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "amborella-trichopoda.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-amborella-trichopoda.md"), "resource audit")
	flag.Parse()
	w, err := plantxlsx.Read(*in, config.Sheet)
	if err != nil {
		panic(err)
	}
	accepted, err := review(w)
	if err != nil {
		panic(err)
	}
	hash, err := planttable.FileHash(*in)
	if err != nil || hash != sourceHash {
		panic(fmt.Errorf("Amborella source hash changed: %s: %v", hash, err))
	}
	if err = planttable.WriteCSV(*out, config, accepted, hash); err != nil {
		panic(err)
	}
	if err = planttable.WriteAudit(*audit, config, accepted, nil, w, hash); err != nil {
		panic(err)
	}
	fmt.Printf("Amborella trichopoda: %d rows accepted\n", len(accepted))
}
