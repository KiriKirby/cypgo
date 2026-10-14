package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestGenerateStaticSearchForAllCategories(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "detailtmp.html"), []byte("detail"), 0o644); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(root, "test.pgd")
	db, err := bolt.Open(dbPath, 0o644, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("records"))
		if err != nil {
			return err
		}
		rows := []record{
			{ID: "gene-1", RecordKey: "row-1", Category: "plants", Species: "Test plant", Symbol: "CYP71A1", Description: "a < b", Sequence: "MPEPTIDE", SourceURL: "https://drnelson.uthsc.edu/example"},
			{ID: "fallback-id", Category: "plants", Species: "Test plant", Sequence: "AAAA"},
			{ID: "animal-1", Category: "animals", Species: "Test animal", Symbol: "CYP2D6"},
		}
		for i, row := range rows {
			data, err := json.Marshal(row)
			if err != nil {
				return err
			}
			if err := b.Put([]byte{'0' + byte(i)}, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	m, err := generate(dbPath, root)
	if err != nil {
		t.Fatal(err)
	}
	if m.PlantRecords != 2 || m.PlantSpecies != 1 || m.TotalRecords != 3 || m.TotalSpecies != 2 || m.DetailShards != 2 {
		t.Fatalf("manifest=%+v", m)
	}
	search, err := os.ReadFile(filepath.Join(root, "static-data", "plants-search.js"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(search)
	if !strings.Contains(text, "CYP71A1") || !strings.Contains(text, "fallback-id") || !strings.Contains(text, "David R. Nelson") || strings.Contains(text, "CYP2D6") {
		t.Fatalf("unexpected search data: %s", text)
	}
	animalSearch, err := os.ReadFile(filepath.Join(root, "static-data", "animals-search.js"))
	if err != nil || !strings.Contains(string(animalSearch), "CYP2D6") {
		t.Fatalf("animal search data=%q err=%v", animalSearch, err)
	}
	emptySearch, err := os.ReadFile(filepath.Join(root, "static-data", "fungi-search.js"))
	if err != nil || !strings.Contains(string(emptySearch), `"species":[]`) || !strings.Contains(string(emptySearch), `"records":[]`) {
		t.Fatalf("empty category search data=%q err=%v", emptySearch, err)
	}
	shard := speciesSlug("Test plant")
	detail, err := os.ReadFile(filepath.Join(root, "static-data", "details", shard+".js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(detail), `\u003c`) || !strings.Contains(string(detail), "MPEPTIDE") {
		t.Fatalf("detail data is missing escaped/full content: %s", detail)
	}
	if !strings.Contains(string(detail), "phgoLite") || !strings.Contains(string(detail), `\u003eTest_plant|gene-1(CYP71A1)`) {
		t.Fatalf("detail data is missing PHgo Lite header: %s", detail)
	}
	animalShard := categorySpeciesShard("animals", "Test animal")
	if _, err := os.Stat(filepath.Join(root, "static-data", "details", animalShard+".js")); err != nil {
		t.Fatal(err)
	}
}

func TestSourceLabel(t *testing.T) {
	tests := map[string]string{
		"https://drnelson.uthsc.edu/x":                   "David R. Nelson",
		"https://ricephylogenomics.ucdavis.edu/p450/x":   "Rice P450 Phylogenomics",
		"https://rapdb.dna.naro.go.jp/download/irgsp1/x": "RAP-DB",
		"https://example.test/x":                         "https://example.test/x",
	}
	for source, want := range tests {
		if got := sourceLabel(source, ""); got != want {
			t.Fatalf("sourceLabel(%q)=%q want %q", source, got, want)
		}
	}
	if got := sourceLabel("https://drnelson.uthsc.edu/x", "bibliod:block-1"); got != "David R. Nelson biblioD" {
		t.Fatalf("biblioD label=%q", got)
	}
}

func TestPhgoLiteHeaderSanitizesFieldsAndOmitsEmptySymbol(t *testing.T) {
	if got := phgoLiteHeader("Test plant", "gene (1)", "~"); got != ">Test_plant|gene__1_" {
		t.Fatalf("header=%q", got)
	}
}

func TestSpeciesSlugIsStableAndCollisionResistant(t *testing.T) {
	a := speciesSlug("Plant a")
	if a != speciesSlug("Plant a") || a == speciesSlug("Plant-a") || !strings.HasPrefix(a, "plant-a-") {
		t.Fatalf("unexpected slugs: %q %q", a, speciesSlug("Plant-a"))
	}
}
