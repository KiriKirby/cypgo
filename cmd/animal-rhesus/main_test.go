package main

import (
	"strings"
	"testing"
)

func TestReviewRhesusRowsPreservesFinalNamedRows(t *testing.T) {
	rows := map[int]map[string]string{
		1: {
			"A": "Gotoh's ID", "B": "Ensembl gene ID", "C": "Ensembl transcript ID", "D": "chromosome",
			"E": "start", "F": "stop", "G": "strand", "H": "best hit in homepage list (49 seqs)",
			"I": "%ID", "J": "comment/CYP name", "K": "sequence", "L": "",
		},
	}
	for row := 2; row <= 111; row++ {
		rows[row] = map[string]string{"A": "MacamulaUn_1.1", "B": "ENSMMUG00000000001", "C": "ENSMMUT00000000001", "D": "1", "E": "10", "F": "20", "G": "1", "H": "CYP1A1_Rhesus", "I": "100", "J": "CYP1A1 comp.", "K": "MACDEFGHIKLMNPQRSTVWY*"}
	}
	rows[26]["B"] = "not in BIOMART set"
	rows[26]["J"] = "CYP2C106P my set"
	rows[26]["K"] = "MAxxCDE*"
	records, err := reviewRhesusRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 110 {
		t.Fatalf("records=%d", len(records))
	}
	r := records[24]
	if r.ID != "MacamulaUn_1.1" || r.Sequence != "MAxxCDE" {
		t.Fatalf("fallback ID/sequence mismatch: %#v", r)
	}
	joined := strings.Join(r.Statuses, ";")
	for _, want := range []string{"source-lowercase-retained", "terminal-stop-removed", "ambiguous-X-retained", "short-source-fragment"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("status %q missing from %q", want, joined)
		}
	}
}

func TestRhesusChangedHeaderIsRejected(t *testing.T) {
	if _, err := reviewRhesusRows(map[int]map[string]string{1: {"A": "wrong"}}); err == nil {
		t.Fatal("expected changed header to fail")
	}
}
