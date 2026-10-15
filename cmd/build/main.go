package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

type record struct {
	ID, RecordKey, Category, Species, Symbol, Description, Sequence, SourceURL string
	ReviewStatus                                                               string `json:"-"`
}
type speciesRecord struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Selectable  bool   `json:"selectable"`
	Description string `json:"description"`
}

var pages = map[string]string{
	"animals":  "https://drnelson.uthsc.edu/animals/",
	"plants":   "https://drnelson.uthsc.edu/plants/",
	"fungi":    "https://drnelson.uthsc.edu/fungal-genomes/",
	"bacteria": "https://drnelson.uthsc.edu/bacteria/",
}

var bracketSpeciesRE = regexp.MustCompile(`\[([^\]]{3,120})\]`)
var speciesShapeRE = regexp.MustCompile(`^(?:[A-Z][a-z]+|[a-z][a-z]+)\s+(?:[a-z][A-Za-z-]+|sp\.?|cf\.?|aff\.?)\b`)

func main() {
	out := flag.String("out", "cypgo.pgd", "output PGD path")
	sources := flag.String("sources", "sources", "reviewed structured source directory")
	docs := flag.String("species-docs", "docs/species", "generated per-species audit documentation")
	resources := flag.String("resource-index", "sources/resources.csv", "resource manifest CSV")
	flag.Parse()
	_, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	// Formal records come only from reviewed, resource-specific structured
	// inputs. The legacy parseExtracted helper is retained temporarily for
	// comparison tests, but must never contribute records to a published PGD.
	records, err := readReviewedSources(*sources)
	if err != nil {
		panic(err)
	}
	allSpecies, err := readManifestSpecies(*resources)
	if err != nil {
		panic(err)
	}
	if err := write(*out, records, allSpecies); err != nil {
		panic(err)
	}
	if err := writeSpeciesDocs(*docs, records); err != nil {
		panic(err)
	}
	if err := writeResourceAuditDocs(*resources, filepath.Join(filepath.Dir(*docs), "resources"), records); err != nil {
		panic(err)
	}
	if err := writeAuditSummary(filepath.Join(filepath.Dir(*docs), "AUDIT_SUMMARY.md"), records, allSpecies); err != nil {
		panic(err)
	}
	if err := writePlantAuditDocs(*resources, filepath.Join(filepath.Dir(*docs), "plants"), records); err != nil {
		panic(err)
	}
	if err := writePlantRecordAudit(*resources, filepath.Join(filepath.Dir(*docs), "plants", "plant-record-audit.csv"), records); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %d records to %s\n", len(records), *out)
}

