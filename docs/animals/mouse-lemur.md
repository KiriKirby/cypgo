# Animal resource review: Microcebus murinus (grey mouse-lemur)

- Source file: `animals-Mouse.lemur.P450s.UCSC.xlsx`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Mouse.lemur.P450s.UCSC.xlsx
- Source SHA-256: `ed63c9d0af23152f43c99b5be79b757eae0d5952dbf154be708a0081df5d2519`
- Authoritative sheet: `Sorted by CYP family`
- Used range: `A1:O187`
- Hidden rows: `[]`
- Hidden columns: `[]`
- Merged cells: `[]`
- Accepted records: `177`
- Literal protein sequences: `177`
- Review status: `complete`

## Resource-specific interpretation

The first worksheet is the UCSC/Gotoh search and joining workspace. Its footer says more than 2,100 browser hits were reduced to 169 and compared with 52 Gotoh sequences, with many still needing assembly. The second worksheet is explicitly the improved CYP-family view and is the only published table. Rows 3–180 contain 177 literal sequence cells. Seven have an empty UCSC-tag/CYP-name cell. A same-row comment/status beginning with CYP is used only when explicit; otherwise the symbol remains empty and Ensembl, Gotoh, or location supplies the ID. No CYP name is inferred from sequence similarity.

All source-distinct fragments and repeated labels remain separate. Lowercase x, gaps, internal stops, Ensembl-inferred/hybrid warnings, incomplete assemblies and short exons remain literal. Only terminal `*` is removed. No external sequence is used and no cross-row assembly is attempted.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 35 |
| internal-stop-retained | 8 |
| short-source-fragment | 128 |
| source-cyp-name-missing | 3 |
| source-gap-retained | 2 |
| source-lowercase-retained | 2 |
| source-partial-or-missing-region | 101 |
| terminal-stop-removed | 25 |
