# Animal resource review: Macaca mulatta (rhesus monkey)

- Source file: `animals-Rhesus.P450s.Dec8.2014.xlsx`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Rhesus.P450s.Dec8.2014.xlsx
- Source SHA-256: `a3ced132952c1f69aabcc7fbee428f8d556b5170457d7a58bd89e3d4ab80741c`
- Workbook sheets: `Rhesus.csv.txt`, `sorted by CYP name`, `Sorted by location`
- Authoritative sheet: `sorted by CYP name`
- Used range: `A1:L120`
- Hidden rows: `[]`
- Hidden columns: `[]`
- Merged cells: `[]`
- Native table parts: `0`
- Accepted records: `110`
- Records with literal protein sequence: `110`
- Missing-sequence records: `0`
- Review status: `complete`

## Resource-specific interpretation

The workbook's first sheet is the 117-row Ensembl BioMart/PF00067 starting candidate set, not the final nomenclature. The other two sheets are manually curated views. `sorted by CYP name` has 110 sequence-bearing CYP rows and is authoritative: it contains the CYP4F77P row absent from `Sorted by location`, calls the hybrid row CYP21A2 rather than CYP21, and uses the later CYP3A143P wording. Neither the BioMart candidates nor the alternate location ordering is appended as additional records.

The footer says `100 sequences are named. 35 are pseudogenes and 65 are genes`, while rows 2–111 actually contain 110 non-empty CYP-name/sequence pairs. All 110 literal named rows are retained and the discrepancy is documented rather than deleting ten records to force the footer total. The exact column-J CYP/comment text remains the displayed symbol because it carries the source's completeness, orthology, old-name, and assembly qualifications. A real Ensembl gene ID is preferred as record ID, followed by Gotoh ID, Ensembl transcript, coordinates, and finally the source CYP/comment when no identifier or complete coordinate exists. Placeholder text such as `not in BIOMART set` and `no gene` is never used as an ID.

The two lowercase `x` residues are retained literally because they occur in the reviewed sequence cells. Only an explicit terminal `*` is removed. Internal stops, pseudogenes, fragments, manually revised sequences, absent BioMart identifiers, and the complete CYP4A22 sequence linked by the source comment to AC203029.4 remain source-distinct. No sequence is filled from the candidate or location sheet, translated, repaired, or deduplicated.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 7 |
| internal-stop-retained | 21 |
| short-source-fragment | 38 |
| source-lowercase-retained | 2 |
| source-partial-or-missing-region | 18 |
| terminal-stop-removed | 50 |

## Representative rows

| Source row | CYP name/comment | ID | Sequence length | Notes |
|---:|---|---|---:|---|
| 2 | CYP1A1 COMP. | ENSMMUG00000022481 | 512 aa | terminal-stop-removed |
| 13 | CYP2B34P (N-term fragment) | chr19:47305256-47305345:1 | 30 aa | short-source-fragment;source-partial-or-missing-region |
| 26 | CYP2C106P my set | MacamulaUn_20.94335 | 202 aa | source-lowercase-retained;ambiguous-X-retained;short-source-fragment |
| 61 | CYP3A155P exon 4 | chr3:46887117-46887212:1 | 32 aa | short-source-fragment;source-partial-or-missing-region |
| 81 | CYP4F77P | chr13:136524019-136525025:-1 | 140 aa | internal-stop-retained;short-source-fragment |
| 111 | CYP51P3 | MacamulaUn_15.114829 | 501 aa | internal-stop-retained |
