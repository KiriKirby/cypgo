# Animal resource review: Callithrix jacchus (marmoset)

- Source file: `animals-Marmoset.P450s.Dec8.2014.xlsx`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Marmoset.P450s.Dec8.2014.xlsx
- Source SHA-256: `6d80ce5b4cb22d693a03bb251fa3c6f8b10ad101000c65fa2efb4a23287700c5`
- Workbook sheets: `Sorted by location`, `Sorted by CYP name`
- Authoritative sheet: `Sorted by CYP name`
- Used range: `A1:E91`
- Hidden rows: `[]`
- Hidden columns: `[]`
- Merged cells: `[]`
- Native table parts: `0`
- Accepted records: `87`
- Records with literal protein sequence: `87`
- Missing-sequence records: `0`
- Review status: `complete`

## Resource-specific interpretation

The workbook has two alternate orderings of the same 87 source-distinct rows. Their CYP-name/sequence multisets are identical, so only `Sorted by CYP name` is published. Rows 2–88 each contain an explicit CYP name, the literal species label `marmoset`, and a sequence. The Ensembl ID is preferred as ID; 32 rows without one use the exact genomic location, and a CYP name is the final fallback. The footer notes that Ensembl regions may span gene clusters and that some nucleotide regions were edited to coding boundaries; these statements are preserved as provenance and do not authorize sequence repair.

Lowercase residues in six cells, one `-x` source gap, internal stops, short fragments, pseudogenes, incomplete genes and revised sequences are retained literally. Only explicit terminal `*` markers are removed. No case conversion, translation, repair, cross-sheet merge, or deduplication is performed.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 4 |
| internal-stop-retained | 14 |
| short-source-fragment | 24 |
| source-gap-retained | 1 |
| source-lowercase-retained | 6 |
| source-partial-or-missing-region | 11 |
| terminal-stop-removed | 3 |

## Representative rows

| Source row | CYP name | ID | Sequence length | Notes |
|---:|---|---|---:|---|
| 2 | CYP1A1 ortholog | ENSCJAG00000017698 | 512 aa |  |
| 12 | CYP2C76 ortholog | ENSCJAG00000015961 | 489 aa | source-lowercase-retained |
| 42 | CYP4A49P (with a gap) | 7:81721642-81728991:-1 | 343 aa | source-lowercase-retained;ambiguous-X-retained;source-gap-retained;short-source-fragment;source-partial-or-missing-region |
| 45 | CYP4F2 (93% TO CYP4F2) MISSING MIDDLE EXONS 5 AND PART OF 6 | ENSCJAG00000000362 | 510 aa | source-lowercase-retained;source-partial-or-missing-region |
| 88 | CYP51P4 (seq revised) | 4:152908787-152910303 | 505 aa | internal-stop-retained |
