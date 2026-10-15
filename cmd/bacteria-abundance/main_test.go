package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAbundanceIsIndexOnly(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "raw", textFile))
	if err != nil {
		t.Fatal(err)
	}
	r, err := inspect(string(b))
	if err != nil {
		t.Fatal(err)
	}
	if r.CYPIndexLines != 490 || r.ProteinLines != 0 || r.FASTAHeaders != 0 || r.AlignmentLines != 0 {
		t.Fatalf("review=%+v", r)
	}
}

func TestSequenceLookingMetadataRejected(t *testing.T) {
	for _, line := range []string{"CYP110E4 complement(3256348..3257724)", "NC_005125, complete genome", "Class Chroobacteria (9)"} {
		if proteinOnly.MatchString(line) {
			t.Fatalf("metadata accepted as protein: %q", line)
		}
	}
}
