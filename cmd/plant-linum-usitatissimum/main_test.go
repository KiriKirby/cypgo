package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesExactLinumBoundaries(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 469; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 470; row <= 479; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("O", 200)}
	}
	rows[2] = map[string]string{"A": "Linuusit731.8", "H": "Linuusit731.8", "I": "CYP51G1", "J": "85.59", "K": "CYP51G", "L": strings.Repeat("A", 489)}
	rows[20]["I"], rows[20]["L"] = "CYP71D163P", strings.Repeat("A", 100)+"-"+strings.Repeat("X", 20)+"O"+strings.Repeat("A", 150)
	rows[236] = map[string]string{"A": "Linuusit231.525", "H": "Linuusit231.525", "I": "CYP85A31", "J": "80", "K": "CYP85A", "L": strings.Repeat("A", 459)}
	rows[469] = map[string]string{"A": "Linuusit641.547", "H": "Linuusit641.547", "I": "CYP89A65P", "J": "33.33", "K": "CYP", "L": strings.Repeat("A", 269)}
	wb := &plantxlsx.Workbook{Dimension: "A1:L479", Rows: rows}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 468 || len(excluded) != 10 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 469 {
		t.Fatalf("accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	if accepted[18].Status != "source-best-hit-pseudogene-label;source-gap;ambiguous-X-or-x;nonstandard-O;short-sequence" {
		t.Fatalf("exception=%#v", accepted[18])
	}
	if accepted[len(accepted)-1].ID != "Linuusit641.547" || accepted[len(accepted)-1].Symbol != "CYP" {
		t.Fatalf("last=%#v", accepted[len(accepted)-1])
	}
}

func TestReviewRejectsLinumIDMismatch(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 469; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 470; row <= 479; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("A", 200)}
	}
	rows[2]["H"] = "different"
	wb := &plantxlsx.Workbook{Dimension: "A1:L479", Rows: rows}
	if _, _, err := review(wb); err == nil {
		t.Fatal("expected A/H ID mismatch")
	}
}
