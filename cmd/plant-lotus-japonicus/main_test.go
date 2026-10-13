package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewRowsUsesExactLotusLayout(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 247; row++ {
		id := "Lotujapo.test"
		rows[row] = map[string]string{"A": id, "H": id, "I": "CYP71D1", "J": "90", "K": "CYP71D", "L": strings.Repeat("A", 500)}
	}
	for row := 248; row <= 286; row++ {
		rows[row] = map[string]string{"A": "unassigned", "H": "unassigned", "L": strings.Repeat("O", 200)}
	}
	rows[2] = map[string]string{"A": "LotujapoCM0846.228", "H": "LotujapoCM0846.228", "I": "CYP51G1", "J": "90.06", "K": "CYP51G1", "L": strings.Repeat("A", 491)}
	rows[34]["I"], rows[34]["L"] = "CYP71D84P", strings.Repeat("A", 100)+"-"+strings.Repeat("X", 33)+"O"+strings.Repeat("A", 200)
	rows[197]["I"], rows[197]["K"] = "CYP706A9P", "CYP706A9P"
	rows[124] = map[string]string{"A": "LotujapoCM1089.101", "H": "LotujapoCM1089.101", "I": "CYP83E2", "J": "100", "K": "CYP83E2", "L": strings.Repeat("A", 510)}
	rows[247] = map[string]string{"A": "LotujapoCM0241.538", "H": "LotujapoCM0241.538", "I": "CYP736A80", "J": "95.06", "K": "CYP736A", "L": strings.Repeat("A", 497)}
	wb := &plantxlsx.Workbook{Dimension: "A1:L286", Rows: rows}
	got, err := reviewRows(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 246 || got[0].Row != 2 || got[len(got)-1].Row != 247 {
		t.Fatalf("boundary=%d first=%d last=%d", len(got), got[0].Row, got[len(got)-1].Row)
	}
	if got[0].SeqID != "LotujapoCM0846.228" || got[0].Symbol != "CYP51G1" || len(got[0].Sequence) != 491 {
		t.Fatalf("first=%#v", got[0])
	}
	if got[32].Status != "source-best-hit-pseudogene-label;source-gap;ambiguous-X;nonstandard-O;short-sequence" {
		t.Fatalf("exception=%#v", got[32])
	}
	if got[195].Status != "source-best-hit-pseudogene-label;source-pseudogene-label" {
		t.Fatalf("assigned pseudogene=%#v", got[195])
	}
}

func TestReviewRowsRejectsAssignmentAfterBoundary(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 247; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 248; row <= 286; row++ {
		rows[row] = map[string]string{"H": "candidate", "L": strings.Repeat("O", 200)}
	}
	rows[248]["K"] = "CYP71A"
	wb := &plantxlsx.Workbook{Dimension: "A1:L286", Rows: rows}
	if _, err := reviewRows(wb); err == nil {
		t.Fatal("expected row-boundary error")
	}
}
