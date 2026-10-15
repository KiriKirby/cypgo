package main

import "testing"

func TestPigResidues(t *testing.T) {
	for _, x := range []string{"MACDEFGHIK*", "gyhipkdtvv", "MA-XCDEFGH"} {
		if !pigResidues(x) {
			t.Fatal(x)
		}
	}
	for _, x := range []string{"(1)", "complete", "123"} {
		if pigResidues(x) {
			t.Fatal(x)
		}
	}
}
