// Resource-specific importer for the Rice P450 database hosted by the Rice
// Phylogenomics site. Protein FASTA rows become records; cDNA and genomic
// downloads are retained as explicit companion provenance, never as protein.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	proteinURL = "https://ricephylogenomics.ucdavis.edu/p450/download/RP450Protein.txt"
	cdnaURL    = "https://ricephylogenomics.ucdavis.edu/p450/download/RP450cDNA.txt"
	genomicURL = "https://ricephylogenomics.ucdavis.edu/p450/download/RP450Genomic.txt"
	proteinSHA = "d08f11744e02d4e8d533f4443eb68eb48f0d02a78a1cad97fcc980726b0d3a4a"
	cdnaSHA    = "b182a68fae454c804d70bccc2e3764039edad43193d163ee4c16b79235d1743b"
	genomicSHA = "679f42486688717ad44170d53ec9e44510bd42493262aa0014974658d0adbf7a"
)

type fastaRow struct{ header, sequence string }

func readFASTA(path string) ([]fastaRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []fastaRow
	var current *fastaRow
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 4096), 4<<20)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ">") {
			out = append(out, fastaRow{header: strings.TrimSpace(strings.TrimPrefix(line, ">"))})
			current = &out[len(out)-1]
			continue
		}
		if current == nil {
			return nil, fmt.Errorf("sequence before FASTA header in %s", path)
		}
		current.sequence += line
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func hashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func buildRows(proteinPath, cdnaPath, genomicPath string) ([][]string, error) {
	for _, x := range []struct{ path, want string }{{proteinPath, proteinSHA}, {cdnaPath, cdnaSHA}, {genomicPath, genomicSHA}} {
		got, err := hashFile(x.path)
		if err != nil {
			return nil, err
		}
		if got != x.want {
			return nil, fmt.Errorf("source SHA-256 changed for %s: got %s want %s", x.path, got, x.want)
		}
	}
	proteins, err := readFASTA(proteinPath)
	if err != nil {
		return nil, err
	}
	cdna, err := readFASTA(cdnaPath)
	if err != nil {
		return nil, err
	}
	genomic, err := readFASTA(genomicPath)
	if err != nil {
		return nil, err
	}
	if len(proteins) != 341 || len(cdna) != 341 || len(genomic) != 302 {
		return nil, fmt.Errorf("unexpected source counts protein=%d cdna=%d genomic=%d", len(proteins), len(cdna), len(genomic))
	}
	rows := make([][]string, 0, len(proteins))
	for i, p := range proteins {
		parts := strings.SplitN(p.header, "\t", 2)
		identity := parts[0]
		note := ""
		if len(parts) == 2 {
			note = parts[1]
		}
		id := identity
		if pipe := strings.IndexByte(id, '|'); pipe >= 0 {
			id = id[:pipe]
		}
		seq := strings.TrimSuffix(p.sequence, "*")
		desc := fmt.Sprintf("Rice P450 Phylogenomics protein row %d; header=%s; annotation=%s; companion cDNA URL=%s; companion genomic URL=%s", i+1, p.header, note, cdnaURL, genomicURL)
		rows = append(rows, []string{"plants", "Oryza sativa", note, id, fmt.Sprintf("ricephylogenomics:protein:row-%04d:%s", i+1, id), proteinURL, desc, seq, filepath.Base(proteinPath), proteinSHA, fmt.Sprintf("protein:row-%d", i+1), "resource-specific-protein;companion-cdna-genomic"})
	}
	return rows, nil
}

func main() {
	protein := flag.String("protein", "raw/plants-ricephylogenomics-RP450Protein.txt", "protein FASTA")
	cdna := flag.String("cdna", "raw/plants-ricephylogenomics-RP450cDNA.txt", "cDNA FASTA")
	genomic := flag.String("genomic", "raw/plants-ricephylogenomics-RP450Genomic.txt", "genomic FASTA")
	out := flag.String("out", "sources/reviewed/plants/ricephylogenomics.csv", "reviewed CSV")
	flag.Parse()
	rows, err := buildRows(*protein, *cdna, *genomic)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		panic(err)
	}
	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "review_status"})
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	fmt.Printf("rice phylogenomics: %d protein records\n", len(rows))
}
