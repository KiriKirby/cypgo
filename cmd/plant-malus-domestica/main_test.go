package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/planttable"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewUsesMalusDomesticaLayout(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "", "H": "seq. ID", "I": "best hit", "J": "%ID", "K": "CYP name", "L": ""}}
	for row := 2; row <= 349; row++ {
		rows[row] = map[string]string{"A": "Maludome.source", "H": "Maludome.id", "L": strings.Repeat("A", 500)}
		if row <= 332 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP71A1", "90", "CYP71A"
		}
	}
	rows[2] = map[string]string{"A": "Maludome13.10293", "H": "Maludome13.10293", "I": "CYP51G1", "J": "88.02", "K": "CYP51G", "L": strings.Repeat("A", 486)}
	rows[167] = map[string]string{"A": "Maludome6.19040", "H": "Maludome6.19040", "I": "CYP89A98", "J": "70", "K": "CYP89A", "L": strings.Repeat("A", 518)}
	rows[332] = map[string]string{"A": "Maludome13.7961", "H": "Maludome13.7961", "I": "CYP76F47", "J": "33.18", "K": "CYP", "L": strings.Repeat("A", 422)}
	rows[49]["I"], rows[49]["K"] = "CYP71AU6P", "CYP71"
	rows[315]["L"] = strings.Repeat("A", 164) + "-" + strings.Repeat("A", 165)
	wb := &plantxlsx.Workbook{Dimension: config.Dimension, Rows: rows}
	accepted, excluded, err := planttable.Review(config, wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 331 || len(excluded) != 17 || accepted[0].Row != 2 || accepted[len(accepted)-1].Row != 332 {
		t.Fatalf("boundary: %d/%d", len(accepted), len(excluded))
	}
	if accepted[0].ID != "Maludome13.10293" || accepted[0].Symbol != "CYP51G" || len(accepted[0].Sequence) != 486 {
		t.Fatalf("first row: %#v", accepted[0])
	}
	if accepted[47].Status != "source-best-hit-pseudogene-label" {
		t.Fatalf("pseudogene row: %#v", accepted[47])
	}
	if accepted[313].Status != "source-gap;short-sequence" {
		t.Fatalf("gap row: %#v", accepted[313])
	}
	if accepted[len(accepted)-1].Symbol != "CYP" {
		t.Fatalf("last row: %#v", accepted[len(accepted)-1])
	}
	if excluded[0].Row != 333 || excluded[len(excluded)-1].Row != 349 {
		t.Fatalf("excluded boundary: %#v", excluded)
	}
}

func TestReviewRejectsChangedMalusDomesticaBoundary(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: "A1:L350", Rows: map[int]map[string]string{1: {"H": "seq. ID"}}}
	if _, _, err := planttable.Review(config, wb); err == nil {
		t.Fatal("expected dimension error")
	}
}
