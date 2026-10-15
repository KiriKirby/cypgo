package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBacteria2003Source(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	if err != nil {
		t.Fatal(err)
	}
	text, err := decodeSource(b)
	if err != nil {
		t.Fatal(err)
	}
	rs, ex, err := parse(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) != 553 || len(ex) != 2 {
		t.Fatalf("accepted=%d excluded=%d", len(rs), len(ex))
	}
	if rs[0].Block != 1 || rs[0].Symbol != "CYP51" || rs[0].Species != "Mycobacterium tuberculosis" || len(rs[0].Sequence) != 451 {
		t.Fatalf("first=%#v", rs[0])
	}
	mid := rs[len(rs)/2]
	if mid.Block != 277 || mid.Symbol != "CYP142B1" || mid.Species != "Mycobacterium vanbaalenii" || len(mid.Sequence) != 422 {
		t.Fatalf("middle=%#v", mid)
	}
	last := rs[len(rs)-1]
	if last.Block != 553 || last.Symbol != "CYP292A2" || last.Species != "Mycobacterium flavescens" || len(last.Sequence) != 405 {
		t.Fatalf("last=%#v", last)
	}
	if ex[0].Block != 554 || !strings.HasPrefix(ex[0].Header, "L76374") || ex[1].Block != 555 || !strings.HasPrefix(ex[1].Header, "AF071147") {
		t.Fatalf("excluded=%#v", ex)
	}
}

func TestBacteria2003ExceptionalLayouts(t *testing.T) {
	b, _ := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	text, _ := decodeSource(b)
	rs, _, err := parse(text)
	if err != nil {
		t.Fatal(err)
	}
	byBlock := map[int]record{}
	for _, r := range rs {
		byBlock[r.Block] = r
	}
	for _, n := range []int{27, 233, 323} {
		if !hasStatus(byBlock[n], "source-inline-protein") {
			t.Fatalf("block %d missing inline status", n)
		}
	}
	if !strings.HasPrefix(byBlock[27].Sequence, inlineHeaderProtein[27]) || byBlock[233].Sequence != inlineHeaderProtein[233] {
		t.Fatal("inline protein was not retained literally")
	}
	if !strings.HasPrefix(byBlock[265].Sequence, "mrtyrtv") || !hasStatus(byBlock[265], "source-lowercase-residue") {
		t.Fatalf("lowercase block=%#v", byBlock[265])
	}
	for _, n := range []int{25, 174, 193, 265, 325} {
		if strings.Contains(byBlock[n].Sequence, `"`) {
			t.Fatalf("block %d retained Word quote", n)
		}
	}
	if len(byBlock[78].Sequence) != 89 || !hasStatus(byBlock[78], "source-pseudogene") || !hasStatus(byBlock[78], "short-sequence") {
		t.Fatalf("short pseudogene=%#v", byBlock[78])
	}
}

func TestProteinLineReviewedForms(t *testing.T) {
	cases := map[string]string{
		"3021360 TSTVVPRVSGGEEEHGHLEEFRTDPIGLMQRVRDECGDVGWFQLVDKHVILLSGAQANEF 3021539": "TSTVVPRVSGGEEEHGHLEEFRTDPIGLMQRVRDECGDVGWFQLVDKHVILLSGAQANEF",
		"1 MDKKVSAIPQ PKTYGPLGNL PLIDKDKPTL":                                           "MDKKVSAIPQPKTYGPLGNLPLIDKDKPTL",
		`TGDWPRMIGHGGARRATGMTVNLGVDR"`:                                                 "TGDWPRMIGHGGARRATGMTVNLGVDR",
		"mrtyrtvRYPLGEALLALYRWRGPLINAGVGGHGY":                                          "mrtyrtvRYPLGEALLALYRWRGPLINAGVGGHGY",
	}
	for raw, want := range cases {
		got, ok := proteinLine(raw)
		if !ok || got != want {
			t.Fatalf("%q => %q,%v want %q", raw, got, ok, want)
		}
	}
	if _, ok := proteinLine("this is annotation prose"); ok {
		t.Fatal("prose accepted")
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
