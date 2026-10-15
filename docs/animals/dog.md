# Animal resource review: Canis familiaris (dog)

- Naming/annotation document: `animals-dogP450s.doc`
- Document URL: https://drnelson.uthsc.edu/wp-content/uploads/sites/130/resources/dogP450s.doc
- DOC SHA-256: `92f49ee9bc0309ed11645aa74a341114d4124635127493372b919c1de1d45d9a`
- Newest-collection FASTA live URL recorded in the Word hyperlink: http://drnelson.uthsc.edu/BLAST/dog.dbs.html
- Archived capture used because the live path now returns 404: https://web.archive.org/web/20110817181055id_/http://drnelson.uthsc.edu/BLAST/dog.dbs.html
- Reviewed archived file: `animals-dog.dbs.20110817.html`
- FASTA SHA-256: `3afe07a441fb26ffe14adcdb8e8ea50dfb4d150c4267c5676bf2cbb3db607306`
- Archived source lines: `595`
- Explicit FASTA records: `56`
- Literal protein sequences: `56`
- Review status: `complete`

## Resource-specific interpretation

The Word document explicitly labels itself modified January 8, 2010 and links `Dog FASTA file` to `/BLAST/dog.dbs.html` as its newest collection. That link is now dead, but the Internet Archive preserves the exact 2011 capture. The archived FASTA is therefore the authoritative publishable row set. Its 56 explicit `>` blocks are parsed independently and kept in source order, including pseudogenes, internal-exon fragments, deletion fragments, internal stops, lowercase residues and ambiguous X/x. Only an explicit terminal `*` is removed. The older Word body was inspected as annotation/candidate material (101 headers, including family aggregation and multi-run blocks) and is retained as companion provenance; it is not appended as duplicate/intermediate records over the source-designated newest FASTA. No external sequence database is queried or used.

## Status counts

| Status | Records |
|---|---:|
| ambiguous-X-retained | 3 |
| internal-stop-retained | 7 |
| short-source-fragment | 9 |
| source-lowercase-retained | 3 |
| terminal-stop-removed | 14 |
