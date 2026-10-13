package main

import (
	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestPhoenix(t *testing.T) {
	w, e := plantxlsx.Read(filepath.Join("..", "..", "raw", config.SourceFile), config.Sheet)
	if e != nil {
		t.Fatal(e)
	}
	a, x, e := planttable.Review(config, w)
	if e != nil {
		t.Fatal(e)
	}
	if len(a) != 209 || len(x) != 22 || a[0].ID != "Phoedact849331.3" || len(a[0].Sequence) != 489 || x[0].Row != 211 || x[21].Row != 232 {
		t.Fatal("Phoenix layout changed")
	}
}
