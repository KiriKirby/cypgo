package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestExactPanicumWorkbook(t *testing.T) {
	p := filepath.Join("..", "..", "raw", config.SourceFile)
	n, e := plantxlsx.ReadSheet(p, config.Sheet, false)
	if e != nil {
		t.Fatal(e)
	}
	l, e := plantxlsx.ReadSheet(p, "Sorted by length", false)
	if e != nil {
		t.Fatal(e)
	}
	a, e := review(n, l)
	if e != nil {
		t.Fatal(e)
	}
	if len(a) != 806 {
		t.Fatal(len(a))
	}
	for _, c := range []struct {
		i, row, z  int
		id, hit, s string
	}{{0, 2, 289, "Panivirg24881.7", "CYP51G1", "CYP51G"}, {402, 404, 497, "Panivirg17853.7", "CYP87A15", "CYP87A"}, {805, 807, 215, "Panivirg146243.1", "CYP76F35", "CYP"}} {
		r := a[c.i]
		if r.Row != c.row || r.ID != c.id || r.BestHit != c.hit || r.Symbol != c.s || len(r.Sequence) != c.z {
			t.Errorf("case=%+v got=%+v", c, r)
		}
	}
}
