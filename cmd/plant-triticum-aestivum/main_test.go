package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestExactTriticumWorkbook(t *testing.T) {
	w, e := plantxlsx.Read(filepath.Join("..", "..", "raw", sourceFile), "Sorted by CYP name")
	if e != nil {
		t.Fatal(e)
	}
	a, x, e := review(w)
	if e != nil {
		t.Fatal(e)
	}
	if len(a) != 1475 || len(x) != 225 {
		t.Fatalf("%d/%d", len(a), len(x))
	}
	for _, c := range []struct {
		i, row, l int
		id, hit   string
	}{{0, 2, 488, "Tritaesb4S4940376.9", "CYP51G1"}, {737, 739, 496, "Tritaesa1S3314486.3", "CYP76V1"}, {1474, 1476, 465, "Tritaesd5L4564033.2", "CYP735A4"}} {
		r := a[c.i]
		if r.Row != c.row || r.ID != c.id || r.Hit != c.hit || len(r.Sequence) != c.l {
			t.Errorf("case=%+v got=%+v", c, r)
		}
	}
	if x[0].Row != 1477 || x[len(x)-1].Row != 1701 || x[len(x)-1].Status != "excluded: explicit bacterial contamination" {
		t.Fatal("exclusions")
	}
}
