package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestExactSetariaWorkbook(t *testing.T) {
	w, e := plantxlsx.Read(filepath.Join("..", "..", "raw", config.SourceFile), config.Sheet)
	if e != nil {
		t.Fatal(e)
	}
	a, x, e := review(w)
	if e != nil {
		t.Fatal(e)
	}
	if len(a) != 413 || len(x) != 5 {
		t.Fatalf("a=%d x=%d", len(a), len(x))
	}
	for _, c := range []struct {
		i, row, l  int
		id, hit, s string
	}{{0, 2, 488, "Setaital8.26299", "CYP51G1", "CYP51G1"}, {206, 208, 523, "Setaital9.4984", "CYP81A27P", "CYP81A"}, {412, 414, 530, "Setaital6.27518", "CYP735A30", "CYP735A"}} {
		r := a[c.i]
		if r.Row != c.row || r.ID != c.id || r.BestHit != c.hit || r.Symbol != c.s || len(r.Sequence) != c.l {
			t.Errorf("case=%+v got=%+v", c, r)
		}
	}
	if x[0].Row != 415 || x[4].Row != 419 {
		t.Fatal("excluded boundary")
	}
}
