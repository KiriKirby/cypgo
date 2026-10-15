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
	"unicode"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

const (
	tarsierSourceFile = "animals-Tarsier.P450s.xlsx"
	tarsierSourceURL  = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Tarsier.P450s.xlsx"
	tarsierSheet      = "combined, sorted by CYP name"
)

var tarsierKeyRE = regexp.MustCompile(`[^A-Za-z0-9._|-]+`)

type tarsierRecord struct {
	Row                                                                                                                      int
	GeneID, TranscriptID, Start, End, HumanHit, PercentID, GotohID, Symbol, Comment, SourceSequence, Sequence, ID, RecordKey string
	Statuses                                                                                                                 []string
}

func main() {
	in := flag.String("input", filepath.Join("raw", tarsierSourceFile), "")
	out := flag.String("out", filepath.Join("sources", "reviewed", "animals", "tarsier.csv"), "")
	audit := flag.String("audit", filepath.Join("docs", "animals", "tarsier.md"), "")
	flag.Parse()
	names, e := plantxlsx.SheetNames(*in)
	if e != nil {
		panic(e)
	}
	want := []string{"Tarsier.ensembl.txt", "Gotoh's set", tarsierSheet, "87 Genbank proteins"}
	if strings.Join(names, "\x00") != strings.Join(want, "\x00") {
		panic(fmt.Errorf("tarsier sheets changed: %q", names))
	}
	wb, e := plantxlsx.ReadSheet(*in, tarsierSheet, false)
	if e != nil {
		panic(e)
	}
	rs, e := reviewTarsierRows(wb.Rows)
	if e != nil {
		panic(e)
	}
	data, e := os.ReadFile(*in)
	if e != nil {
		panic(e)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	if e = writeTarsierCSV(*out, rs, hash); e != nil {
		panic(e)
	}
	if e = writeTarsierAudit(*audit, rs, wb, hash); e != nil {
		panic(e)
	}
	fmt.Printf("Tarsius syrichta: %d reviewed rows, %d literal protein sequences\n", len(rs), len(rs))
}

func reviewTarsierRows(rows map[int]map[string]string) ([]tarsierRecord, error) {
	h := rows[1]
	want := map[string]string{"A": "Gene ID", "B": "transcript ID", "C": "start", "D": "end", "E": "", "F": "best hit human", "G": "%ID", "H": "Gotoh ID", "I": "CYP name", "J": "comments", "K": "seq", "L": ""}
	for c, v := range want {
		if h[c] != v {
			return nil, fmt.Errorf("tarsier header %s got %q want %q", c, h[c], v)
		}
	}
	var out []tarsierRecord
	for row := 2; row <= 109; row++ {
		v := rows[row]
		r := tarsierRecord{Row: row, GeneID: strings.TrimSpace(v["A"]), TranscriptID: strings.TrimSpace(v["B"]), Start: strings.TrimSpace(v["C"]), End: strings.TrimSpace(v["D"]), HumanHit: strings.TrimSpace(v["F"]), PercentID: strings.TrimSpace(v["G"]), GotohID: strings.TrimSpace(v["H"]), Symbol: strings.TrimSpace(v["I"]), Comment: strings.TrimSpace(v["J"]), SourceSequence: strings.TrimSpace(v["K"])}
		if r.Symbol == "" || r.SourceSequence == "" {
			return nil, fmt.Errorf("tarsier row %d missing CYP/sequence", row)
		}
		r.Sequence, r.Statuses = normalizeTarsierSequence(r.SourceSequence)
		for _, x := range r.Sequence {
			u := unicode.ToUpper(x)
			if (u < 'A' || u > 'Z') && x != '*' {
				return nil, fmt.Errorf("tarsier row %d unexpected %q", row, x)
			}
		}
		low := strings.ToLower(r.Comment)
		if strings.Contains(r.Sequence, "*") {
			r.Statuses = append(r.Statuses, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Statuses = append(r.Statuses, "ambiguous-X-retained")
		}
		if len(r.Sequence) < 350 {
			r.Statuses = append(r.Statuses, "short-source-fragment")
		}
		if strings.Contains(low, "missing") || strings.Contains(low, "fragment") || strings.Contains(low, "partial") || strings.Contains(low, "exon") {
			r.Statuses = append(r.Statuses, "source-partial-or-missing-region")
		}
		r.ID = firstTarsier(r.GeneID, r.GotohID, r.TranscriptID, r.Symbol)
		r.RecordKey = fmt.Sprintf("tarsier:sheet-combined-sorted-by-cyp-name:row-%04d:%s", row, tarsierKey(r.ID))
		out = append(out, r)
	}
	if len(out) != 108 {
		return nil, fmt.Errorf("tarsier count=%d want 108", len(out))
	}
	return out, nil
}
func normalizeTarsierSequence(s string) (string, []string) {
	var b strings.Builder
	lo := false
	ws := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			ws = true
			continue
		}
		if unicode.IsLower(r) {
			lo = true
		}
		b.WriteRune(r)
	}
	s = b.String()
	var st []string
	if ws {
		st = append(st, "source-whitespace-joined")
	}
	if lo {
		st = append(st, "source-lowercase-retained")
	}
	if strings.HasSuffix(s, "*") {
		s = strings.TrimSuffix(s, "*")
		st = append(st, "terminal-stop-removed")
	}
	return s, st
}
func firstTarsier(vs ...string) string {
	for _, v := range vs {
		v = strings.TrimSpace(v)
		if v != "" && v != "no exact match" {
			return v
		}
	}
	return ""
}
func tarsierKey(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	return strings.Trim(tarsierKeyRE.ReplaceAllString(v, "-"), "-")
}
func writeTarsierCSV(path string, rs []tarsierRecord, hash string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "gene_id", "transcript_id", "gotoh_id", "start", "end", "human_hit", "percent_identity", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Tarsier combined curated sheet row %d; gene=%s; transcript=%s; source interval=%s-%s; human hit=%s (%s%% ID); Gotoh ID=%s; comment=%s", r.Row, r.GeneID, r.TranscriptID, r.Start, r.End, r.HumanHit, r.PercentID, r.GotohID, r.Comment)
		_ = w.Write([]string{"animals", "Tarsius syrichta", r.Symbol, r.ID, r.RecordKey, tarsierSourceURL, note, r.Sequence, tarsierSourceFile, hash, tarsierSheet, strconv.Itoa(r.Row), r.GeneID, r.TranscriptID, r.GotohID, r.Start, r.End, r.HumanHit, r.PercentID, strings.Join(r.Statuses, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeTarsierAudit(path string, rs []tarsierRecord, wb *plantxlsx.Workbook, hash string) error {
	if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
		return e
	}
	counts := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Statuses {
			counts[s]++
		}
	}
	names := make([]string, 0, len(counts))
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Tarsius syrichta (tarsier)\n\n- Source file: `%s`\n- URL: %s\n- Source SHA-256: `%s`\n- Workbook sheets: `Tarsier.ensembl.txt`, `Gotoh's set`, `%s`, `87 Genbank proteins`\n- Authoritative sheet: `%s`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Records with literal protein sequence: `%d`\n- Review status: `complete`\n\n", tarsierSourceFile, tarsierSourceURL, hash, tarsierSheet, tarsierSheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe workbook separates 46 Ensembl candidates, 31 Gotoh predictions, and 87 GenBank proteins from the manually combined naming table. Only `combined, sorted by CYP name` is published because it records the reviewed relationship among those sources. Its rows 2–109 contain 108 explicit CYP-name/sequence pairs. Repeated CYP names are source-distinct rows and remain separate.\n\nGene/GenBank accession is the preferred ID, followed by Gotoh ID, transcript, and CYP name. Literal long X runs, one lowercase `x`, internal stops, fragments, pseudogenes and missing-region sequences remain intact. Only terminal `*` is removed. Candidate sheets are provenance inputs, not extra records; no cross-sheet concatenation, repair, translation, or deduplication is performed.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range names {
		fmt.Fprintf(&b, "| %s | %d |\n", n, counts[n])
	}
	b.WriteString("\n## Representative rows\n\n| Source row | CYP name | ID | Sequence length | Notes |\n|---:|---|---|---:|---|\n")
	for _, i := range []int{0, 6, 41, len(rs) / 2, len(rs) - 1} {
		r := rs[i]
		fmt.Fprintf(&b, "| %d | %s | %s | %d aa | %s |\n", r.Row, mdTarsier(r.Symbol), mdTarsier(r.ID), len(r.Sequence), mdTarsier(strings.Join(r.Statuses, ";")))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
func mdTarsier(v string) string { return strings.ReplaceAll(strings.TrimSpace(v), "|", "\\|") }