func writeAuditSummary(path string, records []record, species []speciesRecord) error {
	counts := map[string][2]int{}
	for _, r := range records {
		v := counts[r.Category]
		v[0]++
		if strings.TrimSpace(r.Sequence) != "" {
			v[1]++
		}
		counts[r.Category] = v
	}
	var b strings.Builder
	b.WriteString("# Reviewed Dr. Nelson resource audit\n\n")
	fmt.Fprintf(&b, "This summary is generated only from completed resource-specific structured inputs. Listed but unfinished resources remain disabled and do not contribute records. No external sequence database is consulted.\n\nListed species/resource entries: %d\n\n| Category | Reviewed PGD records | Literal source sequences |\n|---|---:|---:|\n", len(species))
	for _, cat := range []string{"animals", "plants", "fungi", "bacteria", "protists"} {
		v := counts[cat]
		fmt.Fprintf(&b, "| %s | %d | %d |\n", cat, v[0], v[1])
	}
	b.WriteString("\nRelease acceptance rules:\n\n- Each resource must have its own reviewed parser or reviewed structured CSV.\n- Source blocks/rows, duplicate names, species relationships, fragments, pseudogenes, and comparison records are decided for that exact file.\n- Scores, alignment snippets, comments, accession text, and ambiguous unlabeled fields are not sequences.\n- Missing or ambiguous sequences remain empty; they are never filled from another database.\n- The legacy category-wide normalized-text scanner does not contribute release records.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// writePlantAuditDocs deliberately emits one audit file per plant resource.
// These files are the review ledger for the plant-only rebuild; each resource
// has its own source file, format decision, and sequence coverage instead of
// inheriting a single category-wide claim.
func writePlantAuditDocs(manifestPath, root string, records []record) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff"))).ReadAll()
	if err != nil || len(rows) < 2 {
		return err
	}
	h := map[string]int{}
	for i, v := range rows[0] {
		h[strings.ToLower(strings.TrimSpace(v))] = i
	}
	get := func(row []string, key string) string {
		i := h[key]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	reviewedFiles, err := reviewedPlantSourceFiles(filepath.Join(filepath.Dir(manifestPath), "reviewed", "plants"))
	if err != nil {
		return err
	}
	for _, row := range rows[1:] {
		if get(row, "category") != "plants" {
			continue
		}
		label, local, source := get(row, "species_label"), get(row, "local_file"), get(row, "source_url")
		if label == "" {
			continue
		}
		// Resource-specific review commands own their completed audit files.
		// Never replace those decisions with the legacy category-wide summary.
		if reviewedFiles[local] {
			continue
		}
		count, seq := 0, 0
		methods := map[string]bool{}
		for _, r := range records {
			if r.SourceURL != source {
				continue
			}
			count++
			if strings.TrimSpace(r.Sequence) != "" {
				seq++
			}
			if strings.Contains(r.Description, "FASTA") {
				methods["FASTA block"] = true
			}
			if strings.Contains(r.Description, "sequence column") {
				methods["explicit sequence column"] = true
			}
		}
		method := "no accepted sequence; inspect source manually"
		if len(methods) > 0 {
			names := make([]string, 0, len(methods))
			for m := range methods {
				names = append(names, m)
			}
			sort.Strings(names)
			method = strings.Join(names, "; ")
		}
		sequenceHeaders, sequenceRows := inspectSequenceColumns(filepath.Join(filepath.Dir(filepath.Dir(manifestPath)), "extracted-v4", strings.TrimSuffix(local, filepath.Ext(local))+".txt"))
		ext := strings.ToLower(filepath.Ext(local))
		format := "Office-normalized text"
		if ext == ".xlsx" {
			format = "Excel workbook normalized to text"
		} else if ext == ".doc" {
			format = "Word document normalized to text"
		}
		slug := regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(strings.ToLower("plant-"+label), "-")
		body := fmt.Sprintf("# Plant resource audit: %s\n\n- Source file: `%s`\n- URL: %s\n- Detected container: `%s`\n- Resource-local records: `%d`\n- Records with accepted sequence: `%d`\n- Extraction method used for this resource: `%s`\n- Detected sequence-column header(s): `%s`\n- Rows with a non-empty sequence-column value: `%d`\n\n## Review rule\n\nThis resource is reviewed independently. FASTA extraction requires a CYP-bearing `>` header and sequence lines bounded by the next CYP header, a terminal `*`, or the first non-sequence annotation. Spreadsheet data is accepted only from an explicit `sequence` column. Alignment text, coordinates, descriptions, and unlabeled short fragments are rejected.\n\nMissing sequences remain empty until this exact source file is manually verified; no other database is used as a substitute.\n", label, local, source, format, count, seq, method, strings.Join(sequenceHeaders, "; "), sequenceRows)
		if err := os.WriteFile(filepath.Join(root, strings.Trim(slug, "-")+".md"), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func reviewedPlantSourceFiles(root string) (map[string]bool, error) {
	out := map[string]bool{}
	matches, err := filepath.Glob(filepath.Join(root, "*.csv"))
	if err != nil {
		return nil, err
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff"))).ReadAll()
		if err != nil || len(rows) < 2 {
			continue
		}
		column := -1
		for i, value := range rows[0] {
			if strings.EqualFold(strings.TrimSpace(value), "source_file") {
				column = i
				break
			}
		}
		if column < 0 {
			continue
		}
		for _, row := range rows[1:] {
			if column < len(row) && strings.TrimSpace(row[column]) != "" {
				out[strings.TrimSpace(row[column])] = true
			}
		}
	}
	return out, nil
}

// writePlantRecordAudit is intentionally verbose. It is the durable review
// ledger used to inspect each accepted or rejected plant record without
// relying on a category-wide aggregate number.
func writePlantRecordAudit(manifestPath, path string, records []record) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff"))).ReadAll()
	if err != nil {
		return err
	}
	h := map[string]int{}
	for i, v := range rows[0] {
		h[strings.ToLower(strings.TrimSpace(v))] = i
	}
	get := func(row []string, key string) string {
		i := h[key]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	profiles, err := readPlantResourceProfiles(filepath.Join(filepath.Dir(manifestPath), "plant_resource_profiles.csv"))
	if err != nil {
		return err
	}
	profileRows, _ := os.ReadFile(filepath.Join(filepath.Dir(manifestPath), "plant_resource_profiles.csv"))
	profileHeader, _ := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(profileRows), "\ufeff"))).ReadAll()
	profileMap := map[string][]string{}
	if len(profileHeader) > 0 {
		ph := map[string]int{}
		for i, v := range profileHeader[0] {
			ph[strings.ToLower(strings.TrimSpace(v))] = i
		}
		for _, row := range profileHeader[1:] {
			if i, ok := ph["local_file"]; ok && i < len(row) {
				profileMap[strings.TrimSpace(row[i])] = row
			}
		}
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	w := csv.NewWriter(out)
	_ = w.Write([]string{"species", "symbol", "id", "source_file", "source_url", "sequence_status", "sequence_length", "sequence_method", "terminal_marker", "fragment_or_pseudogene_hint", "est_hint", "foreign_species_hint", "missing_reason", "paired_resource_checked"})
	for _, row := range rows[1:] {
		if get(row, "category") != "plants" || !profiles[get(row, "local_file")] {
			continue
		}
		local, source, species := get(row, "local_file"), get(row, "source_url"), get(row, "species_label")
		for _, r := range records {
			if r.Category != "plants" || r.SourceURL != source {
				continue
			}
			seq := strings.TrimSpace(r.Sequence)
			status, reason := "missing", "No unambiguous sequence accepted by this resource's explicit profile"
			if seq != "" {
				status, reason = "accepted", ""
			}
			statuses := map[string]bool{}
			for _, statusValue := range strings.Split(strings.ToLower(r.ReviewStatus), ";") {
				statuses[strings.TrimSpace(statusValue)] = true
			}
			terminal := strings.HasSuffix(seq, "*") || statuses["terminal-marker"]
			fragment := statuses["fragment"] || statuses["partial"] || statuses["pseudogene"] || statuses["source-pseudogene-label"]
			est := statuses["est"]
			foreign := statuses["other-species"] || statuses["multi-species"] || statuses["foreign-species"]
			desc := strings.ToLower(r.Description)
			method := "profile-defined"
			if strings.Contains(desc, "fasta") {
				method = "FASTA block"
			} else if strings.Contains(desc, "sequence column") {
				method = "explicit sequence column"
			}
			paired := "not applicable"
			if strings.Contains(local, "Lotus.P450s.Oct31.2012") || strings.Contains(local, "Lotus.P450.set") {
				paired = "paired lotus annotation/sequence resources reviewed"
			}
			if strings.Contains(local, "Prunus.persica") || strings.Contains(local, "Aquilegia") {
				paired = "same-species alternate resource retained separately"
			}
			_ = w.Write([]string{species, r.Symbol, r.ID, local, source, status, fmt.Sprintf("%d", len(seq)), method, fmt.Sprintf("%t", terminal), fmt.Sprintf("%t", fragment), fmt.Sprintf("%t", est), fmt.Sprintf("%t", foreign), reason, paired})
		}
	}
	w.Flush()
	return w.Error()
}

func inspectSequenceColumns(path string) ([]string, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0
	}
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n"), "\n")
	var headers []string
	rows := 0
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		isHeader := false
		for _, field := range fields {
			f := strings.ToLower(strings.TrimSpace(field))
			if f == "sequence" || strings.Contains(f, "protein sequence") {
				value := strings.TrimSpace(field)
				seen := false
				for _, prior := range headers {
					if prior == value {
						seen = true
						break
					}
				}
				if !seen {
					headers = append(headers, value)
				}
				isHeader = true
			}
		}
		if isHeader {
			continue
		}
		for _, field := range fields {
			clean := strings.ToUpper(strings.Map(func(r rune) rune {
				if strings.ContainsRune("ACDEFGHIKLMNPQRSTVWY*", r) {
					return r
				}
				return -1
			}, field))
			if len(clean) >= 100 && strings.Trim(clean, "ACDEFGHIKLMNPQRSTVWY*") == "" {
				rows++
				break
			}
		}
	}
	return headers, rows
}

func writeResourceAuditDocs(manifestPath, root string, records []record) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return nil
	}
	h := map[string]int{}
	for i, v := range rows[0] {
		h[strings.ToLower(strings.TrimSpace(v))] = i
	}
	get := func(row []string, k string) string {
		i := h[k]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	for _, row := range rows[1:] {
		label, cat, local, source := get(row, "species_label"), get(row, "category"), get(row, "local_file"), get(row, "source_url")
		if label == "" {
			continue
		}
		slug := regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(strings.ToLower(cat+"-"+label), "-")
		slug = strings.Trim(slug, "-")
		recordCount, sequenceCount := 0, 0
		for _, r := range records {
			if r.SourceURL == source {
				recordCount++
				if strings.TrimSpace(r.Sequence) != "" {
					sequenceCount++
				}
			}
		}
		parserStatus := "resource-specific normalized-text audit"
		rule := "Parsing rule: FASTA blocks are recognized from `>` headers and wrapped lines; structured tables may provide sequence only from an explicit `sequence` column. Unlabeled tokens, alignment fragments, scores, and comments are not accepted as sequences. This record documents the exact resource-level extraction result; unresolved or ambiguous records remain searchable with an empty sequence."
		if local == "plants-rice_annotation.P450.xls" {
			parserStatus = "resource-specific BIFF8 workbook audit"
			rule = "Workbook rule: Sheet1 header row 6 and rows 7-418 are parsed by cmd/plant-rice-annotation; Sheet2 and Sheet3 are empty. The workbook has no literal protein sequence column, so all 412 accepted rows retain an empty sequence. SHA-256: 3532c8470d07188dc00d5135caa7772af80429d9b5ea471a2bb3ac25214a91c1."
		}
		body := fmt.Sprintf("# Resource audit: %s\n\n- Category: `%s`\n- Resource: `%s`\n- URL: %s\n- Parser status: `%s`\n- PGD records: `%d`\n- Records with verified sequence: `%d`\n\n%s\n", label, cat, local, source, parserStatus, recordCount, sequenceCount, rule)
		if err := os.WriteFile(filepath.Join(root, slug+".md"), []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func readManifestSpecies(path string) ([]speciesRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
	rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}
	h := map[string]int{}
	for i, v := range rows[0] {
		h[strings.ToLower(strings.TrimSpace(v))] = i
	}
	get := func(row []string, k string) string {
		i := h[k]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	seen := map[string]bool{}
	out := []speciesRecord{}
	for _, row := range rows[1:] {
		name, cat := get(row, "species_label"), get(row, "category")
		if name == "" {
			continue
		}
		key := cat + "|" + name
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, speciesRecord{Name: name, Category: cat, Selectable: false, Description: "Listed by Dr Nelson resource; record requires resource-specific review."})
	}
	return out, nil
}

// parseExtracted parses only normalized Office text produced by
// scripts/extract-office.ps1. It deliberately never reads the original DOC/XLSX
// bytes or scans printable binary strings. Each record keeps the resource URL
// and the manifest label so the resulting PGD is auditable.
func parseExtracted(manifestPath, extractedDir string) ([]record, error) {
	profiles, err := readPlantResourceProfiles(filepath.Join(filepath.Dir(manifestPath), "plant_resource_profiles.csv"))
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
	r := csv.NewReader(strings.NewReader(string(data)))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, nil
	}
	h := map[string]int{}
	for i, v := range rows[0] {
		h[strings.ToLower(strings.TrimSpace(v))] = i
	}
	get := func(row []string, key string) string {
		i := h[key]
		if i < 0 || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	var out []record
	for _, row := range rows[1:] {
		label, category, source, local := get(row, "species_label"), get(row, "category"), get(row, "source_url"), get(row, "local_file")
		if local == "" {
			continue
		}
		if category == "plants" && !profiles[local] {
			return nil, fmt.Errorf("plant resource %s has no explicit reviewed profile", local)
		}
		data, e := os.ReadFile(filepath.Join(extractedDir, strings.TrimSuffix(local, filepath.Ext(local))+".txt"))
		if e != nil {
			continue
		}
		seen := map[string]bool{}
		// Word/COM extraction can emit bare CR separators (not only CRLF).
		// Normalize both forms before reading wrapped FASTA blocks.
		normalized := strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")
		lines := strings.Split(normalized, "\n")
		// Office text preserves many Dr. Nelson Word FASTA resources as a
		// header line followed by wrapped sequence lines.  Parse those blocks
		// first; the old one-line token scan cannot recover wrapped FASTA.
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if !strings.HasPrefix(line, ">") {
				continue
			}
			matches := cypRE.FindStringSubmatch(line)
			if len(matches) == 0 {
				continue
			}
			name := strings.ToUpper(matches[0])
			seqParts := make([]string, 0, 8)
			started := false
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, ">") && cypRE.MatchString(next) {
					break
				}
				if strings.HasPrefix(next, ">") {
					continue
				}
				rawSequence := strings.Join(strings.Fields(next), "")
				for _, ch := range rawSequence {
					if !strings.ContainsRune("ACDEFGHIKLMNPQRSTVWY*", rune(ch)) {
						if started {
							break
						}
						rawSequence = ""
						break
					}
				}
				if rawSequence == "" && started {
					break
				}
				if rawSequence == "" {
					continue
				}
				clean := strings.ToUpper(strings.Map(func(r rune) rune {
					if (r >= 'A' && r <= 'Z') || r == '*' {
						return r
					}
					return -1
				}, next))
				valid := len(clean) >= 2 && strings.Trim(clean, "ACDEFGHIKLMNPQRSTVWY*") == ""
				if valid {
					seqParts = append(seqParts, clean)
					started = true
				} else if started {
					// FASTA sequence blocks end at the first non-sequence
					// annotation. Do not skip over BLAST/coordinate text.
					break
				}
				// Dr. Nelson Word resources commonly place BLAST alignment
				// text immediately after the terminal protein marker.  Never
				// consume that text as part of the FASTA sequence.
				if strings.Contains(next, "*") {
					break
				}
			}
			seq := strings.TrimSuffix(strings.Join(seqParts, ""), "*")
			// A normal CYP protein is several hundred residues.  Short
			// sequences are retained only when the Dr. Nelson header explicitly
			// identifies a partial/fragment/EST record; otherwise they are
			// almost always a column/line extraction artifact.
			headerLower := strings.ToLower(line)
			shortFragment := strings.Contains(headerLower, "partial") || strings.Contains(headerLower, "fragment") || strings.Contains(headerLower, "est")
			if ((len(seq) >= 100) || (shortFragment && len(seq) >= 30)) && !seen[name] {
				seen[name] = true
				out = append(out, record{ID: name, Category: category, Species: label, Symbol: name, Sequence: seq, Description: "Parsed FASTA from normalized resource text; source file: " + local, SourceURL: source})
			}
		}
		// Spreadsheet exports are not FASTA: only a column explicitly named
		// "sequence" may supply a sequence.  This prevents alignment scores,
		// comments, and arbitrary long tokens from being mistaken for FASTA.
		tableSequenceColumn := -1
		for _, line := range lines {
			lineSpecies := label
			if m := bracketSpeciesRE.FindStringSubmatch(line); len(m) == 2 {
				candidate := strings.TrimSpace(strings.ReplaceAll(m[1], "_", " "))
				if speciesShapeRE.MatchString(candidate) && !strings.ContainsAny(candidate[:minInt(len(candidate), 80)], "0123456789") && !strings.Contains(strings.ToLower(candidate), "predicted") && !strings.Contains(strings.ToLower(candidate), "cytochrome") {
					lineSpecies = candidate
				}
			}
			fields := strings.Split(line, "\t")
			lowerFields := make([]string, len(fields))
			for i := range fields {
				lowerFields[i] = strings.ToLower(strings.TrimSpace(fields[i]))
			}
			for i, f := range lowerFields {
				if f == "sequence" || strings.Contains(f, "protein sequence") {
					tableSequenceColumn = i
				}
			}
			for _, name := range cypRE.FindAllString(line, -1) {
				name = strings.ToUpper(name)
				if seen[name] {
					continue
				}
				seen[name] = true
				seq := ""
				if tableSequenceColumn >= 0 && tableSequenceColumn < len(fields) {
					candidate := strings.ToUpper(strings.Map(func(r rune) rune {
						if (r >= 'A' && r <= 'Z') || r == '*' {
							return r
						}
						return -1
					}, fields[tableSequenceColumn]))
					if len(candidate) >= 100 && strings.Trim(candidate, "ACDEFGHIKLMNPQRSTVWY*") == "" {
						seq = candidate
					}
				}
				out = append(out, record{ID: name, Category: category, Species: lineSpecies, Symbol: name, Sequence: seq, Description: "Parsed structured resource text; sequence only from explicit sequence column: " + local, SourceURL: source})
			}
		}
	}
	return out, nil
}

