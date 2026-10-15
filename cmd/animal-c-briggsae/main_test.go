package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	rs, lines, err := parse("../../raw/animals-C.briggsae.txt")
	if err != nil {
		t.Fatal(err)
	}
	if lines != 307 || len(rs) != 23 {
		t.Fatalf("lines/records=%d/%d", lines, len(rs))
	}
	if rs[0].Symbol != "CYP13A.a" || len(rs[0].Sequence) != 515 || !strings.HasSuffix(rs[0].Sequence, "KLEPRKDHL") {
		t.Fatalf("first=%s/%d/%s", rs[0].Symbol, len(rs[0].Sequence), rs[0].Sequence)
	}
	if rs[9].Symbol != "CYP14A.h" || len(rs[9].Sequence) != 41 || !hasStatus(rs[9], "source-pseudogene") {
		t.Fatalf("solo exon=%#v", rs[9])
	}
	if rs[11].Symbol != "CYP14A.c" || !strings.Contains(rs[11].Sequence, "WRL*RRT") || !hasStatus(rs[11], "internal-stop-retained") {
		t.Fatalf("internal stop=%#v", rs[11])
	}
	if rs[13].Symbol != "CYP14A.e" || !strings.Contains(rs[13].Sequence, "GNPGXXXXX") || !hasStatus(rs[13], "ambiguous-X-retained") {
		t.Fatalf("X pseudogene=%#v", rs[13])
	}
	if rs[18].Symbol != "CYP33C.c" || strings.Contains(rs[18].Sequence, "Frameshift") || !hasStatus(rs[18], "source-possible-pseudogene") {
		t.Fatalf("frameshift=%#v", rs[18])
	}
	if rs[22].Symbol != "CYP35A.c" || !strings.HasSuffix(rs[22].Sequence, "KLEMQFVKIRS") {
		t.Fatalf("last=%s/%d/%s", rs[22].Symbol, len(rs[22].Sequence), rs[22].Sequence)
	}
}

func TestProteinOnCoordinateLine(t *testing.T) {
	cases := map[string]string{
		"1378847 MFLLFSKL (0)1378650":              "MFLLFSKL",
		"        APFK LYLMIKKAVMERKEQRDRE 1377317": "APFKLYLMIKKAVMERKEQRDRE",
		"723082 GTKPYII (Frameshift) 722080":       "GTKPYII",
		"954733 (0) YHKFIEDEVASQIKEY (2) 954858":   "YHKFIEDEVASQIKEY",
	}
	for input, want := range cases {
		if got := proteinOnCoordinateLine(input); got != want {
			t.Errorf("%q => %q want %q", input, got, want)
		}
	}
}

func hasStatus(r record, want string) bool {
	for _, status := range r.Status {
		if status == want {
			return true
		}
	}
	return false
}
