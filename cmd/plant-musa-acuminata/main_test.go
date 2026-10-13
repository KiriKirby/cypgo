package main

import (
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"path/filepath"
	"testing"
)

func TestMusa(t *testing.T) {
	w, e := plantxlsx.Read(filepath.Join("..", "..", "raw", config.SourceFile), config.Sheet)
	if e != nil {
		t.Fatal(e)
	}
	a, e := review(w)
	if e != nil {
		t.Fatal(e)
	}
	if len(a) != 233 || a[0].Row != 1 || a[0].ID != "Musaacum813984.1.21241" || len(a[0].Sequence) != 488 || a[232].ID != "Musaacum813981.1.21549" || len(a[232].Sequence) != 518 {
		t.Fatal("Musa representatives changed")
	}
}
