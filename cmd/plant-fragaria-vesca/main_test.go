package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesFragariaVescaLayout(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": ""}}
	for row := 2; row <= 336; row++ {
		rows[row] = map[string]string{"A": "Fragvesc.source", "H": "Fragvesc.id", "L": strings.Repeat("A", 500)}
		if row <= 332 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP71A1", "90", "CYP71A"
		}
	}
	rows[2] = map[string]string{"A": "Fragvesc4.13647", "H": "Fragvesc4.13647", "I": "CYP71AH3", "J": "59.48", "K": "CYP71AH", "L": strings.Repeat("A", 506)}
	rows[168] = map[string]string{"A": "Fragvesc5.3625", "H": "Fragvesc5.3625", "I": "CYP93A4", "J": "76.08", "K": "CYP93A", "L": strings.Repeat("A", 514)}
	rows[332] = map[string]string{"A": "Fragvesc3.12935", "H": "Fragvesc3.12935", "I": "CYP728B18", "J": "23.19", "K": "CYP", "L": strings.Repeat("A", 422)}
	rows[56]["I"], rows[56]["K"], rows[56]["L"] = "CYP71D172P", "CYP71", strings.Repeat("A", 248)+"X"+strings.Repeat("A", 249)
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: rows}
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 331 || len(excluded) != 4 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 332 {
		t.Fatalf("boundary: %d/%d", len(accepted), len(excluded))
	}
	if accepted[54].Status != "source-best-hit-pseudogene-label;ambiguous-X-or-x" {
		t.Fatalf("exception status: %#v", accepted[54])
	}
	if accepted[len(accepted)-1].Symbol != "CYP" {
		t.Fatalf("last row: %#v", accepted[len(accepted)-1])
	}
}

func TestReviewRejectsChangedFragariaVescaHeader(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: map[int]map[string]string{1: {"H": "Seq ID"}}}
	if _, _, err := planttable.Review(config, wb); err == nil {
		t.Fatal("expected header error")
	}
}
