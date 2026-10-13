package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const (
	workbookFile = "plants-Lotus.P450s.Oct31.2012.xlsx"
	sequenceFile = "plants-Lotus.P450.set.doc"
	workbookURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Lotus.P450s.Oct31.2012.xlsx"
	sequenceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Lotus.P450.set.doc"
	workbookHash = "7a79841a8c66473596bf5ea1c2d7e3719304eda72e5408ebdafa3b51794434ba"
	sequenceHash = "99fbd1983129e64c2dd6e21b04197fe6c9dc22a43cb6cbdb95631394324994c0"
	nameSheet    = "CYPs in name order"
)

type annotation struct {
	Row                                                                                                           int
	ShortName, Model, Symbol, BestHit, PercentID, Note, Scaffold, Begin, End, Strand, Gene, Pseudogene, MergeNote string
}
type record struct {
	Block, SourceLine, FirstSequenceLine, LastSequenceLine int
	Header, ID, Symbol, RecordKey, Sequence                string
	Status, Annotations                                    []string
}
type excludedRecord struct {
	Block, SourceLine int
	Header, Reason    string
}

var (
	leadingCoordinates  = regexp.MustCompile(`^(?:\d+\s+){1,2}`)
	trailingCoordinates = regexp.MustCompile(`\s+\d+(?:\s+\d+)?\s*$`)
	phaseMarker         = regexp.MustCompile(`\((?:0|1|2|\?|0\?|1\?|2\?)?\)`)
	whitespace          = regexp.MustCompile(`\s+`)
	headerCYP           = regexp.MustCompile(`\bCYP[0-9A-Za-z]+(?:[-.][0-9A-Za-z]+)*`)
	pseudogeneSymbol    = regexp.MustCompile(`(?i)^CYP[0-9A-Z]*P(?:V[0-9]+|X|[-.]|$)`)
	modelID             = regexp.MustCompile(`\b(?:maker|snap|augustus)(?:_masked)?-scaffold_[^ ]+`)
	nonKey              = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
)

func main() {
	workbook := flag.String("workbook", filepath.Join("raw", workbookFile), "reviewed Nelumbo annotation workbook")
	input := flag.String("input", filepath.Join("raw", "plants-Lotus.P450.set.txt"), "Word-normalized Nelumbo sequence source")
	sourceDoc := flag.String("source-doc", filepath.Join("raw", sequenceFile), "original Nelumbo Word source")
	out := flag.String("out", filepath.Join("sources", "reviewed", "plants", "nelumbo-nucifera.csv"), "reviewed CSV")
	audit := flag.String("audit", filepath.Join("docs", "plants", "plant-nelumbo-nucifera.md"), "review ledger")
	flag.Parse()
	annotations, wb, err := reviewWorkbook(*workbook)
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		panic(err)
	}
	records, excluded, err := parseNelumboText(string(data))
	if err != nil {
		panic(err)
	}
	if len(records) != 364 || countSequences(records) != 364 || len(excluded) != 2 {
		panic(fmt.Errorf("Nelumbo counts changed: records=%d sequences=%d excluded=%d", len(records), countSequences(records), len(excluded)))
	}
	wh, err := fileSHA256(*workbook)
	if err != nil {
		panic(err)
	}
	if wh != workbookHash {
		panic(fmt.Errorf("Nelumbo workbook hash changed: %s", wh))
	}
	sh, err := fileSHA256(*sourceDoc)
	if err != nil {
		panic(err)
	}
	if sh != sequenceHash {
		panic(fmt.Errorf("Nelumbo sequence source hash changed: %s", sh))
	}
	if err := writeCSV(*out, records, wh, sh); err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, records, excluded, annotations, wb, wh, sh); err != nil {
		panic(err)
	}
	fmt.Printf("Nelumbo nucifera: %d literal sequence blocks accepted; %d temporary Aquilegia blocks excluded; workbook has %d model rows resolving to 355 curated sequences\n", len(records), len(excluded), len(annotations))
}

