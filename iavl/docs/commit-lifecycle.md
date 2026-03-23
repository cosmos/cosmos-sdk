# Commit Lifecycle

IAVLX makes every attempt possible to reduce commit latency.
There are three things that need to happen when the tree is updated and committed:
1. update the tree
2. compute the root hash
3. write changes to disk

Significant portions of these operations can actually be decoupled so that work can proceed in parallel.

A critical insight around this work is that the IAVL tree is only updated after `FinalizeBlock` when
all mutations have settled into a cache layer which is held in memory as a B-tree.
For in-memory operations, mutations to this B-tree layer are generally more performant than performing
the mutations on an in-memory IAVL tree. More importantly, we must ensure that all IAVL tree operations
nodes are performed in exactly the same order on every node because the insertion order determines the tree
structure and root hash due to tree balancing.

Therefore, when we commit an IAVL tree we start with the list of all update operations from the B-tree cache layer
and apply these in sorted order so that insertion order is deterministic.
Because a B-tree is already a sorted data structure, there is nothing to do here - the data is already sorted.

Now, with this list of update operations from the B-tree, we can actually start all three potentially blocking operations
at the same time: updating the tree, hashing, and writing to disk.
(Also, since each IAVL tree in a multi-tree functions more or less independently, each tree is updated, hashed, and 
flushed to disk in parallel in separate go routines.)

Each set/update operation from the B-tree will actually result in a new leaf node in the next version of the tree.
Before the root hash is computed, we need to hash all the leaf nodes anyway, so with we can start computing
leaf node hashes from this list of updates before we have even started mutating the IAVL tree.
So a go routine with the new leaf nodes is starts right away to compute the leaf nodes.
Because each leaf node can be hashed independently, we can actually do this hashing in parallel batches
to get it done even more quickly.

Now, regarding writing to disk, we are also going to persist all of this leaf node to disk anyway, so we can start
writing it to disk right away as well in an append-only log.
This append-only log tracks which tracks all the update operations we will apply to the tree as well as most of the
leaf node data (key-value pairs) is called the write-ahead log (WAL).
It is written before we have mutated the tree, computed the root hash, or even committed to this set of changes, and
as we will see later, it allows us to roll back commits.

So before the IAVL tree is even mutated at all, we have started two background threads: one to compute leaf hashes (in parallel batches) and
another one to write the leaf data to disk as the WAL.
While these two goroutines are running, we apply each update to the in-memory tree one by one.
Because rebalancing must occur from the root, each update must be applied sequentially, and there is no possibility of
parallelizing this part of the work.
This update operation can be one of the most costly operations overall because as the tree grows,
we are able to keep less and less of it in memory, and update operations require more and more random disk reads.
We will get into this more later, but for large trees, the speed of applying updates becomes more and more bound
by the cost of reading nodes from disk.

Once we have finished applying all updates to the tree sequentially and once all leaf node hashes have completed,
we can start computing the root hash of the tree.
Note that we do not need to wait for the WAL to finish writing for this step to start.
Now, unlike mutating the tree root, hashing the root can happen in parallel because at this point the tree root is stable.
For any branch node, the hashes of the left and right sub-trees can be computed in parallel.

Once root hash and WAL writing have completed, we are almost ready to return from commit - the only thing we are waiting
for is confirmation from the caller that this commit will not be rolled back.
Up to this point we actually haven't committed to any changes. All the hard work is done, but if we want to,
we can simply truncate the WAL back to its previous size and the commit will be rolled back.
If the caller confirms that the commit will be finalized, then we do the following:
* for each tree:
  * update the in-memory version and root node pointer to reference the new tree
  * spin off a go routine to write a checkpoint to disk if needed
* at a multi-tree level:
  * write a `CommitInfo` file which durably finalizes the commits across all trees 

At this point, we can return to the caller and processing of the next block can proceed.
Now notice that before we returned to the caller we did not ensure that any data besides
the WAL and `CommitInfo` was written to disk.
The presence of a multi-tree `CommitInfo` and WAL for each tree is sufficient to ensure tree durability,
because we have everything we need to reconstruct tree nodes and hashes from this data alone.
But, of course, we do also want to persist node data to disk, because otherwise every restart would
require replaying all updates, which would be very inefficient.
Node data is persisted to disk at configurable checkpoint intervals, but checkpoints are always written
optimistically after the commit has actually completed.
If checkpoint writing fails for any reason, the system can detect the error at startup, clean up any garbage
data on disk, and write a new checkpoint.
Checkpoint writing can sometimes take a long time, especially with larger trees, so by default
checkpoints are not written every block, and it is okay if checkpoint writing takes more than one block to complete.
Because checkpoints are written in a background go routine, they are non-blocking, except in pathological cases where
the channel-send buffer (hard-coded to size 32 for now) backs up.
In such a scenario where 32 checkpoints are waiting to complete, the channel send buffer creates back-pressure
to slow things down a bit so checkpointing can catch up.
In a well-configured system, this should never happen unless checkpoints are scheduled to write every block.