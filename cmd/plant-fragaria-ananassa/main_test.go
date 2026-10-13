package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesFragariaAnanassaAColumnIDs(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "Gotoh ID", "H": "Blast output seq ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": "sequence"}}
	for row := 2; row <= 225; row++ {
		id := "Fraganan.test"
		rows[row] = map[string]string{"A": ">" + id, "H": id, "L": strings.Repeat("A", 500)}
		if row <= 207 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP71A1", "90", "CYP71A"
		} else {
			rows[row]["H"], rows[row]["I"], rows[row]["K"] = "opposite strand", "no hit", "no hit"
		}
	}
	rows[2] = map[string]string{"A": ">Fraganan_rscf00000033.1.21", "H": "Fraganan_rscf00000033.1.21", "I": "CYP51G1", "J": "83.74", "K": "CYP51G1", "L": strings.Repeat("A", 487)}
	rows[104] = map[string]string{"A": ">Fraganan_rscf00001030.1.4", "H": "Fraganan_rscf00001030.1.4", "I": "CYP92A84", "J": "70", "K": "CYP92A", "L": strings.Repeat("A", 513)}
	rows[207] = map[string]string{"A": ">Fraganan_icon00003587_a.1.1rev", "H": "opposite strand", "I": "CYP88A50", "J": "60", "K": "CYP88A", "L": strings.Repeat("A", 251)}
	rows[27]["I"], rows[27]["K"], rows[27]["L"] = "CYP72A263P", "CYP72A", strings.Repeat("A", 80)+"-XO"+strings.Repeat("A", 84)
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: rows}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 206 || len(excluded) != 18 || accepted[0].ID != "Fraganan_rscf00000033.1.21" || accepted[len(accepted)-1].ID != "Fraganan_icon00003587_a.1.1rev" {
		t.Fatalf("boundary or IDs changed: %d/%d %#v %#v", len(accepted), len(excluded), accepted[0], accepted[len(accepted)-1])
	}
	if accepted[25].Status != "source-best-hit-pseudogene-label;source-gap;ambiguous-X-or-x;nonstandard-O;short-sequence" {
		t.Fatalf("exception status: %#v", accepted[25])
	}
	if accepted[len(accepted)-1].Status != "short-sequence" {
		t.Fatalf("opposite-strand assigned row: %#v", accepted[len(accepted)-1])
	}
	if excluded[0].Row != 208 || excluded[len(excluded)-1].Row != 225 {
		t.Fatalf("excluded boundary: %#v", excluded)
	}
}

func TestReviewRejectsOppositeStrandWithoutAssignment(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: map[int]map[string]string{1: config.Headers, 2: {"A": ">x", "H": "opposite strand", "I": "no hit", "K": "no hit", "L": strings.Repeat("A", 100)}}}
	if _, _, err := review(wb); err == nil {
		t.Fatal("expected assigned-row error")
	}
}
