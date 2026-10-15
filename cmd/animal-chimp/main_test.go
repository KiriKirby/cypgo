package main

import (
	"strings"
	"testing"
)

func TestReviewChimpRowsUsesOnlyFinalCuratedSheetLayout(t *testing.T) {
	rows := map[int]map[string]string{
		1: {
			"A": "chimp Ensembl gene ID", "B": "transcript ID", "C": "start", "D": "stop ",
			"E": "chr", "F": "strand", "G": "Best hit human", "H": "%ID",
			"I": "Best hit Gotohs seqs", "J": "%ID", "K": "comment", "L": "CYP name", "M": "sequence",
		},
	}
	for row := 2; row <= 122; row++ {
		rows[row] = map[string]string{
			"A": "ENSPTRG" + strings.Repeat("0", 9) + "1", "B": "ENSPTRT1", "C": "100", "D": "200",
			"E": "1", "F": "1", "G": "CYP1A1", "H": "99", "I": "Pan_trogUn_1.100,1:20,1-20",
			"J": "100", "L": "CYP1A1", "M": "MACDEFGHIKLMNPQRSTVWY*",
		}
	}
	rows[10]["A"] = ""
	rows[10]["I"] = "Pan_trogUn_11.46187,1:421,1-421"
	rows[10]["L"] = "CYP2B6 partial seq"
	rows[10]["M"] = "MAcde fGHIK*LMN*"
	records, err := reviewChimpRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 121 {
		t.Fatalf("records=%d", len(records))
	}
	r := records[8]
	if r.ID != "Pan_trogUn_11.46187" || r.Sequence != "MAcdefGHIK*LMN" {
		t.Fatalf("fallback ID/sequence mismatch: %#v", r)
	}
	joined := strings.Join(r.Statuses, ";")
	for _, want := range []string{"source-whitespace-joined", "source-lowercase-retained", "terminal-stop-removed", "internal-stop-retained", "source-partial-or-missing-region", "duplicate-symbol-source-row"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("status %q missing from %q", want, joined)
		}
	}
}

func TestChimpHeaderChangeIsRejected(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "wrong"}}
	if _, err := reviewChimpRows(rows); err == nil {
		t.Fatal("expected changed header to fail")
	}
}
