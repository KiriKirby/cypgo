# Bacteria resource review: allbacteria (2009)

- Source file: `bacteria-allbacteria.txt`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/allbacteria.doc
- Source SHA-256: `ab82910e9f43b306de8946f723df050502eddb088e11c6920c41922c780f819b`
- Physical `>` blocks: `967`
- Accepted source blocks: `967`
- Blocks with literal protein: `933`
- Missing-sequence blocks retained with provenance: `34`
- Distinct organism labels: `191`
- Review status: `complete`

The source title states 923 updated sequences (and an earlier 892-entry release), while the physical normalized file contains 967 CYP blocks. Every block is retained independently in source order, including variants, renamed/deleted records, duplicate names, pseudogenes, fragments and sequence-missing genome annotations. Protein comes only from the longest uninterrupted literal residue run inside that block; coordinates, accession lines, similarity prose and annotations are never concatenated. Only terminal `*` is removed; X/x, gaps, internal stops and short fragments remain literal.

## Status counts

| Status | Count |
|---|---:|
| ambiguous-X-or-x | 7 |
| internal-stop | 21 |
| literal-protein | 933 |
| reviewed-bacteria-2009 | 967 |
| sequence-missing | 34 |
| short-sequence | 62 |
| source-deletion | 4 |
| source-fragment | 2 |
| source-frameshift | 3 |
| source-partial | 2 |
| source-pseudogene | 5 |
| source-renamed | 5 |
