package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rs, lines, err := parse("../../raw/animals-Planarian.txt")
	if err != nil {
		t.Fatal(err)
	}
	if lines != 818 || len(rs) != 69 {
		t.Fatalf("lines/records=%d/%d", lines, len(rs))
	}
	if rs[0].ID != "v31.024852" || !strings.HasPrefix(rs[0].Sequence, "MIYIILILLL") || strings.HasSuffix(rs[0].Sequence, "*") {
		t.Fatalf("first=%#v", rs[0])
	}
	if rs[4].ID != "v31.000094" || !strings.Contains(rs[4].Sequence, "NSTI*ENPNEFNP*R") || !hasStatus(rs[4], "source-pseudogene") {
		t.Fatalf("pseudogene=%#v", rs[4])
	}
	if rs[23].ID != "v31.003664b" || !strings.Contains(rs[23].Sequence, "XXXX") || !strings.Contains(rs[23].Sequence, "-") {
		t.Fatalf("gap=%#v", rs[23])
	}
	if rs[65].Species != "Dugesia ryukyuensis" || rs[65].ID != "BW640569.1" {
		t.Fatalf("foreign=%#v", rs[65])
	}
	if rs[66].ID != "CYP20" || strings.Contains(rs[66].Sequence, "aga") || !strings.HasPrefix(rs[66].Sequence, "KPKQIV") {
		t.Fatalf("CYP20=%#v", rs[66])
	}
	if rs[68].ID != "v31.000354" || !strings.HasSuffix(rs[68].Sequence, "KFQSRI") {
		t.Fatalf("last=%#v", rs[68])
	}
}

func TestPlanarianProteinLine(t *testing.T) {
	cases := map[string]string{
		"23008(2)aga KPKQIVIIPGLQKSDPK(2)": "KPKQIVIIPGLQKSDPK",
		"      xxxxxxxxxxxxxxxxxxxxxxxxx":  "xxxxxxxxxxxxxxxxxxxxxxxxx",
		"EKFE &":                           "EKFE",
		"(small deletion)":                 "",
	}
	for input, want := range cases {
		if got := planarianProteinLine(input); got != want {
			t.Errorf("%q => %q want %q", input, got, want)
		}
	}
}

func hasStatus(r record, want string) bool {
	for _, s := range r.Status {
		if s == want {
			return true
		}
	}
	return false
}