func readPlantResourceProfiles(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\ufeff"))).ReadAll()
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	if len(rows) == 0 {
		return out, nil
	}
	header := map[string]int{}
	for i, value := range rows[0] {
		header[strings.ToLower(strings.TrimSpace(value))] = i
	}
	fileIndex, fileOK := header["local_file"]
	statusIndex, statusOK := header["review_status"]
	if !fileOK || !statusOK {
		return nil, errors.New("invalid plant resource profile header")
	}
	for _, row := range rows[1:] {
		if fileIndex >= len(row) || statusIndex >= len(row) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(row[statusIndex]), "reviewed") {
			out[strings.TrimSpace(row[fileIndex])] = true
		}
	}
	return out, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func aggregateResource(category, label string) bool {
	l := strings.ToLower(label)
	if category == "bacteria" {
		return true
	}
	for _, token := range []string{"public", "all named", "lepidoptera", "environmental", "taxonomic group", "partial collection"} {
		if strings.Contains(l, token) {
			return true
		}
	}
	return false
}

func readReviewedSources(root string) ([]record, error) {
	var matches []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	var out []record
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		data = []byte(strings.TrimPrefix(string(data), "\ufeff"))
		rows, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if len(rows) < 2 {
			continue
		}
		header := map[string]int{}
		for i, v := range rows[0] {
			header[strings.ToLower(strings.TrimSpace(v))] = i
		}
		required := []string{"category", "species", "symbol", "source_url", "source_note"}
		valid := true
		for _, name := range required {
			if _, ok := header[name]; !ok {
				valid = false
				break
			}
		}
		if !valid {
			continue // non-record audit CSVs share this directory
		}
		for _, row := range rows[1:] {
			get := func(name string) string {
				i, ok := header[name]
				if !ok {
					return ""
				}
				if i >= len(row) {
					return ""
				}
				return strings.TrimSpace(row[i])
			}
			symbol := get("symbol")
			species := get("species")
			id := first(get("id"), symbol)
			if id == "" || species == "" {
				continue
			}
			out = append(out, record{ID: id, RecordKey: get("record_key"), Category: get("category"), Species: species, Symbol: symbol, Description: get("source_note"), Sequence: get("sequence"), SourceURL: get("source_url"), ReviewStatus: get("review_status")})
		}
	}
	return out, nil
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func mergeRecords(groups ...[]record) []record {
	seen := map[string]int{}
	var out []record
	for _, group := range groups {
		for _, r := range group {
			k := strings.ToLower(strings.Join([]string{r.Category, r.Species, r.Symbol, r.ID, r.RecordKey, r.SourceURL}, "|"))
			if i, ok := seen[k]; ok {
				if out[i].Description == "" {
					out[i].Description = r.Description
				}
				if out[i].Sequence == "" {
					out[i].Sequence = r.Sequence
				}
				continue
			}
			seen[k] = len(out)
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Species+"|"+out[i].Symbol) < strings.ToLower(out[j].Species+"|"+out[j].Symbol)
	})
	return out
}
func writeSpeciesDocs(root string, records []record) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	old, err := filepath.Glob(filepath.Join(root, "*.md"))
	if err != nil {
		return err
	}
	for _, path := range old {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	bySpecies := map[string][]record{}
	for _, r := range records {
		if strings.TrimSpace(r.Species) != "" {
			bySpecies[r.Species] = append(bySpecies[r.Species], r)
		}
	}
	for species, rows := range bySpecies {
		name := regexp.MustCompile(`[^A-Za-z0-9._-]+`).ReplaceAllString(strings.ToLower(species), "-")
		name = strings.Trim(name, "-")
		if name == "" {
			continue
		}
		var b strings.Builder
		withSequence, short, long := 0, 0, 0
		methods := map[string]bool{}
		for _, r := range rows {
			if strings.TrimSpace(r.Sequence) == "" {
				continue
			}
			withSequence++
			if len(strings.TrimSpace(r.Sequence)) < 100 {
				short++
			}
			if len(strings.TrimSpace(r.Sequence)) > 1000 {
				long++
			}
			if strings.Contains(r.Description, "FASTA") {
				methods["FASTA block"] = true
			} else if strings.Contains(r.Description, "sequence column") {
				methods["explicit sequence column"] = true
			} else {
				methods["reviewed structured record"] = true
			}
		}
		methodNames := make([]string, 0, len(methods))
		for method := range methods {
			methodNames = append(methodNames, method)
		}
		sort.Strings(methodNames)
		fmt.Fprintf(&b, "# %s\n\nCategory: `%s`\n\nRecords: %d\n\nSequence audit: `%d/%d` records have an accepted sequence; short records (<100 aa): `%d`; unusually long records (>1000 aa): `%d`.\n\nExtraction method(s): `%s`.\n\nA missing sequence means the current Dr. Nelson resource did not provide an unambiguous protein sequence for this exact record. No external database sequence is substituted.\n\n| CYP / ID | Sequence | Description | Source |\n|---|---|---|---|\n", species, rows[0].Category, len(rows), withSequence, len(rows), short, long, strings.Join(methodNames, "; "))
		for _, r := range rows {
			status := "missing"
			if strings.TrimSpace(r.Sequence) != "" {
				status = fmt.Sprintf("present (%d aa)", len(strings.TrimSpace(r.Sequence)))
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", md(r.Symbol), status, md(r.Description), md(r.SourceURL))
		}
		if err := os.WriteFile(filepath.Join(root, name+".md"), []byte(b.String()), 0o644); err != nil {
			return err
		}
	}
	return nil
}
func md(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(s), "|", "\\|"), "\n", " ")
}

