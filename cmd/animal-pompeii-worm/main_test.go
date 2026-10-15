package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"testing"
)

func TestReview(t *testing.T) {
	w, e := plantxlsx.ReadSheet("../../raw/animals-POMPEII.worm.ESTs.xlsx", sheet, false)
	if e != nil {
		t.Fatal(e)
	}
	rs, e := review(w.Rows)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 186 {
		t.Fatal(len(rs))
	}
	if rs[0].Sequence == "" || rs[185].Sequence == "" {
		t.Fatal("edge empty")
	}
	n := 0
	for _, r := range rs {
		if r.GI == "" {
			n++
		}
	}
	if n != 26 {
		t.Fatalf("consensus=%d", n)
	}
}
