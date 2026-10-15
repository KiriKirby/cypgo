package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestProfiles(t *testing.T) {
	for _, p := range profiles {
		rs, _, e := parse(p, filepath.Join("../../raw", strings.TrimSuffix(p.File, filepath.Ext(p.File))+".txt"))
		if e != nil {
			t.Fatal(e)
		}
		if len(rs) != p.Expected {
			t.Fatal(p.Slug)
		}
		for _, i := range []int{0, len(rs) / 2, len(rs) - 1} {
			if rs[i].Header == "" || rs[i].ID == "" || rs[i].Key == "" {
				t.Fatalf("%s representative record %d incomplete: %#v", p.Slug, i, rs[i])
			}
		}
	}
}

func TestPublicInsectInlineProtein(t *testing.T) {
	h := "CYP4G7 Tribolium castaneum XP_975951 MVVVEESLNHSLNLGNSVLISLGVVAVILAVYHFWLQSLRYTKLGNKIPGYDPLPIIGNALMVNKNPTQVMELALRVASEKGSVVRFWFGSKLGVALLDPRDVEIILSSNVHLEKSSEYRFFEPWLGDGLLISKGDKWR"
	if got := publicInsectInlineProtein(h); len(got) < 100 || !strings.HasPrefix(got, "MVVV") {
		t.Fatal(got)
	}
}

func TestPublicInsectSpecies(t *testing.T) {
	cases := map[string]string{
		"CYP4G7 Tribolium castaneum (red flour beetle) XP_975951": "Tribolium castaneum",
		"CYP4C22v1 Culex pipiens pallens (mosquito) AF157093":     "Culex pipiens pallens",
		"Cyp6a17 D. melanogaster AC020447":                        "Public insect collection",
		"CYP9X1 probable ortholog from unknown scaffold":          "Public insect collection",
	}
	for h, want := range cases {
		if got := publicInsectSpecies(h); got != want {
			t.Fatalf("%q: got %q want %q", h, got, want)
		}
	}
}
