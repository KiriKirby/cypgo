# Animal resource review: Pan troglodytes (chimp)

- Source file: `animals-Pan.trog.P450s.Dec8.2014.xlsx`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Pan.trog.P450s.Dec8.2014.xlsx
- Source SHA-256: `2bae0602c8d0aaee7e87f9feb826e1705a8124aa419eae124583967c831a9c07`
- Workbook sheets: `Pan.troglodytes Gotoh`, `Pan Troglodytes Ensembl`, `combined by chr order`, `sorted by CYP name`
- Authoritative sheet: `sorted by CYP name`
- Used range: `A1:M122`
- Hidden rows: `[]`
- Hidden columns: `[]`
- Merged cells: `[]`
- Native table parts: `0`
- Accepted records: `121`
- Records with literal protein sequence: `121`
- Missing-sequence records: `0`
- Review status: `complete`

## Resource-specific interpretation

The workbook contains two intermediate source sheets (`Pan.troglodytes Gotoh` and `Pan Troglodytes Ensembl`) and two curated combined views. The two combined views describe the same biological row set in different orders, but the final `sorted by CYP name` sheet contains five later nomenclature edits relative to `combined by chr order` (including CYP3A156P/CYP3A157P, CYP51A1, and the expanded CYP46A11P and CYP21A1 hybrid labels). Only the final sorted sheet is therefore published; intermediate and alternate-sort rows are not appended or deduplicated into extra records.

Rows 2–122 each contain an explicit CYP name in column L and literal protein text in column M. All 121 rows are retained, including complete genes, short exon fragments, partial sequences, pseudogene orthologs, two source-distinct rows both labeled `CYP2B6 partial seq`, sequences with internal stops, and manually revised/augmented sequence text. Column A Ensembl gene ID is the preferred searchable ID. When it is absent, the Gotoh identifier is used; if both are absent, the chromosome coordinate is used. Transcript, Gotoh hit, coordinates, human hit, percent identities, and comments remain row-level provenance.

Whitespace inside the CYP2S1 source cell is spreadsheet layout and is joined without inserting residues. The mixed-case CYP2B6 sequence is retained in its literal source case because lowercase marks the workbook author's augmented segment. Only one explicit terminal `*` is removed. Every internal `*`, `X`, short fragment, duplicate symbol, and source comment remains intact. No sequence is translated, repaired, filled from another sheet/database, or deduplicated.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 2 |
| duplicate-symbol-source-row | 2 |
| internal-stop-retained | 34 |
| short-source-fragment | 49 |
| source-lowercase-retained | 1 |
| source-partial-or-missing-region | 17 |
| source-whitespace-joined | 1 |
| terminal-stop-removed | 44 |

## Representative rows

| Source row | CYP name | ID | Sequence length | Notes |
|---:|---|---|---:|---|
| 2 | CYP1A1 | ENSPTRG00000007280 | 512 aa | terminal-stop-removed |
| 10 | CYP2B6 | ENSPTRG00000023029a | 493 aa | source-lowercase-retained |
| 39 | CYP2S1 | ENSPTRG00000011017 | 565 aa | source-whitespace-joined;terminal-stop-removed |
| 62 | CYP4A27P ORTHOLOG | chr1:47373222-47373377:-1 | 52 aa | short-source-fragment |
| 122 | CYP51P3 | Pan_trogUn_20.150063 | 507 aa | ambiguous-X-retained |
