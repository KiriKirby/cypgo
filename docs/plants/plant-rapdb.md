# Plant resource audit: RAP-DB IRGSP-1.0 rice P450 auxiliary source

- Source: [RAP-DB](https://rapdb.dna.naro.go.jp/index.html), IRGSP-1.0 release `2026-02-05`
- Annotation URL: https://rapdb.dna.naro.go.jp/download/archive/irgsp1/IRGSP-1.0_representative_annotation_2026-02-05.tsv.gz
- Protein URL: https://rapdb.dna.naro.go.jp/download/archive/irgsp1/IRGSP-1.0_protein_2026-02-05.fasta.gz
- Decompressed annotation SHA-256: `9129c8110cacffd987ec25fad231ca48a7b7c205e1beb1a19c6301ce1fc59724`
- Decompressed protein SHA-256: `e29bb4ffb0eb90d8a4248490dfb7f0516644c5940e3116ace98438b5cb56bc34`
- Annotation rows: 40,249 data rows; protein FASTA records: 38,356
- Accepted auxiliary P450 rows: **384**, all with a literal RAP-DB protein sequence

## Resource-specific rule

The importer reads the tab-delimited header and accepts a row only when the exact `Description` or `InterPro` field contains `Cytochrome P450` (case-insensitive). The protein is matched by the exact `Transcript_ID` token in the RAP-DB protein FASTA header. No whole-database protein inference, sequence translation, symbol guessing, or external sequence substitution is performed. Literal terminal `*` markers are removed only at the end of the source protein sequence.

RAP-DB records are auxiliary records. Dr. Nelson and Rice P450 Phylogenomics remain the primary rice sources; RAP-DB IDs, annotation text, IRGSP-1.0 coordinates and protein sequences supplement them and are retained as separate provenance records.

The RAP-DB TSV contains literal unmatched quote characters in some descriptions, so the resource-specific importer uses line-preserving tab splitting rather than a generic CSV quote parser. This preserves every source field and row position.