func collect(ctx context.Context) ([]record, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	linkRE := regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	stripRE := regexp.MustCompile(`(?is)<[^>]+>`)
	seen := map[string]bool{}
	var links []record
	for category, page := range pages {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, page, nil)
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("%s: %s", page, resp.Status)
		}
		base, _ := url.Parse(page)
		for _, match := range linkRE.FindAllSubmatch(body, -1) {
			href := strings.TrimSpace(string(match[1]))
			label := strings.Join(strings.Fields(stripRE.ReplaceAllString(string(match[2]), " ")), " ")
			if href == "" || label == "" {
				continue
			}
			ref, err := base.Parse(href)
			if err != nil || ref.Host == "" {
				continue
			}
			key := category + "|" + ref.String() + "|" + label
			if seen[key] {
				continue
			}
			seen[key] = true
			id := filepath.Base(strings.TrimSuffix(ref.Path, "/"))
			if id == "." || id == "" {
				id = label
			}
			if ext := strings.ToLower(filepath.Ext(ref.Path)); ext == ".doc" || ext == ".xlsx" || ext == ".txt" || ext == ".fasta" || ext == ".fa" {
				links = append(links, record{ID: id, Category: category, Species: label, Symbol: label, Description: label, SourceURL: ref.String()})
			}
		}
	}
	records := fetchResources(ctx, client, links)
	sort.Slice(records, func(i, j int) bool {
		if records[i].Category != records[j].Category {
			return records[i].Category < records[j].Category
		}
		return records[i].ID < records[j].ID
	})
	return records, nil
}

