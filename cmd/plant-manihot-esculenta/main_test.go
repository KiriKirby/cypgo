package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesExactManihotBoundariesAndIDs(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 337; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 338; row <= 349; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("O", 200)}
	}
	rows[2] = map[string]string{"A": "Maniescu00510.82", "H": "Maniescu00510.82", "I": "CYP51G5", "J": "90.1", "K": "CYP51G", "L": strings.Repeat("A", 486)}
	rows[20]["I"], rows[20]["L"] = "CYP71D163P", strings.Repeat("A", 100)+"-"+strings.Repeat("X", 20)+"O"+strings.Repeat("A", 150)
	rows[170] = map[string]string{"A": "Maniescu06512.602", "H": "Maniescu06512.602", "I": "CYP84A10", "J": "80", "K": "CYP84A", "L": strings.Repeat("A", 514)}
	rows[337] = map[string]string{"A": "Maniescu05162.61", "H": "Maniescu05162.61", "I": "CYP82C31", "J": "35.51", "K": "CYP82C31", "L": strings.Repeat("A", 427)}
	wb := &plantxlsx.Workbook{Dimension: "A1:L349", Rows: rows}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 336 || len(excluded) != 12 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 337 {
		t.Fatalf("accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	if accepted[0].ID != "Maniescu00510.82" || accepted[0].Symbol != "CYP51G" || len(accepted[0].Sequence) != 486 {
		t.Fatalf("first=%#v", accepted[0])
	}
	if accepted[18].Status != "source-best-hit-pseudogene-label;source-gap;ambiguous-X-or-x;nonstandard-O;short-sequence" {
		t.Fatalf("exception=%#v", accepted[18])
	}
	if accepted[len(accepted)-1].ID != "Maniescu05162.61" || accepted[len(accepted)-1].Symbol != "CYP82C31" {
		t.Fatalf("last=%#v", accepted[len(accepted)-1])
	}
}

func TestReviewRejectsManihotIDMismatch(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 337; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 338; row <= 349; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("A", 200)}
	}
	rows[2]["H"] = "different"
	wb := &plantxlsx.Workbook{Dimension: "A1:L349", Rows: rows}
	if _, _, err := review(wb); err == nil {
		t.Fatal("expected A/H ID mismatch")
	}
}
