package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewRowsUsesHeaderlessCitrulusLayout(t *testing.T) {
	rows := map[int]map[string]string{}
	for row := 1; row <= 244; row++ {
		rows[row] = map[string]string{"A": "Citrlana.source", "H": "Citrlana.id", "L": strings.Repeat("A", 500)}
		if row <= 233 {
			rows[row]["I"] = "CYP71A1"
			rows[row]["J"] = "90.0"
			rows[row]["K"] = "CYP71A"
		}
	}
	rows[1] = map[string]string{"A": "Citrlana13966.2", "H": "Citrlana13966.2", "I": "CYP51G1", "J": "96.3", "K": "CYP51G1", "L": strings.Repeat("A", 486)}
	rows[117] = map[string]string{"A": "Citrlana00824.53", "H": "Citrlana00824.53", "I": "CYP84A17", "J": "90.0", "K": "CYP84A", "L": strings.Repeat("A", 523)}
	rows[233] = map[string]string{"A": "Citrlana09587.6", "H": "Citrlana09587.6", "I": "CYP749A47", "J": "85.77", "K": "CYP749A", "L": strings.Repeat("A", 530)}
	rows[2]["I"], rows[2]["L"] = "CYP71B81P", strings.Repeat("A", 300)+"-"+strings.Repeat("A", 199)
	wb := &plantxlsx.Workbook{Dimension: "A1:L244", Rows: rows}
	accepted, excluded, err := reviewRows(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted) != 233 || len(excluded) != 11 || accepted[0].Row != 1 || accepted[len(accepted)-1].Row != 233 {
		t.Fatalf("boundary: %d/%d", len(accepted), len(excluded))
	}
	if accepted[0].Symbol != "CYP51G1" || len(accepted[0].Sequence) != 486 {
		t.Fatalf("first row: %#v", accepted[0])
	}
	if accepted[1].Status != "source-best-hit-pseudogene-label;source-gap" {
		t.Fatalf("best-hit status: %#v", accepted[1])
	}
	if excluded[0].Row != 234 || excluded[len(excluded)-1].Row != 244 {
		t.Fatalf("excluded boundary: %#v %#v", excluded[0], excluded[len(excluded)-1])
	}
}

func TestReviewRowsRejectsAssignmentInExcludedRegion(t *testing.T) {
	rows := map[int]map[string]string{}
	for row := 1; row <= 244; row++ {
		rows[row] = map[string]string{"A": "source", "H": "id", "L": strings.Repeat("A", 500)}
		if row <= 233 {
			rows[row]["I"], rows[row]["J"], rows[row]["K"] = "CYP1A1", "90", "CYP1A"
		}
	}
	rows[234]["K"] = "CYP1A"
	if _, _, err := reviewRows(&plantxlsx.Workbook{Dimension: "A1:L244", Rows: rows}); err == nil {
		t.Fatal("expected excluded assignment error")
	}
}
