# Animal resource review: Sus scrofa (pig)

- Source document: `animals-pig.P450s.doc`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/pig.P450s.doc
- DOC SHA-256: `e0b959b9e8ef1692464122f954b6b850fe9e0dac14a161fc9579799da1a0e451`
- Reviewed Word-normalized text: `animals-pig.P450s.txt`
- Text SHA-256: `13ce88c7c9661418235a7b61f786594696da115433de4ef2ec449e7b0e1275df`
- Word layout: `746 paragraphs`, `13 pages`, `0 tables`
- Normalized lines: `747`
- Explicit CYP headers: `87`
- Literal protein sequences: `79`
- Missing sequences: `8`
- Review status: `complete`

## Resource-specific interpretation

The source says 59 different genes but physically contains 87 explicit `>` blocks, including variants, fragments and entries explicitly stating that no ortholog was found. All 87 source blocks remain records. Protein is taken only from complete residue lines inside the same header block; coordinate numbers and the one reviewed `(1)` phase marker in CYP2R1 are removed. CYP3A88 prose before its sequence is excluded. The longest uninterrupted literal residue run is used for this document because some blocks contain prose between independent fragments; CYP2R1's reviewed phase marker remains within one run. Lowercase cow-comparison residues, internal stops and fragments remain literal. Only terminal stops are removed. No missing record is filled externally.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 4 |
| internal-stop-retained | 1 |
| short-source-fragment | 29 |
| source-lowercase-retained | 7 |
| source-sequence-missing | 1 |
| terminal-stop-removed | 21 |
