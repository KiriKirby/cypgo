# Bacteria resource review: Bacterial Cytochrome P450s (partial collection)

- Source document: `bacteria-bacterial.P450s.doc`
- URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/bacterial.P450s.doc
- DOC SHA-256: `2b174b909711d87a3bb30177cd7bdcb74ca4ee737ad8e3cc12fd2f61e380d0bf`
- Reviewed normalized text: `bacteria-bacterial.P450s.txt`
- Normalized text SHA-256: `fd803abf2ae94cef5667156fc154d7abf7c01bf2f5cc8fdf40768be572490b25`
- Accepted independent source sections: `30`
- Accepted literal protein sequences: `30`
- Excluded BLAST Query/Sbjct alignment lines: 18 Query + 25 Sbjct lines
- Excluded foreign fungal section: `X82490 Fusarium oxysporum`
- Excluded BLAST-only `C. crescentus at TIGR` section: 4 HSPs
- Review status: `complete`

## Resource-specific interpretation

This UTF-16LE Word-normalized partial collection is not a FASTA file. Each accepted record is a manually delimited source section with an accession/header, optional gene annotation, and a contiguous literal protein range. The three-gene AJ223999, two-gene AJ223998, two-gene AF040571, and three-gene AF040570 sections are emitted as independent records with separate source IDs and line ranges. AJ223012 is a BLAST Sbjct alignment only and is excluded. The C. crescentus section contains only BLAST Query/Sbjct HSPs and is excluded. X82490 is explicitly a fungal P450 and is excluded from the bacterial category. CYP118 has separately annotated fragment pieces; only the first contiguous fragment is retained, with X markers preserved. Terminal quotes/stops are removed only at the sequence boundary; no alignment, translation, repair, deduplication, or external lookup is used.

| Record | Symbol | Species | Length | Source lines | Status |
|---:|---|---|---:|---|---|
| 1 | CYP127 | Rhizobium sp. NGR234 | 414 aa | 9-16 |  |
| 2 | AF071144 | Streptomyces glaucescens | 120 aa | 20-22 | short-fragment |
| 3 | AF071146 | Micromonospora inyoensis | 120 aa | 26-28 | short-fragment |
| 4 | AF071147 | Streptoalloteichus hindustanus | 120 aa | 32-34 | short-fragment |
| 5 | AF071148 | Amycolata autotrophica | 120 aa | 38-40 | short-fragment |
| 6 | CYP116 | Rhodococcus erythropolis | 437 aa | 50-57 |  |
| 7 | CYP109B1 | Nocardioides sp. | 397 aa | 61-68 |  |
| 8 | CYP103A2 | Agrobacterium tumefaciens | 419 aa | 73-80 |  |
| 9 | CYP165C2 | Amycolatopsis orientalis | 310 aa | 85-90 |  |
| 10 | CYP165C3-gene1 | Amycolatopsis orientalis | 451 aa | 95-103 | n-terminal-fragment |
| 11 | CYP107H2-gene2 | Amycolatopsis orientalis | 397 aa | 105-112 |  |
| 12 | CYP165A2-gene1 | Amycolatopsis orientalis | 391 aa | 118-124 |  |
| 13 | CYP165B2-gene2 | Amycolatopsis orientalis | 406 aa | 128-135 |  |
| 14 | CYP165C3-gene3 | Amycolatopsis orientalis | 151 aa | 140-142 | n-terminal-fragment |
| 15 | CYP107Q1-gene1 | Amycolatopsis mediterranei | 511 aa | 147-156 |  |
| 16 | CYP105C1-gene2 | Amycolatopsis mediterranei | 351 aa | 159-165 | short-fragment |
| 17 | CYP105C1-gene1 | Amycolatopsis mediterranei | 396 aa | 182-189 |  |
| 18 | CYP105D1-gene2 | Amycolatopsis mediterranei | 403 aa | 194-201 |  |
| 19 | CYP107Q2-gene3 | Amycolatopsis mediterranei | 421 aa | 205-212 |  |
| 20 | CYP107M1 | Actinomadura hibisca | 411 aa | 216-223 |  |
| 21 | CYP140 | Mycobacterium tuberculosis | 438 aa | 227-234 |  |
| 22 | CYP144 | Mycobacterium tuberculosis | 434 aa | 238-245 |  |
| 23 | CYP143 | Mycobacterium tuberculosis | 393 aa | 249-255 |  |
| 24 | CYP142 | Mycobacterium tuberculosis | 398 aa | 259-266 |  |
| 25 | CYP135B1 | Mycobacterium tuberculosis | 472 aa | 270-278 |  |
| 26 | CYP? | Mycobacterium leprae | 310 aa | 282-287 |  |
| 27 | CYP136 | Mycobacterium tuberculosis | 492 aa | 293-301 |  |
| 28 | CYP? | Mycobacterium tuberculosis | 400 aa | 308-315 |  |
| 29 | CYP118 | Bacillus subtilis | 66 aa | 321-322 | source-fragment;ambiguous-X;multiple-sequence-runs-unmerged;short-fragment |
| 30 | CYP102A2 | Bacillus subtilis | 479 aa | 333-341 | n-terminal-fragment |
