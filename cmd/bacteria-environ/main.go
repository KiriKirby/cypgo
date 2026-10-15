package main

import (
	"crypto/sha256"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const sourceURL = "https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/environ.doc"
const sourceSHA = "0d7224857fcdef55c7a137dccfac5abb4c288683b99991b26745cf702a20ed79"

var headerRE = regexp.MustCompile(`^>\s*(\S+)(?:\s+(.*))?$`)
var translationRE = regexp.MustCompile(`^\s*/translation="(.*)$`)
var coordLineRE = regexp.MustCompile(`^\s*\d+\s+(.+?)\s+\d+\s*$`)

type rec struct {
	block, line                  int
	id, header, sequence, sample string
}

func parse(text string) ([]rec, error) {
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n"), "\n")
	type st struct {
		i                int
		id, tail, header string
	}
	var starts []st
	for i, line := range lines {
		if m := headerRE.FindStringSubmatch(strings.TrimSpace(line)); len(m) > 0 {
			starts = append(starts, st{i, m[1], m[2], strings.TrimPrefix(strings.TrimSpace(line), ">")})
		}
	}
	if len(starts) != 168 {
		return nil, fmt.Errorf("environment header count changed: %d", len(starts))
	}
	out := make([]rec, 0, len(starts))
	for n, s := range starts {
		end := len(lines)
		if n+1 < len(starts) {
			end = starts[n+1].i
		}
		seq := ""
		var fallback strings.Builder
		inTranslation := false
		for _, line := range lines[s.i+1 : end] {
			t := strings.TrimSpace(line)
			if m := translationRE.FindStringSubmatch(line); len(m) > 0 {
				inTranslation = true
				seq += strings.TrimSpace(strings.TrimSuffix(m[1], `"`))
				if strings.HasSuffix(m[1], `"`) {
					inTranslation = false
				}
				continue
			}
			if inTranslation {
				part := strings.TrimSpace(t)
				if strings.HasSuffix(part, `"`) {
					inTranslation = false
					part = strings.TrimSuffix(part, `"`)
				}
				seq += part
				continue
			}
			if m := coordLineRE.FindStringSubmatch(line); len(m) == 2 {
				part := strings.Map(func(r rune) rune {
					if strings.ContainsRune("ACDEFGHIKLMNPQRSTVWYBXZJUO*-?", r) {
						return r
					}
					return -1
				}, m[1])
				if len(part) >= 8 {
					fallback.WriteString(part)
				}
			}
		}
		seq = strings.Map(func(r rune) rune {
			if strings.ContainsRune("ACDEFGHIKLMNPQRSTVWYBXZJUO*-?", r) {
				return r
			}
			return -1
		}, seq)
		seq = strings.TrimRight(seq, "*")
		if seq == "" {
			seq = strings.TrimRight(fallback.String(), "*")
		}
		sample := "marine environmental sample"
		if s.i >= 2280 {
			sample = "acid mine drainage microbial pink biofilm"
		}
		out = append(out, rec{n + 1, s.i + 1, s.id, s.header, seq, sample})
	}
	return out, nil
}

func main() {
	in := flag.String("input", "raw/bacteria-environ.txt", "")
	out := flag.String("out", "sources/reviewed/bacteria/bacteria-environ.csv", "")
	audit := flag.String("audit", "docs/bacteria/bacteria-environ.md", "")
	flag.Parse()
	b, e := os.ReadFile(*in)
	if e != nil {
		panic(e)
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	if h != sourceSHA {
		panic(fmt.Sprintf("sha changed: %s", h))
	}
	rs, e := parse(string(b))
	if e != nil {
		panic(e)
	}
	if e = os.MkdirAll(filepath.Dir(*out), 0755); e != nil {
		panic(e)
	}
	f, e := os.Create(*out)
	if e != nil {
		panic(e)
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{"category", "species", "symbol", "id", "record_key", "source_url", "source_note", "sequence", "source_file", "source_sha256", "source_block", "source_line", "source_header", "review_status"})
	nseq := 0
	for _, r := range rs {
		if r.sequence != "" {
			nseq++
		}
		note := fmt.Sprintf("Environmental P450 sample block %d at normalized line %d; source header=%s", r.block, r.line, r.header)
		_ = w.Write([]string{"bacteria", r.sample, r.id, r.id, fmt.Sprintf("bacteria-environ:block-%04d:line-%05d:%s", r.block, r.line, strings.ToLower(r.id)), sourceURL, note, r.sequence, "bacteria-environ.doc", sourceSHA, strconv.Itoa(r.block), strconv.Itoa(r.line), r.header, "reviewed-bacteria-environ;literal-protein"})
	}
	w.Flush()
	f.Close()
	if e = w.Error(); e != nil {
		panic(e)
	}
	if e = os.MkdirAll(filepath.Dir(*audit), 0755); e != nil {
		panic(e)
	}
	body := fmt.Sprintf("# Bacteria resource review: environmental samples\n\n- Source file: `bacteria-environ.doc`\n- URL: %s\n- Normalized text SHA-256: `%s`\n- DOC SHA-256: `0f4f54d296572dd42bed199a60b0c42f000936244b13727f24fb589d94a3b2d7`\n- Word layout: 2,337 paragraphs, 38 pages, no tables\n- Explicit accession blocks: `%d`\n- Blocks with literal protein: `%d`\n- Review status: `complete`\n\nThe source explicitly declares 163 seawater and 5 acid-mine-biofilm environmental sequences. Each accession block is retained independently. Species is deliberately represented as the source sample class rather than guessed from an accession or external database. Protein is accepted from the block's literal `/translation` field, or for the eleven blocks that use the alternate source layout, from the complete coordinate-bounded protein lines. DNA coordinates, similarity notes and annotations are excluded.\n", sourceURL, sourceSHA, len(rs), nseq)
	if e = os.WriteFile(*audit, []byte(body), 0644); e != nil {
		panic(e)
	}
	fmt.Printf("environment: %d records, %d sequences\n", len(rs), nseq)
}
