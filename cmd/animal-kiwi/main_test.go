package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"testing"
)

func TestReview(t *testing.T) {
	w, e := plantxlsx.ReadSheet("../../raw/animals-Kiwi.xlsx", sourceSheet, false)
	if e != nil {
		t.Fatal(e)
	}
	rs, e := review(w.Rows)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 65 {
		t.Fatal(len(rs))
	}
	if rs[0].Symbol != "CYP1A4/1A1" || len(rs[0].Sequence) != 530 {
		t.Fatalf("first %s/%d", rs[0].Symbol, len(rs[0].Sequence))
	}
	if rs[64].Row != 66 {
		t.Fatal(rs[64].Row)
	}
}
