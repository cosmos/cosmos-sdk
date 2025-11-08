Framed Gzip WAL Writer (low I/O)

Overview
- Writes append-only `.gz` files composed of multiple completed gzip members (one per flush).
- Maintains a sidecar `.idx` file with byte offsets and basic metadata per frame.
- Rotates files by size and/or time to keep objects bounded and immutable for downstream shippers.

Why
- Avoids tailing an "open gzip member" which is hard for agents to consume safely.
- Minimizes disk I/O by writing compressed bytes directly and a tiny index line per flush.

Files
- `<prefix>-<UTC ts>-<roll>.gz`   : concatenated gzip members, append-only until rotation.
- `<prefix>-<UTC ts>-<roll>.idx`  : newline-delimited JSON, one entry per frame.

Index line schema (JSON)
{
  "file": "<filename.gz>",
  "frame": <uint64>,
  "off": <uint64>,         // start offset in .gz
  "len": <uint64>,         // byte length in .gz
  "recs": <uint32>,        // record count in this frame (optional semantics)
  "first_ts": <int64>,     // unix nanos
  "last_ts": <int64>,      // unix nanos
  "crc32": <uint32>        // crc32 of uncompressed payload (optional)
}

Agent consumption
- Watch the `.idx` file for new lines; each line indicates a completed frame ready to ship.
- Use `off` and `len` to perform a single `pread`/range read on the `.gz` and forward bytes as-is.
- Acknowledge up to the last shipped `{file, frame}` for at-least-once delivery; dedupe by `{file, frame}` or application `event_id`.

Rotation
- Configure either/both: max size (e.g., 64 MiB) and max interval (e.g., 30s).
- On rotation, close current `.gz` and `.idx`, create new pair; agent naturally continues with the new `.idx`.

Notes
- Most off-the-shelf shippers (Fluent Bit/Filebeat/Vector) do not tail growing `.gz` files. If you require continuous shipping from a single growing `.gz`, implement a lightweight custom agent using the provided Reader skeleton.
- If you rotate to immutable `.gz` files frequently, some shippers can pick up completed files (e.g., via directory scan + external uploader), but support varies.

Agent example
- Build: `go build ./cmd/memagent`
- Run (verify mode prints decompressed line counts):
  - `./memagent -root /path/to/app -node <nodeID> -verify -once`
  - Tails the latest `.wal.idx`, reads frames from `.wal.gz`, prints metadata + counts.
