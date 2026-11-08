# SearchForge (eSearch)

SearchForge is an Elasticsearch-inspired distributed search and indexing engine built from scratch in Go.

## Architecture Highlights

- **Custom Inverted Indexing**: Postings lists, term dictionaries, and positional indexing.
- **BM25 Ranking**: Okapi BM25 scoring model with configurable term saturation and length normalization.
- **Durable Storage**: Append-only Write-Ahead Log (WAL) with CRC32 framing and crash replay.
- **Immutable Segments**: Lucene-style disk segments with Variable-Byte compression and zero-copy reads.
- **Segment Compaction**: Background multi-way merge policies with soft-deletion tombstones.
- **Distributed Coordination**: Raft-based consensus, dynamic cluster membership, and failure detection.
- **Primary-Replica Sharding**: Sequence-numbered replication protocol with split-brain mitigation.
- **Scatter-Gather Search**: Distributed query execution across partitioned shards.

## Development Status

- Current Step: **S001** (Repository Initialization & Tooling Baseline)

