# Plant resource audit: biblioD historical bibliography

- Source URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/biblioD.html
- Source file: `plants-biblioD.html`
- Source SHA-256: `aac540d08652c2241b7c500f1b7e2c751ab94a8d5e98dd9e04a954cf48252c7b`
- Page title: `Plant P450s that have appeared since the 1993 P450 nomenclature update`
- Page last-modified statement: June 22, 2010
- Explicit plant/photosynthetic-lineage CYP/species headers accepted: **7,127**
- Accepted protein sequences: **0**

## Resource-specific parsing rule

The Word-generated HTML is decoded as Windows-1252 and normalized into source lines. A new bibliography record begins only when a cleaned line starts with an explicit `CYP` designation followed by at least two spaces and a species phrase. Its block ends immediately before the next such header. Every record preserves the source line, original species text, CYP designation, and complete cleaned bibliography block.

Species names are normalized only to stable selector labels: ordinary binomials use their first genus/species pair; `Genus sp. strain` and named hybrids retain the necessary third component; a small explicit mapping handles the page's common-name or genus-only labels. Four metadata-like headers (`confidential`, a discontinued rename, `Glyma02g31910.1 pseudogene`, and `Os06g0570600 BAF19826`) are rejected because they do not identify a species. Eight explicit non-plant/metadata entries (`Trypanosoma`, three `Leishmania`, `Monosiga`, two `Euglena`, and `This seq`) are also excluded from the plant database; they are not silently reclassified into a reviewed non-plant category.

## Sequence decision

This page mixes amino-acid strings, nucleotide accessions, genome coordinates, PCR fragments, ESTs, citations, old names, duplicates, and explicit non-P450 discussion. It is therefore accepted as an annotation/bibliography auxiliary source, not as a uniform FASTA source. All 7,127 records keep an empty protein `Sequence` field. Literal proteins remain available from independently reviewed sequence resources; no sequence is inferred, translated, repaired, or copied into these bibliography records.

Duplicate CYP names and historical variants remain separate in page order through block-specific `RecordKey` values. These records supplement existing sources and never overwrite them.
