package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rs, l, e := parse("../../raw/animals-Monarch.butterfly.txt")
	if e != nil {
		t.Fatal(e)
	}
	if l != 1288 || len(rs) != 86 {
		t.Fatalf("%d/%d", l, len(rs))
	}
	if rs[0].ID != "DPGLEAN22546" || !strings.HasPrefix(rs[0].Sequence, "MWPLI") {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[54].ID != "DPGLEAN18469" || !has(rs[54], "source-joined-model") {
		t.Fatalf("join=%#v", rs[54])
	}
	if rs[81].ID != "DPGLEAN14757" || strings.Contains(rs[81].Sequence, "GQTSRRP") {
		t.Fatalf("terminal=%#v", rs[81])
	}
	if rs[85].ID != "DPGLEAN11745" {
		t.Fatalf("last=%#v", rs[85])
	}
}
func has(r record, w string) bool {
	for _, s := range r.Status {
		if s == w {
			return true
		}
	}
	return false
}
