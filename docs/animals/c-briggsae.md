# Animal resource review: Caenorhabditis briggsae

- Source document: `animals-C.briggsae.doc`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/C.briggsae.doc
- DOC SHA-256: `11239f44d3bf0eaf9afdcd73985f725a9ab3c94d73db204a46d4573b51e441f5`
- Reviewed normalized text: `animals-C.briggsae.txt`
- Text SHA-256: `5a4d576e29f485970c4fc7143b8a489c0449b0b3309e9911bd1faa9372b84bfa`
- Normalized lines: `307`
- Explicit CYP reconstruction blocks: `23`
- Literal proteins: `23`
- Missing sequences: `0`
- Review status: `complete`

## Resource-specific interpretation

The November 2003 document is explicitly an incomplete, not-yet-formally-named genome assembly and contains 23 physical CYP-labelled reconstruction blocks. Every block remains distinct in source order. Protein is read only from that block's uppercase residue runs after removing genomic coordinates and parenthesized exon phase or frameshift annotations. This also accepts the document's intentional coordinate-free continuation lines and space-separated residue fragments, but never introductory prose. Source pseudogenes, short solo exons, X runs and internal stops are retained exactly; only one explicit terminal stop is removed. No DNA is translated and no external sequence is substituted.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 2 |
| internal-stop-retained | 4 |
| short-source-fragment | 3 |
| source-possible-pseudogene | 1 |
| source-pseudogene | 5 |
| terminal-stop-removed | 21 |
