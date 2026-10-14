# P450 PGD agent instructions

## Plant rebuild authority

- The immediate release target is the complete `plants` section from the Dr. Nelson plants page. Work through the page resources one at a time in page/manifest order.
- Current checkpoint (2026-10-14): the rebuilt PGD and generated static site contain 28,900 reviewed plant records across 409 species labels, including 20,622 literal protein sequences. The historical David R. Nelson `biblioD.html` bibliography contributes 7,127 provenance-distinct auxiliary annotation records and no accepted protein sequences; its mixed citations, ESTs, fragments, coordinates, old names, and occasional literal sequence text are retained in descriptions rather than interpreted as uniform FASTA. Explicit non-plant entries in that page remain excluded from the plant database. Oryza sativa contributes 1,639 records, including 487 biblioD bibliography entries in addition to the 341 Rice P450 Phylogenomics protein records, 412 annotation-workbook rows, 15 Dr. Nelson FASTA records, and 384 RAP-DB IRGSP-1.0 records. Dr. Nelson and Rice P450 Phylogenomics remain the primary rice sources; RAP-DB is the first auxiliary/supplementary rice source. The Rice P450 Phylogenomics cDNA (341 records) and genomic (302 records) remain companion URLs and never enter the protein Sequence field. The plant release gate is satisfied; non-plant resources remain disabled until independently reviewed.
- Download and inspect every plant resource independently. A resource is not approved merely because it has been Office-normalized or mentioned in an audit document.
- Do not publish plant records produced by a category-wide CYP regex scan, printable-byte scan, guessed delimiter, guessed FASTA boundary, or a generic "sequence-looking token" rule.
- Every approved plant resource must have its own parser/profile implementation or a reviewed structured CSV whose columns and row relationships were checked against that exact source file.
- Excel resources must be inspected at workbook, worksheet, header, merged-cell, hidden-row/column, and representative-row level before defining their parser. Sequence data may be accepted only from the exact reviewed sheet/column or cell layout for that workbook.
- Word/text resources must be inspected as complete logical blocks. Define the exact header, annotation, protein-sequence, terminator, fragment, pseudogene, duplicate, and foreign-species rules for that document. Never skip arbitrary intervening text to assemble a sequence.
- Preserve source row order and source provenance. Keep duplicate/alternate resources distinct unless a documented merge rule proves that two rows describe the same biological record.
- A protein sequence must be literal source data. Do not infer, repair, translate, or obtain it from another database. Preserve documented fragment/pseudogene status; remove only an explicit terminal stop marker from the stored amino-acid sequence.
- Each accepted record must keep the source URL and a resource-specific description sufficient to locate the source row/block. Empty fields remain empty.

## Review ledger and release gate

- Maintain a durable per-resource review under `docs/plants/` and a machine-readable plant completion ledger under `sources/`.
- Each completed resource records: source file and URL, source hash, worksheets/sections inspected, exact parser/profile, total accepted records, records with sequence, missing-sequence count, duplicates/alternates, fragments/pseudogenes, foreign-species rows, and unresolved issues.
- A plant species/resource is selectable only after its exact parser or reviewed CSV passes tests and its audit is marked complete. Unreviewed resources remain present but disabled.
- The formal `cypgo.pgd` plant records must be built only from completed resource-specific inputs. The legacy generic `parseExtracted` path must not contribute publishable plant records.
- Tests must include representative first/middle/last records and every exceptional layout found in the source, plus sequence alphabet/length checks and record/sequence-count assertions.
- After every completed resource, rebuild to a temporary PGD, inspect its records, and compare its counts and representative sequences with the source before marking the resource complete.

## Runtime compatibility

- Static CYPgo detail pages display every available protein sequence as a PHgo Lite FASTA record using `>species|ID(symbol)`. Header whitespace is normalized to underscores, reserved Lite separators are replaced safely, empty symbols omit the parenthetical suffix, and sequence-missing records keep the reserved detail row empty.
- Static search results keep David R. Nelson row numbers in the normal color. Every result from another source uses a bold red row number, and the result-count area groups those red row numbers by their human-readable source name.

- Keep the bbolt runtime fields compatible with `phytozome-go/internal/cyp`: `ID`, `Category`, `Species`, `Symbol`, `Description`, `Sequence`, and `SourceURL`, plus the existing `species` bucket contract.
- If richer source metadata is needed, evolve the schema and runtime reader together with backward-compatible tests; do not overload existing fields with ambiguous values.
- Update the main runtime repository's `AGENT.md` CYP status whenever the published plant coverage or database contract materially changes.

## Repository handling

- Existing uncommitted work may be modified; do not discard unrelated user changes.
- Raw downloads and normalized intermediates may remain ignored, but every published result must be reproducible from the manifest, resource-specific parser/profile, and documented source hash.
