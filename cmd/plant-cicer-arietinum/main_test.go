package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesExactCicerBoundariesAndIDs(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 212; row++ {
		id := "Cicearie.test"
		rows[row] = map[string]string{"A": id, "H": id, "I": "CYP71D1", "J": "90", "K": "CYP71D", "L": strings.Repeat("A", 500)}
	}
	for row := 213; row <= 231; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("O", 200)}
	}
	rows[2] = map[string]string{"A": "Cicearie00316.166", "H": "Cicearie00316.166", "I": "CYP51G1", "J": "94.48", "K": "CYP51G1", "L": strings.Repeat("A", 489)}
	rows[8]["I"], rows[8]["L"] = "CYP71D71P", strings.Repeat("A", 100)+"-"+strings.Repeat("X", 20)+"O"+strings.Repeat("A", 150)
	rows[107] = map[string]string{"A": "Cicearie05416.5", "H": "Cicearie05416.5", "I": "CYP85A36", "J": "98", "K": "CYP85A", "L": strings.Repeat("A", 464)}
	rows[212] = map[string]string{"A": "Cicearie00884.10", "H": "Cicearie00884.10", "I": "CYP714A17", "J": "29.8", "K": "CYP", "L": strings.Repeat("A", 417)}
	wb := &plantxlsx.Workbook{Dimension: "A1:L231", Rows: rows}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 211 || len(excluded) != 19 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 212 {
		t.Fatalf("accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	if accepted[0].ID != "Cicearie00316.166" || accepted[0].Symbol != "CYP51G1" || len(accepted[0].Sequence) != 489 {
		t.Fatalf("first=%#v", accepted[0])
	}
	if accepted[6].Status != "source-best-hit-pseudogene-label;source-gap;ambiguous-X-or-x;nonstandard-O;short-sequence" {
		t.Fatalf("exception=%#v", accepted[6])
	}
	if accepted[len(accepted)-1].ID != "Cicearie00884.10" || accepted[len(accepted)-1].Symbol != "CYP" {
		t.Fatalf("last=%#v", accepted[len(accepted)-1])
	}
}

func TestReviewRejectsCicerIDMismatch(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 212; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 213; row <= 231; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("A", 200)}
	}
	rows[2]["H"] = "different"
	wb := &plantxlsx.Workbook{Dimension: "A1:L231", Rows: rows}
	if _, _, err := review(wb); err == nil {
		t.Fatal("expected A/H ID mismatch")
	}
}
