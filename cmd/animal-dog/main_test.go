package main

import "testing"

func TestParseDog(t *testing.T) {
	rs, _, err := parseDog("../../raw/animals-dog.dbs.20110817.html")
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 56 {
		t.Fatalf("records=%d", len(rs))
	}
	checks := map[int]struct {
		symbol string
		length int
	}{0: {"CYP1A1", 521}, 27: {"CYP4A36P", 510}, 55: {"CYP51A1", 503}}
	for i, want := range checks {
		if rs[i].Symbol != want.symbol || len(rs[i].Sequence) != want.length {
			t.Errorf("row %d got %s/%d want %s/%d", i+1, rs[i].Symbol, len(rs[i].Sequence), want.symbol, want.length)
		}
	}
}

func TestDogExceptionalLiteralContent(t *testing.T) {
	rs, _, err := parseDog("../../raw/animals-dog.dbs.20110817.html")
	if err != nil {
		t.Fatal(err)
	}
	if rs[3].Symbol != "CYP1D1P" || !contains(rs[3].Sequence, "*") || !contains(rs[3].Sequence, "x") {
		t.Fatalf("CYP1D1P literal exceptions lost: %q", rs[3].Sequence)
	}
	if rs[10].Symbol != "CYP2C21-ie5b" || len(rs[10].Sequence) >= 100 {
		t.Fatalf("internal-exon fragment not retained: %#v", rs[10])
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
