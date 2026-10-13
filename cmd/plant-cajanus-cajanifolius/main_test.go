package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesExactCajanusBoundariesAndIDs(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 292; row++ {
		id := "Cajacaja.test"
		rows[row] = map[string]string{"A": id, "H": id, "I": "CYP71D1", "J": "90", "K": "CYP71D", "L": strings.Repeat("A", 500)}
	}
	for row := 293; row <= 309; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("O", 200)}
	}
	rows[2] = map[string]string{"A": "Cajacaja03.1045", "H": "Cajacaja03.1045", "I": "CYP51G1", "J": "94", "K": "CYP51G", "L": strings.Repeat("A", 487)}
	rows[36]["I"], rows[36]["L"] = "CYP71D163P", strings.Repeat("A", 100)+"-"+strings.Repeat("X", 20)+"O"+strings.Repeat("A", 150)
	rows[147] = map[string]string{"A": "Cajacaja126288.1", "H": "Cajacaja126288.1", "I": "CYP82L5", "J": "80", "K": "CYP82L", "L": strings.Repeat("A", 288)}
	rows[292] = map[string]string{"A": "Cajacaja000201.48", "H": "Cajacaja000201.48", "I": "CYP88D7", "J": "30", "K": "CYP", "L": strings.Repeat("A", 480)}
	wb := &plantxlsx.Workbook{Dimension: "A1:L309", Rows: rows}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 291 || len(excluded) != 17 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 292 {
		t.Fatalf("accepted=%d excluded=%d", len(accepted), len(excluded))
	}
	if accepted[0].ID != "Cajacaja03.1045" || accepted[0].Symbol != "CYP51G" || len(accepted[0].Sequence) != 487 {
		t.Fatalf("first=%#v", accepted[0])
	}
	if accepted[34].Status != "source-best-hit-pseudogene-label;source-gap;ambiguous-X-or-x;nonstandard-O;short-sequence" {
		t.Fatalf("exception=%#v", accepted[34])
	}
	if accepted[145].ID != "Cajacaja126288.1" || accepted[145].Symbol != "CYP82L" || len(accepted[145].Sequence) != 288 {
		t.Fatalf("middle=%#v", accepted[145])
	}
	if accepted[len(accepted)-1].ID != "Cajacaja000201.48" || accepted[len(accepted)-1].Symbol != "CYP" {
		t.Fatalf("last=%#v", accepted[len(accepted)-1])
	}
}

func TestReviewRejectsCajanusIDMismatch(t *testing.T) {
	rows := map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name"}}
	for row := 2; row <= 292; row++ {
		rows[row] = map[string]string{"A": "id", "H": "id", "I": "CYP71A1", "J": "90", "K": "CYP71A", "L": strings.Repeat("A", 500)}
	}
	for row := 293; row <= 309; row++ {
		rows[row] = map[string]string{"A": "candidate", "H": "candidate", "L": strings.Repeat("A", 200)}
	}
	rows[2]["H"] = "different"
	wb := &plantxlsx.Workbook{Dimension: "A1:L309", Rows: rows}
	if _, _, err := review(wb); err == nil {
		t.Fatal("expected A/H ID mismatch")
	}
}
