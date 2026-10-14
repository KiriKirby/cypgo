package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

type record struct {
	ID          string
	RecordKey   string
	Category    string
	Species     string
	Symbol      string
	Description string
	Sequence    string
	SourceURL   string
}

type searchPayload struct {
	Species []string `json:"species"`
	Records [][]any  `json:"records"`
}

type detailRecord struct {
	DetailID    string `json:"detailId"`
	CYPName     string `json:"cypName"`
	ID          string `json:"id"`
	RecordKey   string `json:"recordKey"`
	Category    string `json:"category"`
	Species     string `json:"species"`
	Description string `json:"description"`
	Sequence    string `json:"sequence"`
	PhgoLite    string `json:"phgoLite"`
	SourceURL   string `json:"sourceUrl"`
}

type detailPayload struct {
	Shard   string         `json:"shard"`
	Records []detailRecord `json:"records"`
}

type manifest struct {
	SchemaVersion  int    `json:"schema_version"`
	SourceSHA256   string `json:"source_sha256"`
	PlantRecords   int    `json:"plant_records"`
	PlantSpecies   int    `json:"plant_species"`
	DetailShards   int    `json:"detail_shards"`
	GeneratedAtUTC string `json:"generated_at_utc"`
}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

func main() {
	dbPath := flag.String("db", "cypgo.pgd", "source CYP bbolt database")
	out := flag.String("out", ".", "site root containing index.html and detailtmp.html")
	flag.Parse()

	result, err := generate(*dbPath, *out)
	if err != nil {
		panic(err)
	}
	fmt.Printf("static plant search: %d records, %d species, %d detail shards\n", result.PlantRecords, result.PlantSpecies, result.DetailShards)
}

