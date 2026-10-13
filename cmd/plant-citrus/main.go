package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KiriKirby/cypgo/internal/plantxlsx"
)

type spec struct {
	slug, species, file, url string
	sheets                   []string
}

var specs = map[string]spec{
	"citrus-clementina": {"citrus-clementina", "Citrus clementina", "plants-Citrus.clementina.P450s.xlsx", "https://drnelson.uthsc.edu/Citrus.clementina.P450s.xlsx", []string{"Citrus clementina other clans", "Citrus clementina CYP71 clan"}},
	"citrus-sinensis":   {"citrus-sinensis", "Citrus sinensis", "plants-Citrus.sinensis.P450s.xlsx", "https://drnelson.uthsc.edu/Citrus.sinensis.P450s.xlsx", []string{"Citrus sinensis CYP71 clan", "Citrus sinensis other clans"}},
}

func main() {
	name := flag.String("resource", "", "")
	in := flag.String("input", "", " ")
	out := flag.String("out", "", " ")
	flag.Parse()
	s, ok := specs[*name]
	if !ok {
		panic("unknown resource")
	}
	if *in == "" {
		*in = filepath.Join("raw", s.file)
	}
	if *out == "" {
		*out = filepath.Join("sources", "reviewed", "plants", s.slug+".csv")
	}
	data, e := os.ReadFile(*in)
	if e != nil {
		panic(e)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	var rows [][]string
	n := 0
	for _, sheet := range s.sheets {
		w, e := plantxlsx.ReadSheet(*in, sheet, false)
		if e != nil {
			panic(e)
		}
		for r := 2; r <= maxRow(w.Rows); r++ {
			v := w.Rows[r]
			id := strings.TrimSpace(v["A"])
			sym := strings.TrimSpace(v["D"])
			if sym == "" {
				sym = strings.TrimSpace(v["F"])
			}
			if id == "" || sym == "" || !strings.HasPrefix(sym, "CYP") {
				continue
			}
			n++
			rows = append(rows, []string{"plants", s.species, sym, id, fmt.Sprintf("%s:%s:%04d", s.slug, sheet, r), s.url, fmt.Sprintf("Workbook sheet %s row %d; gene ID=%s; assigned CYP name=%s; source contains names/annotations only", sheet, r, id, sym), "", s.file, hash, sheet, fmt.Sprint(r), "sequence-missing"})
		}
	}
	if e := os.MkdirAll(filepath.Dir(*out), 0755); e != nil {
		panic(e)
	}
	f, e := os.Create(*out)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_sheet", "source_row", "review_status"})
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	if e := w.Error(); e != nil {
		panic(e)
	}
	fmt.Printf("%s: %d rows\n", s.slug, n)
}

func maxRow(rows map[int]map[string]string) int {
	n := 0
	for r := range rows {
		if r > n {
			n = r
		}
	}
	return n
}