func reviewWorkbook(path string) ([]annotation, *plantxlsx.Workbook, error) {
	names, err := plantxlsx.SheetNames(path)
	if err != nil {
		return nil, nil, err
	}
	want := []string{"CYPs in name order", "CYPs in scaffold order", "Gene Pairs by WGD"}
	if strings.Join(names, "|") != strings.Join(want, "|") {
		return nil, nil, fmt.Errorf("Nelumbo sheet list changed: %v", names)
	}
	for _, sheet := range want[1:] {
		w, err := plantxlsx.ReadSheet(path, sheet, false)
		if err != nil {
			return nil, nil, err
		}
		expected := map[string]string{"CYPs in scaffold order": "A1:M376", "Gene Pairs by WGD": "A1:I41"}[sheet]
		if w.Dimension != expected || len(w.HiddenRows) > 0 || len(w.HiddenColumns) > 0 || len(w.MergedCells) > 0 || w.TableParts != 0 {
			return nil, nil, fmt.Errorf("Nelumbo sheet %q structure changed", sheet)
		}
	}
	wb, err := plantxlsx.ReadSheet(path, nameSheet, false)
	if err != nil {
		return nil, nil, err
	}
	if wb.Dimension != "A1:M384" || len(wb.HiddenRows) > 0 || len(wb.HiddenColumns) > 0 || len(wb.MergedCells) > 0 || wb.TableParts != 0 {
		return nil, nil, fmt.Errorf("Nelumbo name sheet structure changed")
	}
	h := wb.Rows[1]
	if h["A"] != "short name" || h["B"] != "contig and gene ID" || h["C"] != "CYP name" || h["D"] != "Best match" || h["E"] != "% ID" || h["G"] != "scaffold" || h["H"] != "begin" || h["I"] != "end" || h["J"] != "strand" || h["K"] != "gene" || h["L"] != "pseudogene" {
		return nil, nil, fmt.Errorf("Nelumbo headers changed")
	}
	rows := make([]annotation, 0, 372)
	genes, pseudos, continuations := 0, 0, 0
	for row := 2; row <= 373; row++ {
		v := wb.Rows[row]
		if v == nil || strings.TrimSpace(v["C"]) == "" {
			return nil, nil, fmt.Errorf("Nelumbo data row %d missing", row)
		}
		a := annotation{row, strings.TrimSpace(v["A"]), strings.TrimSpace(v["B"]), "CYP" + strings.TrimSpace(v["C"]), strings.TrimSpace(v["D"]), strings.TrimSpace(v["E"]), strings.TrimSpace(v["F"]), strings.TrimSpace(v["G"]), strings.TrimSpace(v["H"]), strings.TrimSpace(v["I"]), strings.TrimSpace(v["J"]), strings.TrimSpace(v["K"]), strings.TrimSpace(v["L"]), strings.TrimSpace(v["M"])}
		if a.Gene == "1" {
			genes++
		}
		if a.Pseudogene == "1" {
			pseudos++
		}
		if a.Pseudogene == "0" {
			continuations++
		}
		rows = append(rows, a)
	}
	if len(rows) != 372 || genes != 172 || pseudos != 179 || continuations != 17 {
		return nil, nil, fmt.Errorf("Nelumbo workbook totals changed rows=%d genes=%d pseudos=%d continuation=%d", len(rows), genes, pseudos, continuations)
	}
	if wb.Rows[376]["B"] != "355 sequences, 175 genes, 180 pseudogenes." || wb.Rows[380]["B"] != "32 fragments join into 15 pseudogenes" {
		return nil, nil, fmt.Errorf("Nelumbo workbook summary changed")
	}
	return rows, wb, nil
}

