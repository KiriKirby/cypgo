# Animal resource review: Danaus plexippus (monarch butterfly)

- Source: `animals-Monarch.butterfly.doc`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/Monarch.butterfly.doc
- DOC SHA-256: `76492051efb222eb09f4329e2c114c7e6cd6279f100fc4a946a7b31baedfcf5f`
- Text: `animals-Monarch.butterfly.txt`
- Text SHA-256: `f5dd54fd81e4447c424929cc3f5c8a71a7312abea056307bfd5f44958118812f`
- Lines: `1288`
- Explicit blocks: `86`
- Literal proteins: `86`
- Review status: `complete`

## Interpretation

The document states that four source sequences were joined to leave 86 sequences, and it physically contains exactly 86 `>` blocks. Each DPGLEAN gene-model block remains distinct even when CYP family labels repeat. Coordinate and phase annotations are removed from literal uppercase protein lines. Parsing stops at a terminal stop so extra exon comparisons and text explicitly marked not-P450 cannot leak into the record. Only terminal stops are removed; internal stops and X remain. No DNA is translated and no external sequence is substituted.

| Status | Records |
|---|---:|
| ambiguous-X-retained | 2 |
| short-source-fragment | 27 |
| source-joined-model | 2 |
| terminal-stop-removed | 4 |
