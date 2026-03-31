# iavlx Performance

iavlx was rigorously benchmarked against iavl/v1, memiavl, and iavl/v2 during its development to inform and
validate its design.

## Commit Performance

## Read Performance Analysis

iavlx read performance turns out to be the biggest factor in determining overall commit performance.
In iavlx's commit lifecycle, WAL writing is sequential and usually completes before the CPU intensive activities
of hashing and tree modification complete. Checkpoint writing is optimistic and usually not done every version
so it is also usually not a bottleneck. Tree modification, however, requires reading the existing nodes in the tree
to know what to change. Whether or not these nodes can be read from memory or need to be read from disk introduces
the biggest differential in performance.

The following benchmark numbers attempted to analyze reads/second for trees with different numbers of leaf nodes
and the read performance depending on whether the read was from an in-memory node, a memory-mapped node file in the
OS cache, or directly from storage. It was impossible to precisely control OS cache vs disk storage behavior,
but by doing aggressive OS cache flushing we were able to get a rough estimate.

| leaf nodes | mem (reads/s) | mmap/OS cache (reads/s) | disk (reads/s) |
|------------|--------|-----------------|--------|
| 1E+06 | 360,645 | 194,054 | 25,377 |
| 1E+07 | 224,442 | 66,080 | 3,758 |
| 2E+07 | 198,024 | 48,522 | 1,705 |
| 3E+07 | 175,662 | 43,128 | 1,268 |
| 4E+07 | 169,647 | 38,334 | 1,086 |
| 5E+07 | 161,664 | 35,851 | 951 |
| 6E+07 | 157,534 | 35,073 | 957 |
| 7E+07 | 155,951 | 32,801 | 807 |
| 8E+07 | 153,978 | 32,546 | 824 |
| 9E+07 | 162,010 | 31,231 | 752 |
| 1E+08 | 151,432 | 30,268 | 771 |

As you can see, as we increase the number of leaf nodes in the tree, read performance generally goes down. This is because we need to do more node-to-node traversals to get to our target node because the tree is deeper.
The read performance of tree with a larger branching factor (such as a B-tree)
would degrade more slowly because the depth of the tree would increase more slowly.
When we are reading from disk, we need to do random disk IO which is much slower than sequential disk IO. When iavlx is writing changeset files, all disk IO is sequential so it is not a bottleneck. But when we need to do lots of random disk IO, then performance degrades quite substantially. As you can see, when mmap pages are likely in the OS cache, they're not that much slower than direct memory reads. But when we hit disk, performance goes way down.

iavlx's disk format has many optimizations to reduce disk read overhead:
* O(1) offsets within a changeset
* O(log log n) interpolation search between changesets
* inlining of the first 8 bytes of a key in node structs

Inlining of the first 8 bytes of a key was benchmarked specifically after the above benchmarks were taken and resulted in generally 2x better read performance than before it was added:

| leaf nodes | disk (reads/s) | disk w/ inline key prefix (reads/s) |
|------------|----------------|--------------------------------------|
| 1E+06 | 25,377 | 23,357 |
| 1E+07 | 3,758 | 7,099 |
| 2E+07 | 1,705 | 6,055 |
| 3E+07 | 1,268 | 1,915 |
| 4E+07 | 1,086 | 2,071 |
| 5E+07 | 951 | 2,050 |
| 6E+07 | 957 | 1,873 |
| 7E+07 | 807 | 1,793 |
| 8E+07 | 824 | 1,688 |
| 9E+07 | 752 | 1,451 |
| 1E+08 | 771 | 1,550 |

## Multi-threaded Read Performance

## Comparison with other IAVL implementations

When iavlx was designed, memiavl already existed as a production alternative, and iavl/v2 had been built, but not
completed. iavl/v1 is designed to work around a generic key-value store, usually LevelDB. iavl/v2 was based on
Sqlite and memiavl, like iavlx, had its own custom storage engine.

In early benchmarks, memiavl generally showed superior performance to other alternatives except when tree size
became rather large, at which point snapshot writing started to take so long that it blocked normal block processing
and the node consumed so much memory that it crashes. Basically, memiavl's disk persistence mechanism is an
all or nothing snapshot. This works well when the tree is small but it doesn't scale.
iavlx takes inspiration from memiavl's design but uses partial checkpoints (inspired by iavl/v2) so that it
doesn't have this scaling problem. On top of that, iavlx makes other concurrency optimizations which allow it to
generally perform better than memiavl.