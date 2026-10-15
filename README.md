# cypgo.pgd data repository

This repository publishes the read-only `cypgo.pgd` bbolt database consumed by phytozome GO's CYP keyword connector. The application never builds this database from the web pages.

The database is being rebuilt category by category, starting with every resource on the Dr. Nelson plants page. Each plant file is downloaded and interpreted independently; a generic CYP/FASTA scan is not accepted as release data. Completed resource parsers write reviewed CSV files under `sources/reviewed/plants/`, with source hashes and per-block provenance.

Bacteria is now reviewed resource-by-resource as well. The release includes dedicated parsers and provenance CSVs for the 2000, 2003, 2009 and 2011 collections, environmental samples, and the partial collection. The taxonomic abundance page was audited as an index and contributes no placeholder records. Bacteria currently contributes 3,027 records, including 2,959 literal source proteins; no external sequence database is used to fill missing values.

Animals is also complete for all 40 resources listed on the Dr. Nelson Animals page. Fixed workbook profiles and complete Word-block reviews contribute 5,408 provenance-distinct records across 113 selectable labels, including 5,368 literal source proteins. The 40 sequence-empty rows remain searchable because their source blocks are explicit placeholders, pseudogenes, fragments, or otherwise contain no literal protein. Public-insect records are assigned to a species only when the complete binomial is printed in that exact header; ambiguous or abbreviated headers retain the `Public insect collection` label. DNA and EST text is never translated and no external sequence is substituted.

The application downloads the release asset on first CYP selection and stores it beside the executable as `cypgo.pgd`. Set `PHGO_CYPGO_URL` during testing to point at a release asset or raw file. The runtime still accepts the former `PHGO_CYP_PGD_URL` name for migration compatibility.

The generator and source-analysis documentation belong here, alongside the generated PGD. Keep the generated database immutable between releases and publish a checksum/manifest with each update. Species whose resource has not completed its resource-specific review remain listed but disabled.

The current GitHub release asset is consumed through `manifest.json`; PHgo verifies both its byte length and SHA-256 before atomically replacing the installed database.
