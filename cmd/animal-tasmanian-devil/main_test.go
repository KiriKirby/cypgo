package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"testing"
)

func TestReview(t *testing.T) {
	wb, e := load()
	if e != nil {
		t.Fatal(e)
	}
	rs, e := review(wb.Rows)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 91 {
		t.Fatal(len(rs))
	}
	for _, i := range []int{0, 45, 90} {
		if rs[i].Symbol == "" || rs[i].Sequence == "" {
			t.Fatalf("row %d empty", i)
		}
	}
	if rs[0].Symbol != "CYP1A1" || len(rs[0].Sequence) != 516 {
		t.Fatalf("first=%s/%d", rs[0].Symbol, len(rs[0].Sequence))
	}
}
func load() (*plantxlsx.Workbook, error) {
	return plantxlsx.ReadSheet("../../raw/animals-Tasmanian.devil.xlsx", sourceSheet, false)
}
