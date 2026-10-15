package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"github.com/KiriKirby/cypgo/internal/plantxlsx"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const file = "animals-POMPEII.worm.ESTs.xlsx"
const url = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/POMPEII.worm.ESTs.xlsx"
const sheet = "All clans unique"

var cypRE = regexp.MustCompile(`(?i)\bCYP[^\s,;]+`)
var keyRE = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

type rec struct {
	Row                                                                                         int
	GI, Accession, ProteinID, Hit, Identity, Comment, Genus, Species, Sequence, Symbol, ID, Key string
	Status                                                                                      []string
}

func main() {
	w, e := plantxlsx.ReadSheet(filepath.Join("raw", file), sheet, false)
	if e != nil {
		panic(e)
	}
	rs, e := review(w.Rows)
	if e != nil {
		panic(e)
	}
	b, e := os.ReadFile(filepath.Join("raw", file))
	if e != nil {
		panic(e)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if e = writeCSV(rs, h); e != nil {
		panic(e)
	}
	if e = writeAudit(rs, w, h); e != nil {
		panic(e)
	}
	fmt.Printf("Alvinella pompejana: %d literal protein rows\n", len(rs))
}
func review(rows map[int]map[string]string) ([]rec, error) {
	want := map[string]string{"A": "gi ID", "B": "Accession", "C": "Protein ID", "D": "best hit", "E": "%ID", "F": "comments", "J": "nuc. seq", "K": "protein seq"}
	for c, v := range want {
		if rows[1][c] != v {
			return nil, fmt.Errorf("header %s=%q", c, rows[1][c])
		}
	}
	var out []rec
	for row := 2; row <= 187; row++ {
		v := rows[row]
		r := rec{Row: row, GI: strings.TrimSpace(v["A"]), Accession: strings.TrimSpace(v["B"]), ProteinID: strings.TrimSpace(v["C"]), Hit: strings.TrimSpace(v["D"]), Identity: strings.TrimSpace(v["E"]), Comment: strings.TrimSpace(v["F"]), Genus: strings.TrimSpace(v["H"]), Species: strings.TrimSpace(v["I"])}
		raw := strings.TrimSpace(v["K"])
		if raw == "" {
			return nil, fmt.Errorf("row %d empty protein", row)
		}
		m := cypRE.FindString(r.Hit + " " + r.Comment + " " + r.ProteinID)
		if m == "" && strings.HasPrefix(strings.ToUpper(strings.TrimSpace(r.Hit)), "2AY") {
			m = "CYP" + strings.Fields(r.Hit)[0]
		}
		if m == "" {
			return nil, fmt.Errorf("row %d no CYP label hit=%q", row, r.Hit)
		}
		r.Symbol = strings.TrimRight(m, ".)")
		var b strings.Builder
		lower := false
		for _, x := range raw {
			if unicode.IsSpace(x) {
				continue
			}
			if unicode.IsLower(x) {
				lower = true
			}
			u := unicode.ToUpper(x)
			if (u < 'A' || u > 'Z') && x != '*' && x != '-' {
				return nil, fmt.Errorf("row %d residue %q", row, x)
			}
			b.WriteRune(x)
		}
		r.Sequence = b.String()
		if strings.HasSuffix(r.Sequence, "*") {
			r.Sequence = strings.TrimSuffix(r.Sequence, "*")
			r.Status = append(r.Status, "terminal-stop-removed")
		}
		if strings.Contains(r.Sequence, "*") {
			r.Status = append(r.Status, "internal-stop-retained")
		}
		if strings.ContainsAny(r.Sequence, "Xx") {
			r.Status = append(r.Status, "ambiguous-X-retained")
		}
		if lower {
			r.Status = append(r.Status, "source-lowercase-retained")
		}
		if len(r.Sequence) < 350 {
			r.Status = append(r.Status, "short-source-fragment")
		}
		if r.GI == "" {
			r.Status = append(r.Status, "source-consensus-row")
		}
		lc := strings.ToLower(r.Comment + " " + r.Hit)
		if strings.Contains(lc, "partial") || strings.Contains(lc, "missing") || strings.Contains(lc, "consensus") {
			r.Status = append(r.Status, "source-partial-or-consensus")
		}
		r.ID = first(r.ProteinID, r.Accession, r.GI, r.Symbol+" consensus row "+strconv.Itoa(row))
		r.Key = fmt.Sprintf("pompeii-worm:all-clans-unique:row-%04d:%s", row, strings.Trim(keyRE.ReplaceAllString(strings.ToLower(r.ID), "-"), "-"))
		out = append(out, r)
	}
	if len(out) != 186 {
		return nil, fmt.Errorf("count=%d", len(out))
	}
	return out, nil
}
func first(v ...string) string {
	for _, x := range v {
		if strings.TrimSpace(x) != "" {
			return strings.TrimSpace(x)
		}
	}
	return ""
}
func writeCSV(rs []rec, h string) error {
	p := filepath.Join("sources", "reviewed", "animals", "pompeii-worm.csv")
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	f, e := os.Create(p)
	if e != nil {
		return e
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "gi", "accession", "protein_id", "best_hit", "identity", "comment", "review_status"})
	for _, r := range rs {
		note := fmt.Sprintf("Pompeii worm All clans unique row %d; GI=%s; accession=%s; protein ID=%s; best hit=%s; identity=%s; comment=%s", r.Row, r.GI, r.Accession, r.ProteinID, r.Hit, r.Identity, r.Comment)
		_ = w.Write([]string{"animals", "Alvinella pompejana", r.Symbol, r.ID, r.Key, url, note, r.Sequence, file, h, sheet, strconv.Itoa(r.Row), r.GI, r.Accession, r.ProteinID, r.Hit, r.Identity, r.Comment, strings.Join(r.Status, ";")})
	}
	w.Flush()
	return w.Error()
}
func writeAudit(rs []rec, wb *plantxlsx.Workbook, h string) error {
	c := map[string]int{}
	for _, r := range rs {
		for _, s := range r.Status {
			c[s]++
		}
	}
	ns := make([]string, 0, len(c))
	for n := range c {
		ns = append(ns, n)
	}
	sort.Strings(ns)
	var b strings.Builder
	fmt.Fprintf(&b, "# Animal resource review: Alvinella pompejana (Pompeii worm)\n\n- Source file: `%s`\n- URL: %s\n- SHA-256: `%s`\n- Workbook sheets inspected: 15\n- Authoritative sheet: `%s`\n- Exact accepted rows: `2–187`\n- Protein column: `K (protein seq)`\n- Used range: `%s`\n- Hidden rows: `%v`\n- Hidden columns: `%v`\n- Merged cells: `%v`\n- Accepted records: `%d`\n- Literal proteins: `%d`\n- Review status: `complete`\n\n", file, url, h, sheet, wb.Dimension, wb.HiddenRows, wb.HiddenColumns, wb.MergedCells, len(rs), len(rs))
	b.WriteString("## Resource-specific interpretation\n\nThe workbook contains BLAST hit tables, per-clan candidate lists, two alternate clan orderings, a consensus-only sheet and the final `All clans unique` relationship table. Rows 2–187 of that table are the authoritative 186 protein-bearing rows; rows 188–190 are blank/count separators and rows 191–194 are explicitly labelled false positives without protein. Protein is taken only from column K, never translated from nucleotide column J. The 26 source consensus rows with blank GI are retained because the final table supplies explicit consensus protein and CYP-labelled hit; they receive stable row-based IDs. Candidate/clan sheets and the separate consensus sheet are not appended as duplicates. Internal stops, X/x, fragments and consensus status remain literal; only terminal stops are removed.\n\n## Status counts\n\n| Status | Records |\n|---|---:|\n")
	for _, n := range ns {
		fmt.Fprintf(&b, "| %s | %d |\n", n, c[n])
	}
	p := filepath.Join("docs", "animals", "pompeii-worm.md")
	if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	return os.WriteFile(p, []byte(b.String()), 0644)
}
