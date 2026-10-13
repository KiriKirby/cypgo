package main

import (
	"strings"
	"testing"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

func TestReviewRowsUsesCucumisLayout(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "Gotoh's seq ID", "H": "seq ID", "I": "CYP name", "J": "sequence"}}
	for row := 2; row <= 230; row++ {
		rows[row] = map[string]string{"A": "Cucusati.source", "H": "Cucusati.id", "I": "CYP71A1", "J": strings.Repeat("A", 500)}
	}
	rows[2]["A"], rows[2]["H"], rows[2]["I"] = "Cucusati00919.1591", "Cucusati00919.1591", "CYP51G1"
	rows[116]["A"], rows[116]["H"], rows[116]["I"] = "Cucusati03487.421", "Cucusati03487.421", "CYP87D19"
	rows[230]["A"], rows[230]["H"], rows[230]["I"] = "Cucusati02653.1189", "Cucusati02653.1189", "CYP749A47"
	rows[4]["I"], rows[4]["J"] = "CYP71B81P", strings.Repeat("A", 200)+"*"+strings.Repeat("A", 300)
	rows[17]["I"], rows[17]["J"] = "CYP71AN38", strings.Repeat("A", 499)+"*"
	rows[147]["J"] = " " + strings.Repeat("A", 500)
	wb := &plantxlsx.Workbook{Dimension: "A1:J230", Rows: rows}
	got, err := reviewRows(wb)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 229 || got[0].Row != 2 || got[len(got)-1].Row != 230 {
		t.Fatalf("boundary: %d", len(got))
	}
	if got[0].SeqID != "Cucusati00919.1591" || got[0].Symbol != "CYP51G1" || len(got[0].Sequence) != 500 {
		t.Fatalf("first row: %#v", got[0])
	}
	if got[2].Status != "source-pseudogene-label;internal-stop" || !strings.Contains(got[2].Sequence, "*") {
		t.Fatalf("internal stop: %#v", got[2])
	}
	if got[15].Status != "source-terminal-stop" || strings.Contains(got[15].Sequence, "*") || len(got[15].Sequence) != 499 {
		t.Fatalf("terminal stop: %#v", got[15])
	}
	if len(got[145].Sequence) != 500 || got[145].Sequence[0] != 'A' {
		t.Fatalf("cell whitespace: %#v", got[145])
	}
}

func TestReviewRowsRequiresExactCucumisHeader(t *testing.T) {
	wb := &plantxlsx.Workbook{Dimension: "A1:J230", Rows: map[int]map[string]string{1: {"A": "seq ID"}}}
	if _, err := reviewRows(wb); err == nil {
		t.Fatal("expected header error")
	}
}
