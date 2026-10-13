package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesCannabisLegendBoundary(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": ""}, 422: {}, 423: {"E": "green name = no hits in blast (64 seqs. out of frame)"}}
	for row := 2; row <= 421; row++ {
		rows[row] = map[string]string{"A": "Cannsati.source", "H": "Cannsati.id", "L": strings.Repeat("A", 500)}
		if row <= 357 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP71A1", "90", "CYP71A"
		}
	}
	rows[2] = map[string]string{"A": "Cannsati32097775.1", "H": "Cannsati32097775.1", "I": "CYP51G1", "J": "87.63", "K": "CYP51G1", "L": strings.Repeat("A", 486)}
	rows[179] = map[string]string{"A": "Cannsati92824.3", "H": "Cannsati92824.3", "I": "CYP81Q57", "J": "50", "K": "CYP81", "L": strings.Repeat("A", 490)}
	rows[357] = map[string]string{"A": "Cannsati114114.1", "H": "Cannsati114114.1", "I": "CYP55B1", "J": "26.64", "K": "CYP", "L": strings.Repeat("A", 261)}
	rows[76]["I"], rows[76]["K"], rows[76]["L"] = "CYP71D391P", "CYP71", strings.Repeat("A", 179)+"-"+strings.Repeat("A", 180)
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: rows}
	accepted, excluded, err := review(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 356 || len(excluded) != 64 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 357 {
		t.Fatalf("boundary: %d/%d", len(accepted), len(excluded))
	}
	if accepted[74].Status != "source-best-hit-pseudogene-label;source-gap" {
		t.Fatalf("exception status: %#v", accepted[74])
	}
	if accepted[len(accepted)-1].Status != "short-sequence" || accepted[len(accepted)-1].Symbol != "CYP" {
		t.Fatalf("last row: %#v", accepted[len(accepted)-1])
	}
}

func TestReviewRejectsChangedCannabisLegend(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: map[int]map[string]string{423: {"E": "64 sequences"}}}
	if _, _, err := review(wb); err == nil {
		t.Fatal("expected legend error")
	}
}
