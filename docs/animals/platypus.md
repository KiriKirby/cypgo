# Animal resource review: Ornithorhynchus anatinus (duck-billed platypus)

- Source document: `animals-duckbill.doc`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/duckbill.doc
- DOC SHA-256: `703f658f714a7f988ea35138f76bff5fdcb6fcff2541de10c163b5cbb1fb3c96`
- Reviewed normalized text: `animals-duckbill.txt`
- Text SHA-256: `7cc53100a2fff24409370c3d5e2df69ea53108c5b45a4565cb0e068b563a2693`
- Normalized lines: `570`
- Explicit CYP blocks: `55`
- Literal proteins: `55`
- Missing sequences: `0`
- Review status: `complete`

## Resource-specific interpretation

The document declares 55 Entrez-protein results, two false positives, and 53 matching P450 sequences, but physically provides 55 explicit CYP-labelled blocks representing the author's curated collection, including source-distinct genomic reconstructions and a combined CYP17A1 record. All 55 blocks are retained in source order. Within each block, only all-uppercase literal residue lines are joined; mixed-case prose, coordinates, `(gap)` markers, BLAST comments and nucleotide annotations are excluded. This resource-specific rule permits separated literal fragments in a single named block without accepting prose. Terminal stops are removed; internal stops, X and source fragments remain literal. No DNA is translated and no external sequence is substituted.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 3 |
| short-source-fragment | 26 |
| source-partial | 27 |
| terminal-stop-removed | 6 |
