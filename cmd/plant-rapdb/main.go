// Resource-specific importer for the RAP-DB IRGSP-1.0 representative rice
// annotation and protein FASTA. Only rows explicitly annotated as cytochrome
// P450 in the Description or InterPro columns are accepted.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const annotationURL = "https://rapdb.dna.naro.go.jp/download/archive/irgsp1/IRGSP-1.0_representative_annotation_2026-02-05.tsv.gz"
const proteinURL = "https://rapdb.dna.naro.go.jp/download/archive/irgsp1/IRGSP-1.0_protein_2026-02-05.fasta.gz"
const annotationSHA = "9129C8110CACFFD987EC25FAD231CA48A7B7C205E1BEB1A19C6301CE1FC59724"
const proteinSHA = "E29BB4FFB0EB90D8A4248490DFB7F0516644C5940E3116ACE98438B5CB56BC34"

var cypName = regexp.MustCompile(`(?i)\bCYP[0-9]{1,4}[A-Z]{1,4}[0-9]{1,4}(?:\.[0-9]+)?\b`)

func hash(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	return fmt.Sprintf("%X", sha256.Sum256(b)), nil
}
func proteins(path string) (map[string]string, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	out := map[string]string{}
	var id string
	var seq strings.Builder
	flush := func() {
		if id != "" {
			out[id] = strings.TrimSuffix(seq.String(), "*")
		}
		seq.Reset()
	}
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 4096), 4<<20)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ">") {
			flush()
			id = strings.Fields(strings.TrimPrefix(line, ">"))[0]
		} else {
			seq.WriteString(line)
		}
	}
	flush()
	return out, s.Err()
}
func main() {
	ann := flag.String("annotation", "raw/plants-rapdb-IRGSP-1.0_representative_annotation_2026-02-05.tsv", "decompressed annotation TSV")
	fa := flag.String("protein", "raw/plants-rapdb-IRGSP-1.0_protein_2026-02-05.fasta", "decompressed protein FASTA")
	out := flag.String("out", "sources/reviewed/plants/rapdb.csv", "reviewed CSV")
	flag.Parse()
	if got, e := hash(*ann); e != nil || !strings.EqualFold(got, annotationSHA) {
		panic(fmt.Sprintf("annotation SHA mismatch: %s %v", got, e))
	}
	if got, e := hash(*fa); e != nil || !strings.EqualFold(got, proteinSHA) {
		panic(fmt.Sprintf("protein SHA mismatch: %s %v", got, e))
	}
	ps, e := proteins(*fa)
	if e != nil {
		panic(e)
	}
	f, e := os.Open(*ann)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 8<<20)
	if !scanner.Scan() {
		panic("empty annotation TSV")
	}
	header := strings.Split(strings.TrimSuffix(scanner.Text(), "\r"), "\t")
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.TrimSpace(h)] = i
	}
	get := func(row []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}
	rows := [][]string{{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "review_status"}}
	count := 0
	for n := 2; scanner.Scan(); n++ {
		row := strings.Split(strings.TrimSuffix(scanner.Text(), "\r"), "\t")
		desc, ipr := get(row, "Description"), get(row, "InterPro")
		if !strings.Contains(strings.ToLower(desc), "cytochrome p450") && !strings.Contains(strings.ToLower(ipr), "cytochrome p450") {
			continue
		}
		id := get(row, "Transcript_ID")
		if id == "" {
			continue
		}
		all := strings.Join([]string{get(row, "RAP-DB Gene Symbol"), get(row, "RAP-DB Gene Name Synonym(s)"), get(row, "CGSNL Gene Symbol"), get(row, "Oryzabase Gene Symbol Synonym(s)"), desc}, " ")
		sym := desc
		if m := cypName.FindString(all); m != "" {
			sym = strings.ToUpper(m)
		}
		note := fmt.Sprintf("RAP-DB IRGSP-1.0 representative annotation row %d; locus=%s; description=%s; InterPro=%s; annotation URL=%s; protein URL=%s", n, get(row, "Locus_ID"), desc, ipr, annotationURL, proteinURL)
		rows = append(rows, []string{"plants", "Oryza sativa", sym, id, fmt.Sprintf("rapdb:IRGSP-1.0:2026-02-05:%s", id), proteinURL, note, ps[id], filepath.Base(*fa), proteinSHA, fmt.Sprintf("annotation-row-%d", n), "reviewed-auxiliary-rapdb"})
		count++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	if count != 384 {
		panic(fmt.Sprintf("unexpected RAP-DB P450 count: %d", count))
	}
	if e := os.MkdirAll(filepath.Dir(*out), 0755); e != nil {
		panic(e)
	}
	o, e := os.Create(*out)
	if e != nil {
		panic(e)
	}
	defer o.Close()
	w := csv.NewWriter(o)
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	if e := w.Error(); e != nil {
		panic(e)
	}
	fmt.Printf("RAP-DB P450 records: %d\n", count)
}
