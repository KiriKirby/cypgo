package main

import (
	"path/filepath"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestAmborellaWorkbook(t *testing.T) {
	w, err := plantxlsx.Read(filepath.Join("..", "..", "raw", config.SourceFile), config.Sheet)
	if err != nil {
		t.Fatal(err)
	}
	a, err := review(w)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 230 || a[0].ID != "Ambotric00016.705" || a[0].Symbol != "CYP51G24P" || len(a[0].Sequence) != 339 {
		t.Fatal("first Amborella record changed")
	}
	if a[114].ID != "Ambotric00059.3827" || a[229].ID != "Ambotric00047.2827" || a[229].Symbol != "CYP865C1" || len(a[229].Sequence) != 510 {
		t.Fatal("middle or last Amborella record changed")
	}
}
