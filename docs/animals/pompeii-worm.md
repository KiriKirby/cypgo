# Animal resource review: Alvinella pompejana (Pompeii worm)

- Source file: `animals-POMPEII.worm.ESTs.xlsx`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/POMPEII.worm.ESTs.xlsx
- SHA-256: `24d4c66e9490b2a91260ea6b86e8ae4c5a8768bbbe52f87d6e10a7f3e97abc52`
- Workbook sheets inspected: 15
- Authoritative sheet: `All clans unique`
- Exact accepted rows: `2–187`
- Protein column: `K (protein seq)`
- Used range: `A1:K194`
- Hidden rows: `[]`
- Hidden columns: `[]`
- Merged cells: `[]`
- Accepted records: `186`
- Literal proteins: `186`
- Review status: `complete`

## Resource-specific interpretation

The workbook contains BLAST hit tables, per-clan candidate lists, two alternate clan orderings, a consensus-only sheet and the final `All clans unique` relationship table. Rows 2–187 of that table are the authoritative 186 protein-bearing rows; rows 188–190 are blank/count separators and rows 191–194 are explicitly labelled false positives without protein. Protein is taken only from column K, never translated from nucleotide column J. The 26 source consensus rows with blank GI are retained because the final table supplies explicit consensus protein and CYP-labelled hit; they receive stable row-based IDs. Candidate/clan sheets and the separate consensus sheet are not appended as duplicates. Internal stops, X/x, fragments and consensus status remain literal; only terminal stops are removed.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 28 |
| internal-stop-retained | 7 |
| short-source-fragment | 179 |
| source-consensus-row | 26 |
| source-partial-or-consensus | 27 |
| terminal-stop-removed | 6 |
