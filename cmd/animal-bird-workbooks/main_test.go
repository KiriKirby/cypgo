package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestProfiles(t *testing.T) {
	for _, p := range profiles {
		w, e := plantxlsx.ReadSheet(filepath.Join("..", "..", "raw", p.File), p.Sheet, false)
		if e != nil {
			t.Fatal(p.Slug, e)
		}
		rs, e := review(p, w.Rows)
		if e != nil {
			t.Fatal(p.Slug, e)
		}
		if len(rs) != p.Expected {
			t.Fatalf("%s=%d", p.Slug, len(rs))
		}
		if rs[0].Sequence == "" || rs[len(rs)-1].Sequence == "" {
			t.Fatal(p.Slug, "empty edge")
		}
	}
}
