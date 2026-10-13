package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesUpdatedPeachLayout(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": ""}}
	for row := 2; row <= 326; row++ {
		rows[row] = map[string]string{"A": "Prunpers.source", "H": "Prunpers.id", "L": strings.Repeat("A", 500)}
		if row <= 320 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP71A1", "90", "CYP71A"
		}
	}
	rows[2] = map[string]string{"A": "Prunpers1.21767", "H": "Prunpers1.21767", "I": "CYP51G1", "J": "88.04", "K": "CYP51G1", "L": strings.Repeat("A", 486)}
	rows[161] = map[string]string{"A": "Prunpers2.25113", "H": "Prunpers2.25113", "I": "CYP89A15v1", "J": "90", "K": "CYP89A", "L": strings.Repeat("A", 485)}
	rows[320] = map[string]string{"A": "Prunpers1.28002", "H": "Prunpers1.28002", "I": "CYP749A9", "J": "53.16", "K": "CYP749A", "L": strings.Repeat("A", 508)}
	rows[63]["I"], rows[63]["L"] = "CYP79A1P", strings.Repeat("A", 250)+"-"+strings.Repeat("A", 249)
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: rows}
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 319 || len(excluded) != 6 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 320 {
		t.Fatalf("boundary: %d/%d", len(accepted), len(excluded))
	}
	if accepted[0].ID != "Prunpers1.21767" || accepted[0].Symbol != "CYP51G1" || len(accepted[0].Sequence) != 486 {
		t.Fatalf("first row: %#v", accepted[0])
	}
	if accepted[61].Status != "source-best-hit-pseudogene-label;source-gap" {
		t.Fatalf("exception status: %#v", accepted[61])
	}
	if excluded[0].Row != 321 || excluded[len(excluded)-1].Row != 326 {
		t.Fatalf("excluded boundary: %#v", excluded)
	}
}

func TestReviewRejectsChangedPeachHeader(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: map[int]map[string]string{1: {"H": "seq ID"}}}
	if _, _, err := planttable.Review(config, wb); err == nil {
		t.Fatal("expected header error")
	}
}
