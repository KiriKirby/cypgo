package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadRecords(t *testing.T) []record {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "raw", sourceFile))
	if err != nil {
		t.Fatal(err)
	}
	text, hash, err := readSource(filepath.Join("..", "..", "raw", sourceFile))
	if err != nil {
		t.Fatal(err)
	}
	if hash != sourceSHA {
		t.Fatalf("hash=%s", hash)
	}
	records, err := parse(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 || len(records) != wantBlocks {
		t.Fatalf("bytes=%d records=%d", len(b), len(records))
	}
	return records
}

func TestParse2000CompleteBlocks(t *testing.T) {
	records := loadRecords(t)
	withSequence := 0
	for _, r := range records {
		if r.Sequence != "" {
			withSequence++
		}
		if r.Symbol == "" || r.Header == "" || r.Species == "" {
			t.Fatalf("incomplete block: %#v", r)
		}
		for _, aa := range r.Sequence {
			if !strings.ContainsRune(sequenceAlphabet, aa) {
				t.Fatalf("block %d unexpected residue %q", r.Block, aa)
			}
		}
	}
	if withSequence != 150 {
		t.Fatalf("literal sequences=%d want 150", withSequence)
	}
	if records[0].Symbol != "CYP51" || len(records[0].Sequence) != 450 {
		t.Fatalf("first=%#v", records[0])
	}
	if records[78].Symbol != "CYP122A3" || len(records[78].Sequence) != 388 {
		t.Fatalf("middle=%#v", records[78])
	}
	if records[len(records)-1].Symbol != "CYP?" || records[len(records)-1].Species != "Streptomyces noursei" {
		t.Fatalf("last=%#v", records[len(records)-1])
	}
}

func TestExceptional2000Blocks(t *testing.T) {
	records := loadRecords(t)
	byBlock := map[int]record{}
	for _, r := range records {
		byBlock[r.Block] = r
	}
	if byBlock[31].Sequence != "" || !strings.Contains(strings.Join(byBlock[31].Status, ";"), "same-gene-alternate") {
		t.Fatalf("block 31=%#v", byBlock[31])
	}
	if byBlock[45].Sequence != "" || byBlock[46].Sequence == "" {
		t.Fatalf("107L1 aliases: %#v / %#v", byBlock[45], byBlock[46])
	}
	if len(byBlock[69].Sequence) != 0 || !strings.Contains(strings.Join(byBlock[69].Status, ";"), "source-pseudogene") {
		t.Fatalf("pseudogene=%#v", byBlock[69])
	}
	if len(byBlock[73].Sequence) != 66 || !strings.Contains(byBlock[73].Sequence, "X") || !strings.Contains(strings.Join(byBlock[73].Status, ";"), "multiple-sequence-runs-unmerged") {
		t.Fatalf("CYP118=%#v", byBlock[73])
	}
	if byBlock[85].Sequence != "" || !strings.Contains(strings.Join(byBlock[85].Status, ";"), "source-confidential") {
		t.Fatalf("confidential=%#v", byBlock[85])
	}
	if len(byBlock[152].Sequence) != 161 || !strings.Contains(byBlock[152].Sequence, "XX") {
		t.Fatalf("reverse numbered block=%#v", byBlock[152])
	}
	if strings.Contains(byBlock[145].Sequence, "*") || strings.Contains(byBlock[146].Sequence, "*") {
		t.Fatalf("terminal stops not removed")
	}
}

func TestResidueLineLayouts(t *testing.T) {
	if got, ok := residueLine("483 LIFTDPPRHRQLRKLINSGFXXRRVSVLEPKIRKIVXXILDGIEXGAVHEFTEQITAPLP 304"); !ok || len(got) != 60 {
		t.Fatalf("reverse numbered line: %q ok=%v", got, ok)
	}
	if got, ok := residueLine("MAHGIDQHGIDQHGIDQHGIDQHGIDQHSIDQHG\""); !ok || !strings.HasSuffix(got, "QHG") {
		t.Fatalf("quoted line: %q ok=%v", got, ok)
	}
	if got, ok := residueLine("89% to CYP1046A1"); ok || got != "" {
		t.Fatalf("annotation accepted: %q", got)
	}
}
