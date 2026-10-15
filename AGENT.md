# P450 PGD agent instructions

## Plant rebuild authority

- The immediate release target is the complete `plants` section from the Dr. Nelson plants page. Work through the page resources one at a time in page/manifest order.
- Current checkpoint (2026-10-14): the rebuilt PGD and generated static site contain 28,907 reviewed records across 415 category/species labels. Plants contribute 28,892 records across 405 species labels and 22,485 literal protein sequences. The historical David R. Nelson `biblioD.html` resource contributes 7,134 provenance-distinct auxiliary records and 1,870 literal protein sequences; 7,119 records remain plants and 15 records across 10 species are classified as protists instead of being mixed into plants, animals, fungi, or bacteria. biblioD remains an auxiliary source and its web result numbers are red. Oryza sativa contributes 1,639 records, including 487 biblioD records, 341 Rice P450 Phylogenomics protein records, 412 annotation-workbook rows, 15 Dr. Nelson FASTA records, and 384 RAP-DB IRGSP-1.0 records. Dr. Nelson and Rice P450 Phylogenomics remain the primary rice sources; RAP-DB is the first auxiliary/supplementary rice source. The Rice P450 Phylogenomics cDNA and genomic files remain companion URLs and never enter the protein Sequence field. Static category controls are enabled for animals, plants, fungi, bacteria, and protists; a category receives species only from reviewed records, so the current biblioD expansion populates plants and protists while the other non-plant source sets remain pending independent review.
- Bacteria checkpoint (2026-10-15): all seven resources on the Dr. Nelson bacteria page have been independently reviewed. Six sequence-bearing resources contribute 3,027 provenance-distinct records and 2,959 literal source proteins across 268 selectable labels; the taxonomic abundance document is a reviewed index and contributes zero placeholder records. Historical 2000/2003/2009/2011 releases remain source-distinct rather than being silently deduplicated. Environmental records retain seawater or acid-mine sample labels instead of guessed organisms. The rebuilt database contains 31,934 total records and the static site contains 683 category/species detail shards; PGD SHA-256 is `8d8a625f4a7190b1e75fd25d741a9063b684c52d7622be501cf059a89e5ced96`.
- Animals checkpoint (2026-10-15, complete): all 40 resources on the Dr. Nelson Animals page have independent resource-specific parsers or fixed workbook/Word profiles, tests, reviewed CSVs, source hashes and ledgers. They contribute 5,408 provenance-distinct records, 5,368 literal source proteins and 113 selectable labels. Forty explicit source rows/blocks remain sequence-empty; they are never filled externally. The public insect collection uses only reviewed complete binomials printed in exact headers for species assignment, leaving abbreviated or ambiguous headers under the aggregate label. Candidate sheets, alternate orderings, superseded collections, Word prose, DNA/EST translation and comparison material are excluded. Together with Plants, Protists and Bacteria, the completed build contains 37,342 records across 796 category/species labels and produces 796 static detail shards; the runtime schema is unchanged.
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
- Static detail actions are table-scoped: Download Table produces a real OOXML `.xlsx`, Print prints only the detail table, and Close only calls `window.close()` for the current detail tab/window; it must never navigate backward or redirect. Browser printing on the search page, including Ctrl+P, prints only the current results table and no surrounding page content.

- Keep the bbolt runtime fields compatible with `phytozome-go/internal/cyp`: `ID`, `Category`, `Species`, `Symbol`, `Description`, `Sequence`, and `SourceURL`, plus the existing `species` bucket contract.
- If richer source metadata is needed, evolve the schema and runtime reader together with backward-compatible tests; do not overload existing fields with ambiguous values.
- Update the main runtime repository's `AGENT.md` CYP status whenever the published plant coverage or database contract materially changes.

## Repository handling

- Existing uncommitted work may be modified; do not discard unrelated user changes.
- Raw downloads and normalized intermediates may remain ignored, but every published result must be reproducible from the manifest, resource-specific parser/profile, and documented source hash.
