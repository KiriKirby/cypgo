# CYPgo static plant search

## Scope

The static site is generated from the reviewed `cypgo.pgd` release database. The category row supports animals, plants, fungi, bacteria, and protists; Plants remains the default. Each category has its own generated search payload, and records are never mixed across categories.

The existing `index.html` and `detailtmp.html` files are the visual templates. The generated site keeps their 760-pixel layout, colors, images, footer, and table proportions.

## Search behavior

- The selector to the left of the query contains every species in the selected category that has at least one searchable record.
- **Search** restricts matches to the selected species.
- **Cross-Species Search** searches all species in the selected category.
- Matching is case-insensitive against the displayed CYP name. The displayed CYP name is `Symbol`, falling back to `ID` when `Symbol` is empty.
- The complete trimmed query is attempted first. If it has no matches, one Unicode character is removed from the right and the shorter query is attempted. This repeats until a non-empty result set is found or the query becomes empty.
- Once a query length matches, every record whose CYP name contains that successful query is returned in database order.
- Before a search, only the introductory image area is displayed. After a search attempt, only the results area is displayed. An unsuccessful attempt shows an empty-result message instead of switching back to the introductory image.

## Static data layout

`go run ./cmd/build-static -db cypgo.pgd -out .` regenerates all browser data:

- `static-data/<category>-search.js` contains the species list and compact search rows for that category.
- `static-data/details/<category>-<species-slug>.js` contains full records for one category/species only.
- `static-data/build-manifest.json` records per-category and total counts plus a deterministic source database SHA-256.

Search rows contain only the detail ID, displayed CYP name, species, detail shard, and the row offset inside that shard. The detail page therefore loads one species shard on demand instead of downloading every protein sequence on the home page.

The detail page reserves rows for every runtime database field: CYP Name, ID, Record Key, Category, Species, Description, Sequence, and Source URL. The Sequence row contains the available FASTA text using the compact `>species|ID(symbol)` header format; fields are whitespace-normalized and separator-safe, and the symbol suffix is omitted when empty. Missing values leave their row empty. Its browser title and visible heading append the current CYP name.

## Layout and wrapping rules

All result and detail tables use fixed table layout. Cells may wrap long words and protein sequences, row height grows with content, and column proportions remain fixed:

- results: `10% / 45% / 45%`;
- details: `30% / 70%`.

## Verification

The generator rejects a missing bucket, malformed record, unsupported category, duplicate generated detail ID, unsafe output path component, or a search/detail count mismatch. Unit tests cover deterministic generation, category filtering, species sharding, name fallback, escaping, and manifest counts.


