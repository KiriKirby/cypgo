package main

import (
	"os"
	"strings"
	"testing"
)

func TestParseBiblioD(t *testing.T) {
	b, err := os.ReadFile("../../raw/plants-biblioD.html")
	if err != nil {
		t.Skipf("raw source unavailable: %v", err)
	}
	rows := parse(b)
	if len(rows) != 7134 {
		t.Fatalf("rows=%d", len(rows))
	}
	checks := []struct {
		index         int
		name, species string
	}{
		{0, "CYP51G1", "Arabidopsis thaliana"},
		{3567, "CYP84A24v1", "Brassica napus"},
		{7133, "CYP805A1", "Thuja plicata"},
	}
	for _, c := range checks {
		r := rows[c.index]
		if r.Name != c.name || r.Species != c.species || !strings.Contains(r.Text, r.Name) {
			t.Fatalf("row %d=%+v", c.index, r)
		}
	}
}

func TestClassifyOrganismRejectsMetadata(t *testing.T) {
	for _, value := range []string{"confidential", "discontinued, renamed CYP97A3", "Glyma02g31910.1 pseudogene", "Os06g0570600 BAF19826", "This seq"} {
		if category, got, ok := classifyOrganism(value); ok || category != "" || got != "" {
			t.Fatalf("accepted %q as %s/%q", value, category, got)
		}
	}
}

func TestClassifyNonPlantAndExtractProtein(t *testing.T) {
	category, species, ok := classifyOrganism("Trypanosoma cruzi")
	if !ok || category != "protists" || species != "Trypanosoma cruzi" {
		t.Fatalf("classification=%s/%q/%v", category, species, ok)
	}
	b, err := os.ReadFile("../../raw/plants-biblioD.html")
	if err != nil {
		t.Skipf("raw source unavailable: %v", err)
	}
	for _, row := range parse(b) {
		if row.Name != "CYP84A34" {
			continue
		}
		if len(row.Sequence) < 500 || strings.HasSuffix(row.Sequence, "*") {
			t.Fatalf("CYP84A34 sequence length=%d sequence=%q", len(row.Sequence), row.Sequence)
		}
		return
	}
	t.Fatal("CYP84A34 not found")
}
