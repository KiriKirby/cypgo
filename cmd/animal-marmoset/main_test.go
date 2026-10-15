package main

import (
	"strings"
	"testing"
)

func TestReviewMarmosetRowsPreservesWorkbookSequence(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "Ensembl ID", "B": "Location chr:position:strand", "C": "CYP name", "D": "species", "E": "sequence"}}
	for row := 2; row <= 88; row++ {
		rows[row] = map[string]string{"A": "ENSCJAG1", "B": "1:10-20:1", "C": "CYP1A1 ortholog", "D": "marmoset", "E": "MACDEFGHIKLMNPQRSTVWY*"}
	}
	rows[42]["A"] = ""
	rows[42]["C"] = "CYP4A49P (with a gap)"
	rows[42]["E"] = "MA-xCD*"
	records, err := reviewMarmosetRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 87 {
		t.Fatalf("records=%d", len(records))
	}
	r := records[40]
	if r.ID != "1:10-20:1" || r.Sequence != "MA-xCD" {
		t.Fatalf("row=%#v", r)
	}
	joined := strings.Join(r.Statuses, ";")
	for _, want := range []string{"source-lowercase-retained", "terminal-stop-removed", "ambiguous-X-retained", "source-gap-retained", "source-partial-or-missing-region"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %s in %s", want, joined)
		}
	}
}

func TestMarmosetHeaderChangeRejected(t *testing.T) {
	if _, err := reviewMarmosetRows(map[int]map[string]string{1: {"A": "wrong"}}); err == nil {
		t.Fatal("expected error")
	}
}
