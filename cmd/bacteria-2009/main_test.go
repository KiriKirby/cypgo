package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBacteria2009(t *testing.T) {
	b, e := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	if e != nil {
		t.Fatal(e)
	}
	rs, e := parse(string(b))
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 967 {
		t.Fatal(len(rs))
	}
	if rs[0].Species != "Mycobacterium tuberculosis" || len(rs[0].Sequence) != 450 {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[3].Sequence != "" || rs[3].Symbol != "CYP51B1" {
		t.Fatalf("missing=%#v", rs[3])
	}
	if rs[len(rs)-1].Symbol != "CYP1014A1" || len(rs[len(rs)-1].Sequence) != 432 {
		t.Fatalf("last=%#v", rs[len(rs)-1])
	}
}
