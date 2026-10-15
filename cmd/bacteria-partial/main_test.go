package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPartialCollectionRecords(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	if err != nil {
		t.Fatal(err)
	}
	text, err := decodeSource(b)
	if err != nil {
		t.Fatal(err)
	}
	records, err := parse(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 30 {
		t.Fatalf("records=%d", len(records))
	}
	for _, r := range records {
		if r.Sequence == "" {
			t.Fatalf("missing sequence block %d", r.Block)
		}
		for _, aa := range r.Sequence {
			if !strings.ContainsRune(alphabet, aa) {
				t.Fatalf("block %d residue %q", r.Block, aa)
			}
		}
	}
	if records[0].Spec.symbol != "CYP127" || len(records[0].Sequence) != 414 {
		t.Fatalf("first=%#v", records[0])
	}
	if records[len(records)-1].Spec.symbol != "CYP102A2" || len(records[len(records)-1].Sequence) != 479 {
		t.Fatalf("last=%#v", records[len(records)-1])
	}
}

func TestIndependentGenesAndExclusions(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	if err != nil {
		t.Fatal(err)
	}
	text, err := decodeSource(b)
	if err != nil {
		t.Fatal(err)
	}
	records, err := parse(text)
	if err != nil {
		t.Fatal(err)
	}
	var aj, af int
	for _, r := range records {
		if strings.HasPrefix(r.Spec.id, "AJ223998") {
			aj++
		}
		if strings.HasPrefix(r.Spec.id, "AF040570") {
			af++
		}
		if strings.Contains(r.Spec.species, "Fusarium") {
			t.Fatal("fungal section entered")
		}
	}
	if aj != 2 || af != 3 {
		t.Fatalf("gene counts AJ223998=%d AF040570=%d", aj, af)
	}
	for _, r := range records {
		if r.Spec.symbol == "CYP118" && (!strings.Contains(r.Sequence, "X") || len(r.Sequence) >= 150) {
			t.Fatalf("CYP118 handling=%#v", r)
		}
	}
}
