# Animal resource review: Aedes aegypti

- Authoritative named collection: `animals-AedesFASTA.June25.doc`
- Historical URL (redirects to the named DOC): https://drnelson.uthsc.edu/AedesFasta.June25.htm
- Named DOC SHA-256: `d54b27a8e282a3a2505b21f241c6d97b2119cf6ff95168973be880fa898ce086`
- Named normalized text: `animals-AedesFASTA.June25.txt`
- Named text SHA-256: `4e6159a93d6d63ada9535a41c7aa5cbab6ca18d7d80f2dc28a0a1c3b565f6fce`
- Named text lines: `1522`
- Archival reconstruction document: `animals-aedesaegypti.doc`
- Archival URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/aedesaegypti.doc
- Archival DOC SHA-256: `6a5f50716ff8d33458679c2bae37bd61eae951357f5eed590cfc3c9f5e946b1f`
- Archival text SHA-256: `a3a0a64cf4f83bb4a53db159f77e479659b13532c069fbbf172e26cb39d09eee`
- Explicit named blocks: `167`
- Literal proteins: `167`
- Missing sequences: `0`
- Review status: `complete`

## Resource-specific interpretation

The archival document explicitly directs readers to the 25 June 2006 completed-and-named FASTA collection and describes itself as detailed archival material. The named collection is therefore the release input; the 206 unfinished archival candidates, BLAST Query/Sbjct alignments and assembly evidence are not added again. Although its title says 164 sequences (11 CYP2 + 84 CYP3 + 59 CYP4 + 10 mitochondrial), the named document physically contains 167 explicit `>` blocks. All physical named blocks remain source-distinct, including versions, alleles and partials, instead of silently discarding three records to match the prose total. Only literal uppercase protein lines are joined after removing coordinates and intron-phase annotations. Mixed prose/alignment lines are rejected. Terminal stops are removed; internal stops, X and source gaps remain. No DNA is translated and no external sequence is substituted.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 1 |
| internal-stop-retained | 3 |
| short-source-fragment | 5 |
| source-allele | 6 |
| source-missing-exon | 1 |
| source-partial | 1 |
| source-pseudogene | 4 |
| terminal-stop-removed | 89 |