var cypRE = regexp.MustCompile(`(?i)\bCYP[0-9]{1,4}[A-Z]{1,4}[0-9]{1,4}(?:\.[0-9]+)?\b`)

func fetchResources(ctx context.Context, client *http.Client, links []record) []record {
	workers := 12
	if len(links) < workers {
		workers = len(links)
	}
	jobs := make(chan record)
	out := make(chan []record, len(links))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range jobs {
				out <- fetchResource(ctx, client, link)
			}
		}()
	}
	go func() {
		for _, l := range links {
			jobs <- l
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()
	var records []record
	for batch := range out {
		records = append(records, batch...)
	}
	return records
}
func fetchResource(ctx context.Context, client *http.Client, link record) []record {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, link.SourceURL, nil)
	resp, err := client.Do(req)
	if err != nil {
		return []record{link}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return []record{link}
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return []record{link}
	}
	text := extractPrintable(data)
	names := cypRE.FindAllString(text, -1)
	if len(names) == 0 {
		return []record{link}
	}
	seen := map[string]bool{}
	out := make([]record, 0, len(names))
	for _, name := range names {
		key := strings.ToUpper(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, record{ID: key, Category: link.Category, Species: link.Species, Symbol: key, Description: link.Description, SourceURL: link.SourceURL})
	}
	return out
}
func extractPrintable(data []byte) string {
	var b strings.Builder
	b.Grow(len(data))
	space := false
	for _, c := range data {
		if c >= 32 && c <= 126 {
			b.WriteByte(c)
			space = false
		} else if !space {
			b.WriteByte(' ')
			space = true
		}
	}
	return b.String()
}

