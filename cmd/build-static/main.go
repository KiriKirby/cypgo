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
	"strconv"
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
	SchemaVersion  int                      `json:"schema_version"`
	SourceSHA256   string                   `json:"source_sha256"`
	PlantRecords   int                      `json:"plant_records"`
	PlantSpecies   int                      `json:"plant_species"`
	TotalRecords   int                      `json:"total_records"`
	TotalSpecies   int                      `json:"total_species"`
	DetailShards   int                      `json:"detail_shards"`
	Categories     map[string]categoryStats `json:"categories"`
	GeneratedAtUTC string                   `json:"generated_at_utc"`
}

type categoryStats struct {
	Records int `json:"records"`
	Species int `json:"species"`
}

var categoryOrder = []string{"animals", "plants", "fungi", "bacteria", "protists"}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

func main() {
	dbPath := flag.String("db", "cypgo.pgd", "source CYP bbolt database")
	out := flag.String("out", ".", "site root containing index.html and detailtmp.html")
	flag.Parse()

	result, err := generate(*dbPath, *out)
	if err != nil {
		panic(err)
	}
	fmt.Printf("static CYP search: %d records, %d species, %d detail shards\n", result.TotalRecords, result.TotalSpecies, result.DetailShards)
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

	byShard := make(map[string][]detailRecord)
	speciesByCategory := make(map[string][]string)
	seenSpecies := make(map[string]bool)
	searchRowsByCategory := make(map[string][][]any)
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
			category := strings.ToLower(strings.TrimSpace(r.Category))
			if !knownCategory(category) {
				return nil
			}
			species := strings.TrimSpace(r.Species)
			if species == "" {
				return fmt.Errorf("%s record %q has empty species", category, key)
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
				return fmt.Errorf("%s record %q has no displayable CYP name", category, key)
			}
			speciesKey := category + "\x00" + species
			if !seenSpecies[speciesKey] {
				seenSpecies[speciesKey] = true
				speciesByCategory[category] = append(speciesByCategory[category], species)
			}
			shard := categorySpeciesShard(category, species)
			offset := len(byShard[shard])
			byShard[shard] = append(byShard[shard], detailRecord{
				DetailID: detailID, CYPName: name, ID: r.ID, RecordKey: r.RecordKey,
				Category: r.Category, Species: r.Species, Description: r.Description,
				Sequence: r.Sequence, PhgoLite: phgoLiteHeader(species, r.ID, r.Symbol), SourceURL: r.SourceURL,
			})
			searchRowsByCategory[category] = append(searchRowsByCategory[category], []any{detailID, name, species, shard, offset, sourceLabel(r.SourceURL, r.RecordKey)})
			return nil
		})
	})
	if err != nil {
		return result, err
	}
	if len(searchRowsByCategory) == 0 {
		return result, errors.New("database contains no supported CYP records")
	}

	// Species are sorted for a predictable selector; database record order is
	// preserved in search results and within each detail shard.
	for _, category := range categoryOrder {
		sort.Strings(speciesByCategory[category])
	}
	target := filepath.Join(outRoot, "static-data")
	temp, err := os.MkdirTemp(outRoot, ".static-data-")
	if err != nil {
		return result, err
	}
	defer os.RemoveAll(temp)
	if err := os.MkdirAll(filepath.Join(temp, "details"), 0o755); err != nil {
		return result, err
	}

	stats := make(map[string]categoryStats, len(categoryOrder))
	for _, category := range categoryOrder {
		species := speciesByCategory[category]
		if species == nil {
			species = []string{}
		}
		rows := searchRowsByCategory[category]
		if rows == nil {
			rows = [][]any{}
		}
		search := searchPayload{Species: species, Records: rows}
		prefix := "window.CYP_SEARCH_DATA=window.CYP_SEARCH_DATA||{};window.CYP_SEARCH_DATA[" + strconv.Quote(category) + "]="
		if err := writeJS(filepath.Join(temp, category+"-search.js"), prefix, search); err != nil {
			return result, err
		}
		stats[category] = categoryStats{Records: len(search.Records), Species: len(search.Species)}
		result.TotalRecords += len(search.Records)
		result.TotalSpecies += len(search.Species)
	}
	for shard, records := range byShard {
		payload := detailPayload{Shard: shard, Records: records}
		if err := writeJS(filepath.Join(temp, "details", shard+".js"), "window.CYP_DETAILS=", payload); err != nil {
			return result, err
		}
	}

	result.SchemaVersion = 2
	result.SourceSHA256 = hash
	result.PlantRecords = stats["plants"].Records
	result.PlantSpecies = stats["plants"].Species
	result.DetailShards = len(byShard)
	result.Categories = stats
	result.GeneratedAtUTC = time.Now().UTC().Format(time.RFC3339)
	if err := writeJSON(filepath.Join(temp, "build-manifest.json"), result); err != nil {
		return result, err
	}
	if err := replaceDirectory(temp, target); err != nil {
		return result, err
	}
	return result, nil
}

func knownCategory(category string) bool {
	for _, candidate := range categoryOrder {
		if category == candidate {
			return true
		}
	}
	return false
}

func categorySpeciesShard(category, species string) string {
	if category == "plants" {
		return speciesSlug(species)
	}
	return category + "-" + speciesSlug(species)
}

func sourceLabel(sourceURL, recordKey string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(recordKey)), "bibliod:") {
		return "David R. Nelson biblioD"
	}
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
