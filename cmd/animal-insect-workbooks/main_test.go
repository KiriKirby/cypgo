package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"testing"
)

func TestProfiles(t *testing.T) {
	for _, p := range profiles {
		wb, e := plantxlsx.ReadSheet("../../raw/"+p.File, p.Sheet, false)
		if e != nil {
			t.Fatal(e)
		}
		rs, e := review(p, wb.Rows)
		if e != nil {
			t.Fatal(e)
		}
		if len(rs) != p.Expected {
			t.Fatalf("%s=%d", p.Slug, len(rs))
		}
		if rs[0].Sequence == "" || rs[len(rs)-1].Sequence == "" {
			t.Fatal(p.Slug)
		}
	}
}
