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
	if len(rows) != 7127 {
		t.Fatalf("rows=%d", len(rows))
	}
	checks := []struct {
		index         int
		name, species string
	}{
		{0, "CYP51G1", "Arabidopsis thaliana"},
		{3567, "CYP84A24v1", "Brassica napus"},
		{7126, "CYP805A1", "Thuja plicata"},
	}
	for _, c := range checks {
		r := rows[c.index]
		if r.Name != c.name || r.Species != c.species || !strings.Contains(r.Text, r.Name) {
			t.Fatalf("row %d=%+v", c.index, r)
		}
	}
}

func TestCanonicalSpeciesRejectsMetadata(t *testing.T) {
	for _, value := range []string{"confidential", "discontinued, renamed CYP97A3", "Glyma02g31910.1 pseudogene", "Os06g0570600 BAF19826", "This seq", "Trypanosoma cruzi", "Leishmania major", "Monosiga ovata", "Euglena gracilis"} {
		if got, ok := canonicalSpecies(value); ok || got != "" {
			t.Fatalf("accepted %q as %q", value, got)
		}
	}
}
