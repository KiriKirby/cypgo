package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rs, l, e := parse("../../raw/animals-nematostella.txt")
	if e != nil {
		t.Fatal(e)
	}
	if l != 878 || len(rs) != 91 {
		t.Fatalf("%d/%d", l, len(rs))
	}
	if rs[0].ID != "XM_001641370" || !strings.HasPrefix(rs[0].Sequence, "MDRLRD") {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[45].ID != "XM_001630310" || rs[90].ID != "XM_001617356" || rs[90].Sequence == "" {
		t.Fatalf("middle/last=%#v/%#v", rs[45], rs[90])
	}
}
