package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestReviewExactOryzaWorkbook(t *testing.T) {
	wb, err := plantxlsx.Read(filepath.Join("..", "..", "raw", config.SourceFile), config.Sheet)
	if err != nil {
		t.Fatal(err)
	}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 316 || len(excluded) != 4 {
		t.Fatalf("accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	for _, c := range []struct {
		i, row, length  int
		id, hit, symbol string
	}{
		{0, 2, 489, "Oryzbrac006267462.1.2", "CYP51G1", "CYP51G1"},
		{158, 160, 401, "Oryzbrac006267384.1.12964", "CYP81P1", "CYP81P"},
		{315, 317, 477, "Oryzbrac006267381.1.7194", "CYP735A4", "CYP735A"},
	} {
		r := accepted[c.i]
		if r.Row != c.row || r.ID != c.id || r.BestHit != c.hit || r.Symbol != c.symbol || len(r.Sequence) != c.length {
			t.Errorf("case=%+v got row=%d id=%s hit=%s symbol=%s len=%d", c, r.Row, r.ID, r.BestHit, r.Symbol, len(r.Sequence))
		}
	}
	if excluded[0].Row != 318 || excluded[3].Row != 321 {
		t.Fatalf("excluded boundaries=%d..%d", excluded[0].Row, excluded[3].Row)
	}
}
