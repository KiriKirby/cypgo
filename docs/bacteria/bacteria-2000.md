# Bacteria resource review: over 130 bacterial cytochrome P450 sequences (July 27, 2000)

- Source file: `bacteria-bacteria.2000.txt`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacteria.2000.doc
- Source SHA-256: `0ef41b98f5a1bc779a186d9c604c290eb7b93cb9d71778cf040a2ebb5b308ade`
- Container: UTF-16LE Word-normalized text, 1,563 logical lines
- Explicit `>` blocks retained: `157`
- Blocks with literal protein sequence: `150`
- Missing-sequence blocks: `7`
- Distinct explicit organism labels: `46`
- Review status: `complete`

## Resource-specific interpretation

The source is parsed as 157 complete logical blocks beginning at a literal `>` line. Every block is retained in source order, including accession-only records, repeated CYP symbols, aliases, fragments, partials, confidential entries, and the final CYP? record. The first whitespace token is the source symbol; accession tokens are retained as IDs when explicitly present. Organism labels are copied only from an explicit binomial or `sp.` phrase in the same header; accession/CYP inference is not used, and unstated organisms use the explicit `Bacterial collection (organism not stated)` label.

Protein is accepted only from uppercase residue lines inside that exact block. Word wrapping and surrounding spaces are removed; terminal quote characters and explicit terminal `*` markers are removed, while internal `*`, X, gaps, and unusual residues are retained. The numbered reverse-orientation L76374 block removes only its printed end coordinates. CYP118 has two separately annotated fragments; only its first contiguous fragment is stored and the later fragment is not concatenated across prose. Empty blocks remain records with an empty sequence. No sequence is translated, repaired, deduplicated, or fetched externally.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X | 2 |
| c-terminal-fragment | 2 |
| literal-protein | 150 |
| multiple-sequence-runs-unmerged | 1 |
| n-terminal-fragment | 2 |
| reviewed-bacteria-2000 | 157 |
| same-gene-alternate | 3 |
| sequence-missing | 7 |
| short-fragment | 11 |
| source-confidential | 3 |
| source-fragment | 1 |
| source-partial | 3 |
| source-pseudogene | 1 |

## Representative and exceptional blocks

| Block | Line | Symbol | Species | Length | Status |
|---:|---:|---|---|---:|---|
| 1 | 2 | CYP51 | Bacterial collection (organism not stated) | 450 aa | reviewed-bacteria-2000;literal-protein |
| 31 | 304 | CYP105M1 | Streptomyces clavuligerus | 0 aa | reviewed-bacteria-2000;same-gene-alternate;sequence-missing |
| 45 | 436 | CYP107L1 | Streptomyces venezuelae | 0 aa | reviewed-bacteria-2000;sequence-missing |
| 69 | 677 | CYP115P | Bacterial collection (organism not stated) | 0 aa | reviewed-bacteria-2000;source-pseudogene;sequence-missing |
| 73 | 715 | CYP118 | Bacterial collection (organism not stated) | 66 aa | reviewed-bacteria-2000;source-fragment;literal-protein;multiple-sequence-runs-unmerged;ambiguous-X;short-fragment |
| 85 | 835 | CYP127A2 | Rhizobium sp | 0 aa | reviewed-bacteria-2000;source-confidential;sequence-missing |
| 93 | 904 | CYP133A1 | Erwinia herbicola | 0 aa | reviewed-bacteria-2000;source-confidential;sequence-missing |
| 118 | 1172 | CYP153A1 | Acinetobacter calcoaceticus | 0 aa | reviewed-bacteria-2000;source-confidential;sequence-missing |
| 144 | 1435 | CYP167A1 | Sorangium cellulosum | 0 aa | reviewed-bacteria-2000;same-gene-alternate;sequence-missing |
| 152 | 1525 | L76374 | Mycobacterium avium | 161 aa | reviewed-bacteria-2000;literal-protein;ambiguous-X;short-fragment |
| 157 | 1555 | CYP? | Streptomyces noursei | 144 aa | reviewed-bacteria-2000;source-partial;literal-protein;short-fragment |