func parseNelumboText(text string) ([]record, []excludedRecord, error) {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	var headers []int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), ">") {
			headers = append(headers, i)
		}
	}
	if len(headers) != 366 {
		return nil, nil, fmt.Errorf("Nelumbo header count changed: %d", len(headers))
	}
	var records []record
	var excluded []excludedRecord
	for bi, start := range headers {
		block := bi + 1
		end := len(lines)
		if bi+1 < len(headers) {
			end = headers[bi+1]
		}
		header := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[start]), ">"))
		if block == 276 || block == 277 {
			excluded = append(excluded, excludedRecord{block, start + 1, header, "source preface explicitly says two temporary Aquilegia sequences are included"})
			continue
		}
		r := record{Block: block, SourceLine: start + 1, Header: header}
		r.Symbol = headerCYP.FindString(header)
		if r.Symbol == "" {
			return nil, nil, fmt.Errorf("Nelumbo block %d missing CYP name", block)
		}
		r.ID = modelID.FindString(header)
		if r.ID == "" {
			r.ID = r.Symbol
		}
		r.RecordKey = fmt.Sprintf("nelumbo-nucifera:block-%04d:%s", block, keyPart(r.ID))
		var parts []string
		for i := start + 1; i < end; i++ {
			seq, ok := nelumboSequenceLine(lines[i])
			if ok {
				if r.FirstSequenceLine == 0 {
					r.FirstSequenceLine = i + 1
				}
				r.LastSequenceLine = i + 1
				parts = append(parts, seq)
				continue
			}
			if v := strings.TrimSpace(lines[i]); v != "" && len(r.Annotations) < 8 {
				r.Annotations = append(r.Annotations, v)
			}
		}
		raw := strings.Join(parts, "")
		if strings.HasSuffix(raw, "*") {
			raw = strings.TrimSuffix(raw, "*")
			r.Status = append(r.Status, "source-terminal-stop")
		}
		r.Sequence = raw
		if r.Sequence == "" {
			return nil, nil, fmt.Errorf("Nelumbo block %d (%s) has no literal sequence", block, header)
		}
		lower := strings.ToLower(strings.Join(lines[start:end], "\n"))
		flags := map[string]bool{}
		if pseudogeneSymbol.MatchString(r.Symbol) || strings.Contains(lower, "pseudogene") {
			flags["source-pseudogene"] = true
		}
		for _, p := range []string{"fragment", "partial", "missing", "runs off", "n-term", "c-term", "one exon"} {
			if strings.Contains(lower, p) {
				flags["source-fragment-or-missing-region"] = true
			}
		}
		if strings.Contains(lower, "join") {
			flags["source-joined-piece"] = true
		}
		if strings.Contains(lower, "gap") {
			flags["source-gap-annotation"] = true
		}
		if strings.Contains(r.Sequence, "*") {
			flags["internal-stop"] = true
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			flags["ambiguous-X-or-x"] = true
		}
		if strings.Contains(r.Sequence, "-") {
			flags["source-gap"] = true
		}
		if len(r.Sequence) < 350 {
			flags["short-sequence"] = true
		}
		for _, aa := range r.Sequence {
			if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXYx*-", aa) {
				return nil, nil, fmt.Errorf("Nelumbo block %d unexpected residue %q", block, aa)
			}
		}
		for _, name := range []string{"source-pseudogene", "source-fragment-or-missing-region", "source-joined-piece", "source-gap-annotation", "source-gap", "internal-stop", "ambiguous-X-or-x", "short-sequence"} {
			if flags[name] {
				r.Status = append(r.Status, name)
			}
		}
		records = append(records, r)
	}
	return records, excluded, nil
}

func nelumboSequenceLine(raw string) (string, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", false
	}
	for _, r := range v {
		if (r >= 'a' && r <= 'w') || r == 'y' || r == 'z' {
			return "", false
		}
	}
	v = leadingCoordinates.ReplaceAllString(v, "")
	v = trailingCoordinates.ReplaceAllString(v, "")
	v = phaseMarker.ReplaceAllString(v, "")
	v = strings.ReplaceAll(v, "&", "")
	v = whitespace.ReplaceAllString(v, "")
	if v == "" {
		return "", false
	}
	for _, aa := range v {
		if !strings.ContainsRune("ACDEFGHIKLMNOPQRSTVWXYx*-", aa) {
			return "", false
		}
	}
	return v, true
}

