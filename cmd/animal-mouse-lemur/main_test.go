package main

import (
	"strings"
	"testing"
)

func TestReviewMouseLemurRows(t *testing.T) {
	rows := map[int]map[string]string{2: {"A": "UCSC tag", "B": "scaffold", "C": "range", "D": "Genbank hit", "K": "status", "L": "comments", "M": "Gotoh ID", "N": "Ensembl sequences", "O": "seq"}}
	for i := 3; i <= 178; i++ {
		rows[i] = map[string]string{"A": "CYP1A1", "B": "1", "C": "1-10", "K": "complete", "N": "ENSMICG1", "O": "MACDEFGHIKLMNPQRSTVWY*"}
	}
	rows[179] = map[string]string{"B": "2", "C": "2-20", "L": "CYP4X1", "N": "ENSMICG2", "O": "MA-xCD*EF*"}
	rs, e := reviewMouseLemurRows(rows)
	if e != nil {
		t.Fatal(e)
	}
	if len(rs) != 177 {
		t.Fatal(len(rs))
	}
	r := rs[len(rs)-1]
	if r.Symbol != "CYP4X1" || r.Sequence != "MA-xCD*EF" {
		t.Fatalf("%#v", r)
	}
	got := strings.Join(r.Statuses, ";")
	for _, x := range []string{"source-lowercase-retained", "terminal-stop-removed", "internal-stop-retained", "ambiguous-X-retained", "source-gap-retained"} {
		if !strings.Contains(got, x) {
			t.Fatalf("missing %s", x)
		}
	}
}
