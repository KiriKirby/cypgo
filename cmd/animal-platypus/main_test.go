package main

import "testing"

func TestParse(t *testing.T) {
	rs, _, e := parse("../../raw/animals-duckbill.txt")
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 55 {
		t.Fatal(len(rs))
	}
	for _, i := range []int{0, 27, 54} {
		if rs[i].Sequence == "" {
			t.Fatal(i)
		}
	}
	if rs[0].Symbol != "CYP1A1" || len(rs[0].Sequence) != 519 {
		t.Fatalf("first %s/%d", rs[0].Symbol, len(rs[0].Sequence))
	}
	if rs[1].Symbol != "CYP1B1" || len(rs[1].Sequence) != 350 {
		t.Fatalf("split block not joined: %d", len(rs[1].Sequence))
	}
}
func TestUpperResidues(t *testing.T) {
	if !isUpperResidues("MAAAX*") {
		t.Fatal()
	}
	if isUpperResidues("Cyandoesnotmatch") {
		t.Fatal()
	}
}
