package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesPrunusMumeLayout(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "", "H": "seq. ID", "I": "best hit ", "J": "%ID", "K": "CYP name", "L": ""}}
	for row := 2; row <= 288; row++ {
		rows[row] = map[string]string{"A": "Prunmume.source", "H": "Prunmume.id", "L": strings.Repeat("A", 500)}
		if row <= 283 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP71A1", "90", "CYP71A"
		}
	}
	rows[2] = map[string]string{"A": "Prunmume007362126.1.236", "H": "Prunmume007362126.1.236", "I": "CYP51G1", "J": "88.04", "K": "CYP51G1", "L": strings.Repeat("A", 486)}
	rows[140] = map[string]string{"A": "Prunmume007362363.1.141", "H": "Prunmume007362363.1.141", "I": "CYP89A18", "J": "65.26", "K": "CYP89A", "L": strings.Repeat("A", 515)}
	rows[283] = map[string]string{"A": "Prunmume007362430.1.83", "H": "Prunmume007362430.1.83", "I": "CYP706B1", "J": "34.07", "K": "CYP", "L": strings.Repeat("A", 341)}
	rows[98]["I"], rows[98]["K"], rows[98]["L"] = "CYP82D22P", "CYP82D", strings.Repeat("A", 224)+"X"+strings.Repeat("A", 224)
	rows[210]["L"] = strings.Repeat("A", 147) + "-" + strings.Repeat("A", 148)
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: rows}
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 282 || len(excluded) != 5 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 283 {
		t.Fatalf("boundary: %d/%d", len(accepted), len(excluded))
	}
	if accepted[0].ID != "Prunmume007362126.1.236" || accepted[0].Symbol != "CYP51G1" || len(accepted[0].Sequence) != 486 {
		t.Fatalf("first row: %#v", accepted[0])
	}
	if accepted[96].Status != "source-best-hit-pseudogene-label;ambiguous-X-or-x" {
		t.Fatalf("pseudogene row: %#v", accepted[96])
	}
	if accepted[208].Status != "source-gap;short-sequence" {
		t.Fatalf("gap row: %#v", accepted[208])
	}
	if accepted[len(accepted)-1].Symbol != "CYP" || accepted[len(accepted)-1].Status != "short-sequence" {
		t.Fatalf("last row: %#v", accepted[len(accepted)-1])
	}
	if excluded[0].Row != 284 || excluded[len(excluded)-1].Row != 288 {
		t.Fatalf("excluded boundary: %#v", excluded)
	}
}

func TestReviewRejectsChangedPrunusMumeHeaderWhitespace(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: map[int]map[string]string{1: {"H": "seq. ID", "I": "best hit"}}}
	if _, _, err := planttable.Review(config, wb); err == nil {
		t.Fatal("expected header error")
	}
}
