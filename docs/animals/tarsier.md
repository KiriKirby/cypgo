# Animal resource review: Tarsius syrichta (tarsier)

- Source file: `animals-Tarsier.P450s.xlsx`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Tarsier.P450s.xlsx
- Source SHA-256: `6ce969639f72ef5dd62f8a6aeb23f37c8f61b59259be9cfc517b548f64ca8b81`
- Workbook sheets: `Tarsier.ensembl.txt`, `Gotoh's set`, `combined, sorted by CYP name`, `87 Genbank proteins`
- Authoritative sheet: `combined, sorted by CYP name`
- Used range: `A1:L113`
- Hidden rows: `[]`
- Hidden columns: `[]`
- Merged cells: `[]`
- Accepted records: `108`
- Records with literal protein sequence: `108`
- Review status: `complete`

## Resource-specific interpretation

The workbook separates 46 Ensembl candidates, 31 Gotoh predictions, and 87 GenBank proteins from the manually combined naming table. Only `combined, sorted by CYP name` is published because it records the reviewed relationship among those sources. Its rows 2–109 contain 108 explicit CYP-name/sequence pairs. Repeated CYP names are source-distinct rows and remain separate.

Gene/GenBank accession is the preferred ID, followed by Gotoh ID, transcript, and CYP name. Literal long X runs, one lowercase `x`, internal stops, fragments, pseudogenes and missing-region sequences remain intact. Only terminal `*` is removed. Candidate sheets are provenance inputs, not extra records; no cross-sheet concatenation, repair, translation, or deduplication is performed.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 33 |
| internal-stop-retained | 7 |
| short-source-fragment | 23 |
| source-lowercase-retained | 1 |
| source-partial-or-missing-region | 57 |
| terminal-stop-removed | 27 |

## Representative rows

| Source row | CYP name | ID | Sequence length | Notes |
|---:|---|---|---:|---|
| 2 | CYP1A1 | ENSTSYG00000000100 | 515 aa | terminal-stop-removed;ambiguous-X-retained;source-partial-or-missing-region |
| 8 | CYP2A13 | ENSTSYG00000003073 | 84 aa | short-source-fragment;source-partial-or-missing-region |
| 43 | CYP2D76P | LOC103264345 | 497 aa | source-lowercase-retained;internal-stop-retained;ambiguous-X-retained |
| 56 | CYP2U1 | ENSTSYG00000008958 | 483 aa | ambiguous-X-retained;source-partial-or-missing-region |
| 109 | CYP51P5 | ABRT02387594.1 | 512 aa | internal-stop-retained |