func writeCSV(path string, records []record, wh, sh string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "source_line", "source_header", "review_status"})
	for _, r := range records {
		note := fmt.Sprintf("Nelumbo Word block %d at normalized line %d; literal protein lines %d-%d; paired annotation workbook=%s sha256=%s; header=%s", r.Block, r.SourceLine, r.FirstSequenceLine, r.LastSequenceLine, workbookFile, wh, r.Header)
		if len(r.Annotations) > 0 {
			note += "; annotations: " + strings.Join(r.Annotations, " / ")
		}
		_ = w.Write([]string{"plants", "Nelumbo nucifera", r.Symbol, r.ID, r.RecordKey, sequenceURL, note, r.Sequence, sequenceFile, sh, strconv.Itoa(r.Block), strconv.Itoa(r.SourceLine), r.Header, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}

func writeAudit(path string, records []record, excluded []excludedRecord, annotations []annotation, wb *plantxlsx.Workbook, wh, sh string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	counts := map[string]int{}
	seqs := map[string]int{}
	for _, r := range records {
		seqs[r.Sequence]++
		for _, s := range r.Status {
			counts[s]++
		}
	}
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	dupes := 0
	for _, n := range seqs {
		if n > 1 {
			dupes++
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Plant resource review: Nelumbo nucifera\n\n- Annotation workbook: `%s`\n- Workbook URL: %s\n- Workbook SHA-256: `%s`\n- Sequence Word file: `%s`\n- Sequence URL: %s\n- Sequence SHA-256: `%s`\n- Workbook sheets: `CYPs in name order` (`A1:M384`), `CYPs in scaffold order` (`A1:M376`), `Gene Pairs by WGD` (`A1:I41`)\n- Workbook annotation/model rows: `%d` (172 gene rows, 179 primary pseudogene rows, 17 merged-model continuation rows)\n- Workbook curated-sequence summary: `355 sequences, 175 genes, 180 pseudogenes`\n- Word container inspected: legacy Word, 2,948 paragraphs, 77 pages, no tables\n- Word `>` blocks: `366`\n- Accepted Nelumbo literal sequence blocks: `%d`\n- Temporary Aquilegia blocks excluded: `%d`\n- Duplicate literal-sequence groups retained: `%d`\n- Review status: `complete`\n\n", workbookFile, workbookURL, wh, sequenceFile, sequenceURL, sh, len(annotations), len(records), len(excluded), dupes)
	b.WriteString("## Resource-specific interpretation\n\nThe workbook and Word file are two parts of one Nelumbo release. The workbook contains curated names, coordinates, gene/pseudogene calls, merged-model notes, scaffold ordering, and WGD pairs, but no amino-acid sequence column. It therefore acts as the annotation ledger and does not independently create empty-sequence PGD records. Its 372 model rows resolve, per its own footer, to 355 curated sequences because 17 rows are continuation models in merged pseudogenes.\n\nThe Word file supplies literal protein sequences. Its preface says that, beyond the 355 sequence set, nine redundant contig sequences and two temporary `Lotus japonicus` sequences are present. The only two explicit foreign blocks are instead labelled `Aquilegia` in their headers (blocks 276-277); this source inconsistency is preserved in the audit, and those two non-Nelumbo blocks are excluded. The remaining 364 Nelumbo blocks are retained in source order, including the nine redundant contig representations.\n\nProtein lines use several reviewed layouts: whole proteins with spaces, scaffold fragments with one or two leading coordinates, trailing coordinates, phase markers, source `&` joins, X/x, gaps and internal stops. Layout metadata is removed while literal uncertainty remains; only one terminal `*` is removed. No sequence is translated, repaired or externally completed.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	b.WriteString("\n## Representative records\n\n| Accepted index | Source block | Source line | ID | CYP symbol | Sequence | Status |\n|---:|---:|---:|---|---|---:|---|\n")
	for _, i := range []int{0, len(records) / 2, len(records) - 1} {
		r := records[i]
		fmt.Fprintf(&b, "| %d | %d | %d | %s | %s | %d aa | %s |\n", i+1, r.Block, r.SourceLine, md(r.ID), md(r.Symbol), len(r.Sequence), md(strings.Join(r.Status, ";")))
	}
	b.WriteString("\n## Excluded blocks\n\n| Source block | Source line | Header | Reason |\n|---:|---:|---|---|\n")
	for _, r := range excluded {
		fmt.Fprintf(&b, "| %d | %d | %s | %s |\n", r.Block, r.SourceLine, md(r.Header), md(r.Reason))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func countSequences(rs []record) int {
	n := 0
	for _, r := range rs {
		if r.Sequence != "" {
			n++
		}
	}
	return n
}
func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
func keyPart(v string) string {
	v = nonKey.ReplaceAllString(strings.TrimSpace(v), "-")
	return strings.ToLower(strings.Trim(v, "-"))
}
func md(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(v), "|", "\\|"), "\n", " ")
}
