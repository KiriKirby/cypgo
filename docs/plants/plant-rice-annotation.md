# Plant resource audit: rice annotation

- Source file: `plants-rice_annotation.P450.xls`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/rice_annotation.P450.xls
- SHA-256: `3532c8470d07188dc00d5135caa7772af80429d9b5ea471a2bb3ac25214a91c1`
- Container: BIFF8/OLE legacy Excel workbook (`.xls`)
- Parser: `cmd/plant-rice-annotation`
- Accepted records: 412
- Literal protein sequences: 0
- Missing protein sequences: 412
- Duplicate Gene Name values: 0
- Foreign-species rows: 0 (the workbook identifies the table as Japonica rice)
- Unresolved rows: 0

## Workbook inspection

Excel reports three worksheets. `Sheet1` has a used range of 2057 rows by 11 columns; only rows 1-418 contain values. `Sheet2` and `Sheet3` are empty. There are no hidden rows or hidden columns in the inspected data range. The source contains no protein-sequence column and no literal FASTA records.

The reviewed table header is row 6. Records are exactly rows 7-418, in source order. Its 11 columns are `Locus`, `Gene Name`, `Gene Description`, `Alt Gene Name`, `GenBank Acc Genomic`, `GenBank Acc cDNA`, `GenBank Acc Protein`, `Mutant`, `Criteria`, `Source`, and `Comment`.

All 412 rows have a unique non-empty `Gene Name`, a non-empty `Gene Description`, and a non-empty `GenBank Acc Genomic`. Other column coverage is: Alt Gene Name 9, GenBank Acc cDNA 0, GenBank Acc Protein 0, Mutant 0, Criteria 1, Source 23, Comment 150. Blank fields remain blank.

## Import rule

Each table row becomes one `Oryza sativa` record. `Gene Name` is both ID and Symbol. The description preserves every source column with `Sheet1` and the exact Excel row number. Record keys use `rice-annotation:sheet1:row-NNNN`. Sequence is deliberately empty because the workbook has no literal protein sequence. No outside sequence database is consulted.

This annotation workbook remains a distinct resource from `rice.FASTA.dec30.html`. The 15 FASTA-page records are not used to fill workbook rows because the two source resources must retain independent provenance. Where names overlap, both source records remain available.

## Representative records

| Position | Source row | Gene Name | Description | Genomic accession |
|---|---:|---|---|---|
| First | 7 | CYP51G1 | Obtusifoliol 14-alpha demethylase | AP008217.1 |
| Middle | 212 | CYP81L4 | cytochrome P450 of unknown function | AP005285.1 |
| Last | 418 | CYP735A4 | possible cytoknin hydroxylase | AP005732.2 |

The parser test fixes the workbook hash, worksheet structure, headers, 412-row count, empty-sequence rule, source URL, and these first/middle/last records.
