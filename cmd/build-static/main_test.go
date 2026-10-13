package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestGeneratePlantStaticSearch(t *testing.T) {
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
			{ID: "gene-1", RecordKey: "row-1", Category: "plants", Species: "Test plant", Symbol: "CYP71A1", Description: "a < b", Sequence: "MPEPTIDE", SourceURL: "https://example.test/1"},
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
	if m.PlantRecords != 2 || m.PlantSpecies != 1 || m.DetailShards != 1 {
		t.Fatalf("manifest=%+v", m)
	}
	search, err := os.ReadFile(filepath.Join(root, "static-data", "plants-search.js"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(search)
	if !strings.Contains(text, "CYP71A1") || !strings.Contains(text, "fallback-id") || strings.Contains(text, "CYP2D6") {
		t.Fatalf("unexpected search data: %s", text)
	}
	shard := speciesSlug("Test plant")
	detail, err := os.ReadFile(filepath.Join(root, "static-data", "details", shard+".js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(detail), `\u003c`) || !strings.Contains(string(detail), "MPEPTIDE") {
		t.Fatalf("detail data is missing escaped/full content: %s", detail)
	}
}

func TestSpeciesSlugIsStableAndCollisionResistant(t *testing.T) {
	a := speciesSlug("Plant a")
	if a != speciesSlug("Plant a") || a == speciesSlug("Plant-a") || !strings.HasPrefix(a, "plant-a-") {
		t.Fatalf("unexpected slugs: %q %q", a, speciesSlug("Plant-a"))
	}
}
