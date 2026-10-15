# PGD schema

The database is bbolt. It contains a `records` bucket. Keys are stable zero-padded record numbers and values are JSON objects with these fields:

`ID`, optional `RecordKey`, `Category`, `Species`, `Symbol`, `Description`, `Sequence`, and `SourceURL`.

`ID` and `Symbol` preserve the biological/source identifier shown to users. `RecordKey` is a stable row-level key used when one resource contains several independently reviewed blocks with the same `ID`; it prevents FASTA lookup from resolving a duplicate name to the wrong block. Runtime readers fall back to `ID` for older records without `RecordKey`.

Empty source fields remain empty. `Category` is one of `animals`, `plants`, `fungi`, `bacteria`, or `protists`. This format preserves the maximum information available from the source materials without requiring the application to access Dr Nelson online during search.

Plant release records are admitted only from reviewed resource-specific CSV inputs under `sources/reviewed/plants/`. The legacy normalized-text scanner is not a release input.

Bacteria release records are admitted only from reviewed resource-specific CSV inputs under `sources/reviewed/bacteria/`. The reviewed set includes the 2000, 2003, 2009 and 2011 historical collections, environmental samples, and the partial collection. The taxonomic abundance document is an index only and intentionally contributes zero records. Each record retains its source block and source hash; environmental samples use source sample labels instead of guessed species.

Animals release records are admitted only from reviewed resource-specific CSV inputs under `sources/reviewed/animals/`. Each of the 40 page resources has a dedicated parser or an explicitly fixed workbook/Word profile, a source hash, physical row/block provenance, and representative tests. Literal protein preserves internal stops, ambiguity characters, source gaps and case; only an explicit terminal stop marker is removed. Empty pseudogene/placeholder blocks remain empty, and DNA/EST is not translated. Aggregate public-insect headers are assigned to a species only through the resource's reviewed complete-binomial allowlist; ambiguous and abbreviated headers retain the collection label.

The optional `species` bucket contains JSON objects with `name`, `category`, `selectable`, and `description`. It lists every discovered species, including species without searchable records; those entries use `selectable: false` and remain visible but disabled in PHgo.
