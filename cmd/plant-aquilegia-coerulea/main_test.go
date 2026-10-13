package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestReviewedAquilegiaWorkbook(t *testing.T) {
	w, e := plantxlsx.Read(filepath.Join("..", "..", "raw", sourceFile), sheetName)
	if e != nil {
		t.Fatal(e)
	}
	r, e := reviewRows(w)
	if e != nil {
		t.Fatal(e)
	}
	if len(r) != 551 {
		t.Fatal(len(r))
	}
	for _, c := range []struct {
		i, row, l int
		id, s     string
	}{{0, 2, 496, "Aquicoer8.821", "CYP51G1"}, {275, 277, 512, "Aquicoer7.2951", "CYP81C"}, {550, 552, 472, "Aquicoer116.257", "CYP865A"}} {
		x := r[c.i]
		if x.Row != c.row || x.ID != c.id || x.Symbol != c.s || len(x.Sequence) != c.l {
			t.Errorf("%d: %#v len=%d", c.i, x, len(x.Sequence))
		}
	}
}
