---
sidebar_position: 1
---

# Per-Block Application Logs

:::note Synopsis
A node can keep the application log lines emitted while executing each of the last N blocks and serve them over gRPC and REST, so an operator can ask "what did the application log during block 12345?" without grepping process output or a log aggregator.
:::

The feature is node-local and off by default. Enabling it never affects consensus or the app hash: the capture is a pure passthrough on top of the node's logger, reads and writes no application state, and swallows every error.

## Enabling

Set `retain-blocks` in the `[block-logs]` section of `app.toml`:

```toml
[block-logs]
# 0 (the default) disables capture.
retain-blocks = 100
```

The same value can be given as a start flag or an environment variable, following the usual precedence:

```shell
simd start --block-logs.retain-blocks=100
SIMD_BLOCK_LOGS_RETAIN_BLOCKS=100 simd start
```

Existing `app.toml` files are not rewritten when a node upgrades; add the section by hand or run `confix migrate`.

## What is captured

While a block executes, every log line emitted through the application's logger is recorded, **at every level, regardless of `log_level`**. Debug lines are captured even when the node prints at `info`. This covers `ctx.Logger()` and keepers that received their logger at construction (for example `x/bank`), because the wrapper is installed on the root logger handed to the application and `With(...)` keeps children captured.

The capture window is `FinalizeBlock` start to `Commit` end, so `BeginBlock`, transaction execution, `EndBlock`, the app hash computation and everything logged up to and including `Commit` belong to that height. When optimistic execution is enabled (`baseapp.SetOptimisticExecution()`, as in `simapp`), the block body already runs during `ProcessProposal`, so the window opens there instead; a proposal that is not the block finally decided is discarded and the height re-executed and re-captured from scratch.

Caveats:

* **`CheckTx` attribution.** CometBFT runs `CheckTx` on its mempool connection concurrently with block execution, and background goroutines share the same loggers. Lines they emit while a block is open are recorded under that block's height; lines they emit between blocks are dropped. `PrepareProposal` and `ProcessProposal` validation run outside the window and are not captured. Vote extension handlers run between `ProcessProposal` and `FinalizeBlock`, so they are captured only when optimistic execution is enabled.
* **CometBFT is excluded.** Only the logger handed to the application is wrapped. Consensus, p2p and mempool lines stay in the process log.
* Querying over REST requires the gRPC server (`[grpc] enable = true`, the default): the REST route is proxied to it and the query is intentionally not available over ABCI Query.

## Disk cost

Entries are written to `<home>/data/block-logs/<height>.jsonl`, one file per height, and the file for height `H - retain-blocks` is deleted when height `H` commits. The first commit after a restart sweeps every file at or below the cutoff, so lowering `retain-blocks` frees space at the next block.

Because every level is recorded, the disk cost is roughly the **debug-level** log volume of the application for the last `retain-blocks` blocks, which can be far more than what the process prints at `info`. Start with a small value and measure the directory.

A file that already exists when its height starts executing is truncated: a height runs at most once per process, so a pre-existing file is a partial write from a crashed run that CometBFT is now replaying.

## Querying

The query is the `Logs` method of `cosmos.base.node.v1beta1.LogsService`, exposed on the node's gRPC server and as `GET /cosmos/base/node/v1beta1/logs` on the REST API.

| Parameter           | Meaning                                                                                                             |
| ------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `start`, `end`      | Height range, inclusive. `0` means the last captured height. `start` is clipped up to `end - retain-blocks + 1`. |
| `level`             | Minimum level, inclusive: `debug` < `info` < `warn` < `error`. Defaults to `info`. Unknown levels are an error.     |
| `module`            | Exact match on the `module` field. Entries without a module are excluded when set.                                 |
| `pagination.limit`  | Page size, default 100, maximum 1000 (larger values are clamped).                                                   |
| `pagination.key`    | Opaque cursor from the previous response's `next_key`.                                                              |

`pagination.offset`, `pagination.count_total` and `pagination.reverse` are rejected. Errors are returned for `start > end`, `start < 1`, `end` beyond the last captured height, and when capture is disabled. Malformed lines left by a crash are skipped. Files are streamed, so memory use is bounded by the page size, and the query never holds the writer's lock while scanning.

```shell
curl -s 'http://localhost:1317/cosmos/base/node/v1beta1/logs?start=120&end=121&level=debug&module=x/bank&pagination.limit=2'
```

```json
{
  "entries": [
    {
      "height": "120",
      "time": "2026-09-03T17:41:02.113402Z",
      "level": "debug",
      "module": "x/bank",
      "msg": "minted coins from module account",
      "fields": {
        "amount": "1233stake",
        "from": "mint"
      }
    },
    {
      "height": "121",
      "time": "2026-09-03T17:41:07.220119Z",
      "level": "debug",
      "module": "x/bank",
      "msg": "minted coins from module account",
      "fields": {
        "amount": "1233stake",
        "from": "mint"
      }
    }
  ],
  "pagination": {
    "next_key": "AAAAAAAAAHkAAAAAAAAB9A==",
    "total": "0"
  }
}
```

Pass `next_key` back as `pagination.key` to continue; an empty `next_key` means the range is exhausted. `time` is the wall clock at emission, not the block time.
