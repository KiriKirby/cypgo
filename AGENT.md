# P450 PGD agent instructions

## Plant rebuild authority

- The immediate release target is the complete `plants` section from the Dr. Nelson plants page. Work through the page resources one at a time in page/manifest order.
- Current checkpoint (2026-10-13): all plant resources are complete in manifest order through Red Algae. The published PGD and generated audit contain 20,636 reviewed plant records across 70 species, including 19,897 literal sequences. The plant release gate is satisfied; non-plant resources remain disabled until independently reviewed. User screenshots and literature tables are test fixtures only and are not PGD sources.
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

- Keep the bbolt runtime fields compatible with `phytozome-go/internal/cyp`: `ID`, `Category`, `Species`, `Symbol`, `Description`, `Sequence`, and `SourceURL`, plus the existing `species` bucket contract.
- If richer source metadata is needed, evolve the schema and runtime reader together with backward-compatible tests; do not overload existing fields with ambiguous values.
- Update the main runtime repository's `AGENT.md` CYP status whenever the published plant coverage or database contract materially changes.

## Repository handling

- Existing uncommitted work may be modified; do not discard unrelated user changes.
- Raw downloads and normalized intermediates may remain ignored, but every published result must be reproducible from the manifest, resource-specific parser/profile, and documented source hash.