func generate(dbPath, outRoot string) (manifest, error) {
	var result manifest
	if _, err := os.Stat(filepath.Join(outRoot, "index.html")); err != nil {
		return result, fmt.Errorf("site root must contain index.html: %w", err)
	}
	if _, err := os.Stat(filepath.Join(outRoot, "detailtmp.html")); err != nil {
		return result, fmt.Errorf("site root must contain detailtmp.html: %w", err)
	}

	hash, err := fileSHA256(dbPath)
	if err != nil {
		return result, err
	}

	db, err := bolt.Open(dbPath, 0o444, &bolt.Options{ReadOnly: true, Timeout: 3 * time.Second})
	if err != nil {
		return result, err
	}
	defer db.Close()

	bySpecies := make(map[string][]detailRecord)
	order := make([]string, 0)
	seenSpecies := make(map[string]bool)
	searchRows := make([][]any, 0, 18172)
	seenIDs := make(map[string]bool)

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("records"))
		if bucket == nil {
			return errors.New("missing records bucket")
		}
		return bucket.ForEach(func(key, value []byte) error {
			var r record
			if err := json.Unmarshal(value, &r); err != nil {
				return fmt.Errorf("record %q: %w", key, err)
			}
			if !strings.EqualFold(strings.TrimSpace(r.Category), "plants") {
				return nil
			}
			species := strings.TrimSpace(r.Species)
			if species == "" {
				return fmt.Errorf("plant record %q has empty species", key)
			}
			detailID := string(key)
			if detailID == "" || seenIDs[detailID] {
				return fmt.Errorf("duplicate or empty detail ID %q", detailID)
			}
			seenIDs[detailID] = true
			name := strings.TrimSpace(r.Symbol)
			if name == "" {
				name = strings.TrimSpace(r.ID)
			}
			if name == "" {
				return fmt.Errorf("plant record %q has no displayable CYP name", key)
			}
			if !seenSpecies[species] {
				seenSpecies[species] = true
				order = append(order, species)
			}
			shard := speciesSlug(species)
			offset := len(bySpecies[species])
			bySpecies[species] = append(bySpecies[species], detailRecord{
				DetailID: detailID, CYPName: name, ID: r.ID, RecordKey: r.RecordKey,
				Category: r.Category, Species: r.Species, Description: r.Description,
				Sequence: r.Sequence, PhgoLite: phgoLiteHeader(species, r.ID, r.Symbol), SourceURL: r.SourceURL,
			})
			searchRows = append(searchRows, []any{detailID, name, species, shard, offset, sourceLabel(r.SourceURL)})
			return nil
		})
	})
	if err != nil {
		return result, err
	}
	if len(searchRows) == 0 {
		return result, errors.New("database contains no plant records")
	}

	// Species are sorted for a predictable selector; database record order is
	// preserved in search results and within each detail shard.
	sort.Strings(order)
	target := filepath.Join(outRoot, "static-data")
	temp, err := os.MkdirTemp(outRoot, ".static-data-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(temp)
	if err := os.MkdirAll(filepath.Join(temp, "details"), 0o755); err != nil {
		return result, err
	}

	search := searchPayload{Species: order, Records: searchRows}
	if err := writeJS(filepath.Join(temp, "plants-search.js"), "window.CYP_PLANT_SEARCH=", search); err != nil {
		return result, err
	}
	for _, species := range order {
		shard := speciesSlug(species)
		payload := detailPayload{Shard: shard, Records: bySpecies[species]}
		if err := writeJS(filepath.Join(temp, "details", shard+".js"), "window.CYP_PLANT_DETAILS=", payload); err != nil {
			return result, err
		}
	}

	result = manifest{
		SchemaVersion: 1, SourceSHA256: hash, PlantRecords: len(searchRows),
		PlantSpecies: len(order), DetailShards: len(order), GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeJSON(filepath.Join(temp, "build-manifest.json"), result); err != nil {
		return result, err
	}
	if err := replaceDirectory(temp, target); err != nil {
		return result, err
	}
	return result, nil
}

func sourceLabel(sourceURL string) string {
	u := strings.ToLower(strings.TrimSpace(sourceURL))
	if strings.Contains(u, "drnelson.uthsc.edu") {
		return "David R. Nelson"
	}
	if strings.Contains(u, "ricephylogenomics.ucdavis.edu") {
		return "Rice P450 Phylogenomics"
	}
	if strings.Contains(u, "rapdb.dna.naro.go.jp") {
		return "RAP-DB"
	}
	return strings.TrimSpace(sourceURL)
}

func phgoLiteHeader(species, id, symbol string) string {
	species = sanitizeLitePart(species)
	id = sanitizeLitePart(id)
	if species == "" {
		species = "~"
	}
	if id == "" {
		id = "~"
	}
	header := ">" + species + "|" + id
	symbol = strings.TrimSpace(symbol)
	if symbol == "" || symbol == "~" || symbol == "~~" {
		return header
	}
	symbol = sanitizeLitePart(symbol)
	if symbol == "" || symbol == "~" || symbol == "~~" {
		return header
	}
	return header + "(" + symbol + ")"
}

func sanitizeLitePart(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Join(strings.Fields(value), "_")
	return strings.NewReplacer("|", "_", "(", "_", ")", "_").Replace(value)
}

func speciesSlug(species string) string {
	base := strings.Trim(slugRE.ReplaceAllString(strings.ToLower(species), "-"), "-")
	if base == "" {
		base = "species"
	}
	sum := sha256.Sum256([]byte(species))
	return base + "-" + hex.EncodeToString(sum[:4])
}

func writeJS(path, prefix string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	data = append([]byte(prefix), data...)
	data = append(data, ';', '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func replaceDirectory(temp, target string) error {
	backup := target + ".previous"
	_ = os.RemoveAll(backup)
	if _, err := os.Stat(target); err == nil {
		if err := os.Rename(target, backup); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(temp, target); err != nil {
		_ = os.Rename(backup, target)
		return err
	}
	return os.RemoveAll(backup)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