func write(path string, records []record, manifestSpecies []speciesRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	tmp := path + ".tmp"
	compact := path + ".compact"
	_ = os.Remove(tmp)
	_ = os.Remove(compact)
	db, err := bolt.Open(tmp, 0o644, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucket([]byte("records"))
		if e != nil {
			return e
		}
		for i, r := range records {
			v, e := json.Marshal(r)
			if e != nil {
				return e
			}
			if e = b.Put([]byte(fmt.Sprintf("%08d", i)), v); e != nil {
				return e
			}
		}
		speciesBucket, e := tx.CreateBucket([]byte("species"))
		if e != nil {
			return e
		}
		seen := map[string]bool{}
		index := 0
		for _, r := range records {
			key := r.Category + "|" + r.Species
			if strings.TrimSpace(r.Species) == "" || seen[key] {
				continue
			}
			seen[key] = true
			// A species backed by at least one reviewed record is selectable in
			// every supported category. Manifest-only resources remain disabled
			// until a resource-specific parser contributes reviewed records.
			s := speciesRecord{Name: r.Species, Category: r.Category, Selectable: true, Description: r.Description}
			v, e := json.Marshal(s)
			if e != nil {
				return e
			}
			if e = speciesBucket.Put([]byte(fmt.Sprintf("%08d", index)), v); e != nil {
				return e
			}
			index++
		}
		for i := range manifestSpecies {
			for _, r := range records {
				if strings.EqualFold(r.Category, manifestSpecies[i].Category) && strings.EqualFold(r.Species, manifestSpecies[i].Name) {
					manifestSpecies[i].Selectable = true
					manifestSpecies[i].Description = "Enabled by a completed resource-specific reviewed record set."
					break
				}
			}
		}
		for _, s := range manifestSpecies {
			key := s.Category + "|" + s.Name
			if seen[key] {
				continue
			}
			seen[key] = true
			v, e := json.Marshal(s)
			if e != nil {
				return e
			}
			if e = speciesBucket.Put([]byte(fmt.Sprintf("%08d", index)), v); e != nil {
				return e
			}
			index++
		}
		return nil
	})
	closeErr := db.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	// Bolt grows geometrically while inserting tens of thousands of records.
	// Compacting the completed temporary database keeps the published download
	// proportional to live data and avoids shipping unused preallocated pages.
	src, err := bolt.Open(tmp, 0o444, &bolt.Options{ReadOnly: true})
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	dst, err := bolt.Open(compact, 0o644, nil)
	if err != nil {
		_ = src.Close()
		_ = os.Remove(tmp)
		return err
	}
	err = bolt.Compact(dst, src, 64<<20)
	closeDst, closeSrc := dst.Close(), src.Close()
	if err == nil {
		err = closeDst
	}
	if err == nil {
		err = closeSrc
	}
	_ = os.Remove(tmp)
	if err != nil {
		_ = os.Remove(compact)
		return err
	}
	return os.Rename(compact, path)
}
