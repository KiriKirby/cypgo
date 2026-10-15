package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	textFile = "bacteria-bact.abund.txt"
	docFile  = "bacteria-bact.abund.doc"
	url      = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bact.abund.doc"
	textSHA  = "ef0fb32a419d754428a74cf68c8de9b2dc44b4a95f8533ebb43fca2b4731e257"
	docSHA   = "436fa7259c42a3ad6600e0bc90e4884b383e6af9e486d387ac31d1fb0350fc45"
)

var proteinOnly = regexp.MustCompile(`^[ACDEFGHIKLMNPQRSTVWYBXZJUO*?-]{30,}$`)

type review struct {
	Lines, CYPIndexLines, FASTAHeaders, AlignmentLines, ProteinLines int
}

func verify(path, want string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	got := fmt.Sprintf("%x", sha256.Sum256(b))
	if got != want {
		return nil, fmt.Errorf("hash changed for %s: got %s want %s", path, got, want)
	}
	return b, nil
}

func inspect(text string) (review, error) {
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	r := review{Lines: len(lines)}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "CYP") {
			r.CYPIndexLines++
		}
		if strings.HasPrefix(line, ">") {
			r.FASTAHeaders++
		}
		if strings.HasPrefix(line, "Query:") || strings.HasPrefix(line, "Sbjct:") {
			r.AlignmentLines++
		}
		if proteinOnly.MatchString(line) {
			r.ProteinLines++
		}
	}
	if r.Lines != 1086 || r.CYPIndexLines != 490 || r.FASTAHeaders != 0 || r.AlignmentLines != 0 || r.ProteinLines != 0 {
		return r, fmt.Errorf("abundance index structure changed: %+v", r)
	}
	return r, nil
}

func writeAudit(path string, r review) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body := fmt.Sprintf(`# Bacteria resource review: bacterial P450 abundance by taxonomic group

- Source document: %s
- URL: %s
- DOC SHA-256: %s
- Reviewed normalized text: %s
- Normalized text SHA-256: %s
- Logical lines: %d
- CYP name/index lines: %d
- FASTA headers: %d
- BLAST Query/Sbjct lines: %d
- Independent protein-sequence lines: %d
- PGD records contributed: 0
- Review status: complete

## Resource-specific interpretation

This document is a taxonomy and abundance index. It contains hierarchy headings, organism labels, counts, CYP names, accessions, coordinates, aliases, pseudogene labels, and comments. Its CYP lines are references to entries counted under taxa; they are not sequence records and have no literal protein field. No line starts a FASTA block, and no independent amino-acid sequence is present. Consequently this reviewed source intentionally contributes zero PGD records. CYP names and coordinates are not converted into placeholder records, and no sequence is fetched from another source.
`, docFile, url, docSHA, textFile, textSHA, r.Lines, r.CYPIndexLines, r.FASTAHeaders, r.AlignmentLines, r.ProteinLines)
	return os.WriteFile(path, []byte(body), 0o644)
}

func main() {
	in := flag.String("input", filepath.Join("raw", textFile), "normalized abundance index")
	doc := flag.String("doc", filepath.Join("raw", docFile), "original Word document")
	audit := flag.String("audit", filepath.Join("docs", "resources", "bacteria-bacterial-cytochrome-p450-sequences-by-taxonomic-group.md"), "resource audit")
	flag.Parse()
	b, err := verify(*in, textSHA)
	if err != nil {
		panic(err)
	}
	if _, err := verify(*doc, docSHA); err != nil {
		panic(err)
	}
	r, err := inspect(string(b))
	if err != nil {
		panic(err)
	}
	if err := writeAudit(*audit, r); err != nil {
		panic(err)
	}
	fmt.Printf("bacteria abundance index: %d CYP index lines, 0 PGD records\n", r.CYPIndexLines)
}
