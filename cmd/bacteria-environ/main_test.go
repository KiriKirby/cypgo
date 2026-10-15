package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironment(t *testing.T) {
	b, e := os.ReadFile(filepath.Join("..", "..", "raw", "bacteria-environ.txt"))
	if e != nil {
		t.Fatal(e)
	}
	rs, e := parse(string(b))
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 168 {
		t.Fatalf("records=%d", len(rs))
	}
	if rs[0].sequence == "" || rs[0].sample != "marine environmental sample" {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[len(rs)-1].sample != "acid mine drainage microbial pink biofilm" {
		t.Fatalf("last=%#v", rs[len(rs)-1])
	}
}
