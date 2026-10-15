package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rs, lines, e := parse("../../raw/animals-AedesFASTA.June25.txt")
	if e != nil {
		t.Fatal(e)
	}
	if lines != 1522 || len(rs) != 167 {
		t.Fatalf("lines/records=%d/%d", lines, len(rs))
	}
	if rs[0].Symbol != "CYP15B1" || !strings.HasPrefix(rs[0].Sequence, "MWQNLV") || !strings.HasSuffix(rs[0].Sequence, "VQLKERLI") {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[7].Symbol != "CYP305A5" || strings.HasSuffix(rs[7].Sequence, ".") {
		t.Fatalf("dot=%#v", rs[7])
	}
	if rs[153].Symbol != "CYP325Y2v2" || !hasStatus(rs[153], "source-allele") {
		t.Fatalf("allele=%#v", rs[153])
	}
	if rs[166].Symbol != "CYP315A1" || len(rs[166].Sequence) < 300 {
		t.Fatalf("last=%#v", rs[166])
	}
}
func TestAedesProteinLine(t *testing.T) {
	cases := map[string]string{"2678 LLVFCCTLSCVVFLQWLFRLVCQIKKL (2) 2502": "LLVFCCTLSCVVFLQWLFRLVCQIKKL", "MIVLVLTSVLIIAFSYWLLQELRR.": "MIVLVLTSVLIIAFSYWLLQELRR", "Query: 1 MLLVLSVFVVIL": ""}
	for in, want := range cases {
		if got := aedesProteinLine(in); got != want {
			t.Errorf("%q => %q want %q", in, got, want)
		}
	}
}
func hasStatus(r record, want string) bool {
	for _, s := range r.Status {
		if s == want {
			return true
		}
	}
	return false
}
