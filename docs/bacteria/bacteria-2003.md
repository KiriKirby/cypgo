# Bacteria resource review: bacteria.dbs (2003)

- Source file: `bacteria-bacteria.dbs.txt`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacteria.dbs.doc
- Source SHA-256: `bc78753c1953f6cc328fdf608454a7669cf9baffc306cb3d53b0e01415c25eed`
- Normalized text: UTF-16 Word text export, 5,169 lines, no tables
- Source title declaration: `430 bacterial cytochrome P450 sequences Nov. 5, 2003`
- Physical `>` blocks: `555` (553 CYP blocks plus 2 accession-only comparison blocks)
- Accepted CYP blocks: `553`
- Records with literal protein: `553`
- Excluded accession-only comparison blocks: `2`
- Distinct source organism labels: `149`
- Sequence length range: `57-938` aa
- Review status: `complete`

The document title says 556 sequences, but the downloaded normalized source contains 555 physical `>` blocks. Blocks 1-553 have CYP model headers and are retained in source order, including variants, pseudogenes, fragments, half proteins, short N-terminal/C-terminal pieces, duplicate names and confidential source text. Blocks 554-555 are accession-only comparison alignments (`L76374` and `AF071147`) and are excluded because they do not define a CYP record. Protein is read only from literal residue lines inside each block. Standalone numeric coordinate fields and Word line wrapping are removed; five trailing Word quote artifacts are removed only at line ends. Three reviewed blocks carry a literal protein segment on the header line and are handled by explicit block number. Lowercase residues, X/x, gaps and internal stops remain source literal; only terminal `*` markers are removed. No translation, repair, deduplication or external completion is performed.

## Status counts

| Status | Count |
|---|---:|
| ambiguous-X-or-x | 7 |
| internal-stop | 10 |
| literal-protein | 553 |
| reviewed-bacteria-2003 | 553 |
| short-sequence | 39 |
| source-confidential | 1 |
| source-fragment-or-partial | 8 |
| source-inline-protein | 3 |
| source-lowercase-residue | 1 |
| source-pseudogene | 19 |
| source-terminal-stop | 19 |

## Excluded blocks

| Block | Line | Header | Reason |
|---:|---:|---|---|
| 554 | 5159 | L76374 Mycobacterium avium paratuberculosis 39% identical to 107B1 | accession-only comparison block has no CYP model header |
| 555 | 5164 | AF071147 Streptoalloteichus hindustanus cytochrome 61% to 105C1 | accession-only comparison block has no CYP model header |

## Representatives

| Accepted index | Block | Line | Species | Symbol | Length | Status |
|---:|---:|---:|---|---|---:|---|
| 1 | 1 | 4 | Mycobacterium tuberculosis | CYP51 | 451 | reviewed-bacteria-2003;literal-protein |
| 277 | 277 | 2549 | Mycobacterium vanbaalenii | CYP142B1 | 422 | reviewed-bacteria-2003;literal-protein |
| 553 | 553 | 5150 | Mycobacterium flavescens | CYP292A2 | 405 | reviewed-bacteria-2003;literal-protein |
