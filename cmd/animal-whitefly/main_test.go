package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rs, lines, e := parse("../../raw/animals-whitefly.txt")
	if e != nil {
		t.Fatal(e)
	}
	if lines != 440 || len(rs) != 78 {
		t.Fatalf("%d/%d", lines, len(rs))
	}
	if rs[0].Symbol != "CYP18A1" || len(rs[0].Sequence) < 150 {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[7].Symbol != "CYP6CM2v2" || !hasStatus(rs[7], "source-missing-region") {
		t.Fatalf("missing=%#v", rs[7])
	}
	if rs[52].Symbol != "seq100" || !hasStatus(rs[52], "source-unnamed") {
		t.Fatalf("unnamed=%#v", rs[52])
	}
	if rs[77].Symbol != "CYP404A1" || !strings.HasSuffix(rs[77].Sequence, "KVQFIPR") {
		t.Fatalf("last=%#v", rs[77])
	}
}
func TestProteinLine(t *testing.T) {
	if got := whiteflyProteinLine("VALENIC VTAMGIKIDALKNPNM"); got != "VALENICVTAMGIKIDALKNPNM" {
		t.Fatal(got)
	}
	if got := whiteflyProteinLine("same as Seq 51"); got != "" {
		t.Fatal(got)
	}
}
func hasStatus(r record, w string) bool {
	for _, s := range r.Status {
		if s == w {
			return true
		}
	}
	return false
}
