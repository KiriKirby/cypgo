# Resource audit: biblioD historical bibliography

- Source URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/biblioD.html
- Source file: `plants-biblioD.html`
- Source SHA-256: `aac540d08652c2241b7c500f1b7e2c751ab94a8d5e98dd9e04a954cf48252c7b`
- Explicit supported-organism CYP/species headers accepted: **7,134**
- Plant records: **7,119**
- Protist records: **15**
- Accepted literal protein sequences: **1,870** (1,863 plants; 7 protists)

## Resource-specific parsing rule

The Word-generated HTML is decoded as Windows-1252 and normalized into source lines. A new bibliography record begins only when a cleaned line starts with an explicit `CYP` designation followed by at least two spaces and an organism phrase. Its block ends immediately before the next such header. Every record preserves the source line, original organism text, CYP designation, and complete cleaned bibliography block.

Species names are normalized only to stable selector labels. Metadata-like headers that do not identify an organism are rejected. Explicit non-plant headers are retained as `protists` when the source identifies protist or non-green-algal organisms; they are never mixed into plants, animals, fungi, or bacteria.

## Sequence decision

The page is heterogeneous and is not parsed as one uniform FASTA file. A protein is accepted only from contiguous uppercase amino-acid lines inside the current CYP block. Optional source coordinate prefixes/suffixes are removed, literal internal gap/stop characters are retained, and only explicit terminal `*` markers are removed. ACGTN-only runs are rejected as nucleotide. No sequence is inferred, translated, repaired, or copied from another database.

Duplicate CYP names and historical variants remain separate in page order through block-specific `RecordKey` values. These records supplement existing sources and never overwrite them. Because biblioD is an auxiliary source, its search-result numbers are red even though the file is hosted on the David R. Nelson domain.
