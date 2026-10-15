package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse2011Source(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	if err != nil {
		t.Fatal(err)
	}
	records, err := parse(string(b))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1152 {
		t.Fatalf("records=%d", len(records))
	}
	for _, index := range []int{0, 575, 1151} {
		r := records[index]
		if r.Symbol == "" || r.Species == "" || r.Sequence == "" {
			t.Fatalf("representative block %d incomplete: %#v", index+1, r)
		}
	}
	if records[0].Symbol != "CYP51B1" || records[0].Species != "Mycobacterium tuberculosis" {
		t.Fatalf("first=%#v", records[0])
	}
	last := records[len(records)-1]
	if last.Symbol != "CYP1047A1" || last.Species != "Streptomyces griseus" {
		t.Fatalf("last=%#v", last)
	}
}

func TestReviewedExceptionalSpecies(t *testing.T) {
	tests := map[string]string{
		"uncultured soil bacterium":        "uncultured soil bacterium",
		"JCVI_PEP_1 seawater sequencing":   "seawater environmental sample",
		"sce5725 49% to CYP261A1":          "Sorangium cellulosum",
		"SP0881 91% to 107N1":              "Streptomyces peucetius",
		"uncultured haloarchaeon FLAS10H9": "uncultured haloarchaeon FLAS10H9",
	}
	for input, want := range tests {
		if got := speciesFromHeader(input); got != want {
			t.Errorf("%q => %q want %q", input, got, want)
		}
	}
}

func TestProteinLineRejectsAnnotation(t *testing.T) {
	if got := proteinLine("63362  MSVDRTTCPGQPFAQPGAGVLAHSRALRFWLDPAN  63541"); got != "MSVDRTTCPGQPFAQPGAGVLAHSRALRFWLDPAN" {
		t.Fatalf("protein=%q", got)
	}
	if got := proteinLine("89% to CYP1046A1"); got != "" {
		t.Fatalf("annotation accepted: %q", got)
	}
	if got := proteinLine("AP009493.1"); got != "" {
		t.Fatalf("accession accepted: %q", got)
	}
}
