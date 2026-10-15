package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"testing"
)

func TestReview(t *testing.T) {
	w, e := plantxlsx.ReadSheet("../../raw/animals-Anolis.carolinensis.xlsx", sourceSheet, false)
	if e != nil {
		t.Fatal(e)
	}
	rs, e := review(w.Rows)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 168 {
		t.Fatal(len(rs))
	}
	if rs[0].Symbol != "CYP1A1" || rs[167].Symbol != "CYP51A1" {
		t.Fatalf("edges %s %s", rs[0].Symbol, rs[167].Symbol)
	}
}
