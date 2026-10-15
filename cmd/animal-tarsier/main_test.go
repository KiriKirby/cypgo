package main

import (
	"strings"
	"testing"
)

func TestReviewTarsierCombinedRows(t *testing.T) {
	rows := map[int]map[string]string{1: {"A": "Gene ID", "B": "transcript ID", "C": "start", "D": "end", "E": "", "F": "best hit human", "G": "%ID", "H": "Gotoh ID", "I": "CYP name", "J": "comments", "K": "seq", "L": ""}}
	for i := 2; i <= 109; i++ {
		rows[i] = map[string]string{"A": "ENSTSYG1", "B": "ENSTSYT1", "F": "CYP1A1", "G": "90", "H": "1.1", "I": "CYP1A1", "J": "complete", "K": "MACDEFGHIKLMNPQRSTVWY*"}
	}
	rows[43]["A"] = ""
	rows[43]["H"] = "TarssyriUn_1.1"
	rows[43]["I"] = "CYP2D76P"
	rows[43]["J"] = "partial"
	rows[43]["K"] = "MAxCD*EF*"
	rs, e := reviewTarsierRows(rows)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 108 {
		t.Fatalf("count=%d", len(rs))
	}
	r := rs[41]
	if r.ID != "TarssyriUn_1.1" || r.Sequence != "MAxCD*EF" {
		t.Fatalf("%#v", r)
	}
	got := strings.Join(r.Statuses, ";")
	for _, want := range []string{"source-lowercase-retained", "terminal-stop-removed", "internal-stop-retained", "ambiguous-X-retained", "source-partial-or-missing-region"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s in %s", want, got)
		}
	}
}
func TestTarsierHeaderRejected(t *testing.T) {
	if _, e := reviewTarsierRows(map[int]map[string]string{1: {"A": "wrong"}}); e == nil {
		t.Fatal("expected error")
	}
}
