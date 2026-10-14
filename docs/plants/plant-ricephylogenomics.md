# Plant resource audit: Rice P450 Phylogenomics downloads

- Protein source: https://ricephylogenomics.ucdavis.edu/p450/download/RP450Protein.txt
- cDNA companion: https://ricephylogenomics.ucdavis.edu/p450/download/RP450cDNA.txt
- Genomic companion: https://ricephylogenomics.ucdavis.edu/p450/download/RP450Genomic.txt
- Protein SHA-256: `d08f11744e02d4e8d533f4443eb68eb48f0d02a78a1cad97fcc980726b0d3a4a`
- cDNA SHA-256: `b182a68fae454c804d70bccc2e3764039edad43193d163ee4c16b79235d1743b`
- Genomic SHA-256: `679f42486688717ad44170d53ec9e44510bd42493262aa0014974658d0adbf7a`
- Parser: `cmd/plant-ricephylogenomics`
- Accepted protein records: 341
- Accepted cDNA records: 341 companion records
- Accepted genomic records: 302 companion records
- Protein records with sequence: 341

## Import rule

The protein download is the release record source. Each protein FASTA header is retained as provenance, the first pipe-delimited field is used as the stable ID, and the tab-separated annotation is retained in the description. A terminal `*` is removed as an explicit protein terminator. The cDNA and genomic downloads are not inserted into the protein `Sequence` field. Their URLs are retained in every protein record description so users can access the nucleotide forms.

This resource is distinct from the Dr. Nelson `rice.FASTA.dec30.html` and `rice_annotation.P450.xls` resources. Same-species or similar CYP names do not cause cross-resource overwrites.

## Representative records

| Position | Protein header prefix | ID | Sequence |
|---|---|---|---:|
| First | `LOC_Os11g32240.1` | `LOC_Os11g32240.1` | present |
| Middle | `LOC_Os04g48210.1` | `LOC_Os04g48210.1` | present |
| Last | `LOC_Os12g44290.1` | `LOC_Os12g44290.1` | present |

The parser test fixes all three source hashes, the 341/341/302 record counts, source URLs, companion provenance, and representative records.
