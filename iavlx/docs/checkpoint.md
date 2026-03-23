# Checkpoints

A checkpoint persists to disk, the set of IAVL tree nodes which were not present in a previous checkpoint.
You can think of it like a snapshot of the nodes in the tree at a given version, minus any nodes that
were already stored in a previous checkpoint.
So unlike a snapshot, a checkpoint does not contain all the nodes in the tree at that version, rather
only the new nodes which were not persisted on disk earlier.
