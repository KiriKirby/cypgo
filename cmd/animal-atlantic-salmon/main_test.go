package main

import "testing"

func TestParse(t *testing.T) {
	rs, _, e := parse("../../raw/animals-salmo.txt")
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 64 {
		t.Fatal(len(rs))
	}
	if rs[0].Symbol != "CYP1A" || len(rs[0].Sequence) != 522 {
		t.Fatalf("first %s/%d", rs[0].Symbol, len(rs[0].Sequence))
	}
	if rs[59].Symbol != "CYP27B" || rs[59].Sequence != "" {
		t.Fatalf("missing row %#v", rs[59])
	}
}
