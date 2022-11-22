# Mutable Tree

### Structure

The MutableTree struct is a wrapper around ImmutableTree to allow for updates that get stored in successive versions.

The MutableTree stores the last saved ImmutableTree and the current working tree in its struct while all other saved, available versions are accessible from the nodeDB.

```golang
// MutableTree is a persistent tree which keeps track of versions.
type MutableTree struct {
	*ImmutableTree                  // The current, working tree.
	lastSaved      *ImmutableTree   // The most recently saved tree.
	orphans        map[string]int64 // Nodes removed by changes to working tree.
	versions       map[int64]bool   // The previous, saved versions of the tree.
	ndb            *nodeDB
}
```

The versions map stores which versions of the IAVL are stored in nodeDB. Anytime a version `v` gets persisted, `versions[v]` is set to `true`. Anytime a version gets deleted, `versions[v]` is set to false.

### Set

Set can be used to add a new key-value pair to the IAVL tree, or to update an existing key with a new value.

Set starts at the root of the IAVL tree, if the key is less than or equal to the root key, it recursively calls set on root's left child. Else, it recursively calls set on root's right child. It continues to recurse down the IAVL tree based on comparing the set key and the node key until it reaches a leaf node.

If the leaf node has the same key as the set key, then the set is just updating an existing key with a new value. The value is updated, and the old version of the node is orphaned.

If the leaf node does not have the same key as the set key, then the set is trying to add a new key to the IAVL tree. The leaf node is replaced by an inner node that has the original leaf node and the new node from the set call as its children.

If the `setKey` < `leafKey`:

```golang
// new leaf node that gets created by Set
// since this is a new update since latest saved version,
// this node has version=latestVersion+1
newVersion := latestVersion+1
newNode := NewNode(key, value, newVersion)
// original leaf node: originalLeaf gets replaced by inner node below
Node{
    key:       setKey,       // inner node key is equal to left child's key
    height:    1,            // height=1 since node is parent of leaves
    size:      2,            // 2 leaf nodes under this node
    leftNode:  newNode,      // left Node is the new added leaf node
    rightNode: originalLeaf, // right Node is the original leaf node
    version:   newVersion,   // inner node is new since lastVersion, so it has an incremented version
}
```

If `setKey` > `leafKey`:

```golang
// new leaf node that gets created by Set
// since this is a new update since latest saved version,
// this node has version=latestVersion+1
newVersion := latestVersion+1
newNode := NewNode(key, value, latestVersion+1)
// original leaf node: originalLeaf gets replaced by inner node below
Node{
    key:       leafKey,      // inner node key is equal to left child's key
    height:    1,            // height=1 since node is parent of leaves
    size:      2,            // 2 leaf nodes under this node
    leftNode:  originalLeaf, // left Node is the original leaf node
    rightNode: newNode,      // right Node  is the new added leaf node
    version:   newVersion,   // inner node is new since lastVersion, so it has an incremented version
}
```

Any node that gets recursed upon during a Set call is necessarily orphaned since it will either have a new value (in the case of an update) or it will have a new descendant. The recursive calls accumulate a list of orphans as it descends down the IAVL tree. This list of orphans is ultimately added to the mutable tree's orphan list at the end of the Set call.

After each set, the current working tree has its height and size recalculated. If the height of the left branch and right branch of the working tree differs by more than one, then the mutable tree has to be balanced before the Set call can return.

### Remove

Remove is another recursive function to remove a key-value pair from the IAVL pair. If the key that is trying to be removed does not exist, Remove is a no-op.

Remove recurses down the IAVL tree in the same way that Set does until it reaches a leaf node. If the leaf node's key is equal to the remove key, the node is removed, and all of its parents are recursively updated. If not, the remove call does nothing.

#### Recursive Remove

Remove works by calling an inner function `recursiveRemove` that returns the following values after a recursive call `recursiveRemove(recurseNode, removeKey)`:

##### NewHash

If a node in recurseNode's subtree gets removed, then the hash of the recurseNode will change. Thus during the recursive calls down the subtree, all of recurseNode's children will return their new hashes after the remove (if they have changed). Using this information, recurseNode can calculate its own updated hash and return that value.

If recurseNode is the node getting removed itself, NewHash is `nil`.

##### ReplaceNode

Just like with recursiveSet, any node that gets recursed upon (in a successful remove) will get orphaned since its hash must be updated and the nodes are immutable. Thus, ReplaceNode is the new node that replaces `recurseNode`.

If recurseNode is the leaf that gets removed, then ReplaceNode is `nil`.

If recurseNode is the direct parent of the leaf that got removed, then it can simply be replaced by the other child. Since the parent of recurseNode can directly refer to recurseNode's remaining child. For example if recurseNode's left child gets removed, the following happens:


Before LeftLeaf removed:
```
                        |---RightLeaf
IAVLTREE---recurseNode--|
                        |---LeftLeaf
```

After LeftLeaf removed:
```
IAVLTREE---RightLeaf

ReplaceNode = RightLeaf
orphaned = [LeftLeaf, recurseNode]
```

If recurseNode is an inner node that got called in the recursiveRemove, but is not a direct parent of the removed leaf. Then an updated version of the node will exist in the tree. Notably, it will have an incremented version, a new hash (as explained in the `NewHash` section), and recalculated height and size.

The ReplaceNode will be a cloned version of `recurseNode` with an incremented version. The hash will be updated given the NewHash of recurseNode's left child or right child (depending on which branch got recurse upon).

The height and size of the ReplaceNode will have to be calculated since these values can change after the `remove`.

It's possible that the subtree for `ReplaceNode` will have to be rebalanced (see `Balance` section). If this is the case, this will also update `ReplaceNode`'s hash since the structure of `ReplaceNode`'s subtree will change.

##### LeftmostLeafKey

The LeftmostLeafKey is the key of the leftmost leaf of `recurseNode`'s subtree. This is only used if `recurseNode` is the right child of its parent. Since inner nodes should have their key equal to the leftmost key of their right branch (if leftmostkey is not `nil`). If recurseNode is the right child of its parent `parentNode`, `parentNode` will set its key to `parentNode.key = leftMostKeyOfRecurseNodeSubTree`.

If `recurseNode` is a leaf, it will return `nil`.

If `recurseNode` is a parent of the leaf that got removed, it will return its own key if the left child was removed. If the right child is removed, it will return `nil`.

If `recurseNode` is a generic inner node that isn't a direct parent of the removed node, it will return the leftmost key of its child's recursive call if `node.key < removeKey`. It will return `nil` otherwise.

If `removeKey` does not exist in the IAVL tree, leftMostKey is `nil` for entire recursive stack.

##### RemovedValue

RemovedValue is the value that was at the node that was removed. It does not get changed as it travels up the recursive stack.

If `removeKey` does not exist in the IAVL tree, RemovedValue is `nil`.

##### Orphans

Just like `recursiveSet`, any node that gets recursed upon by `recursiveRemove` in a successful `Remove` call will have to be orphaned. The Orphans list in `recursiveRemove` accumulates the list of orphans so that it can return them to `Remove`. `Remove` will then iterate through this list and add all the orphans to the mutable tree's `orphans` map.

If the `removeKey` does not exist in the IAVL tree, then the orphans list is `nil`.

### Balance

Anytime a node is unbalanced such that the height of its left branch and the height of its right branch differs by more than 1, the IAVL tree will rebalance itself.

This is acheived by rotating the subtrees until there is no more than one height difference between two branches of any subtree in the IAVL.

Since Balance is mutating the structure of the tree, any displaced nodes will be orphaned.

#### RotateRight

To rotate right on a node `rotatedNode`, we first orphan its left child. We clone the left child to create a new node `newNode`. We set `newNode`'s right hash and child to the `rotatedNode`. We now set `rotatedNode`'s left child to be the old right child of `newNode`.

Visualization (Nodes are numbered to show correct key order is still preserved):

Before `RotateRight(node8)`:
```
    |---9
8---|
    |       |---7
    |   |---6
    |   |   |---5
    |---4
        |   |---3
        |---2
            |---1
```

After `RotateRight(node8)`:
```
         |---9
     |---8
     |   |   |---7
     |   |---6
     |       |---5
4'---|
     |   |---1
     |---2
         |---3

Orphaned: 4
```

Note that the key order for subtrees is still preserved.

#### RotateLeft

Similarly, to rotate left on a node `rotatedNode` we first orphan its right child. We clone the right child to create a new node `newNode`. We set the `newNode`'s left hash and child to the `rotatedNode`. We then set the `rotatedNode`'s right child to be the old left child of the node.

Before `RotateLeft(node2)`:
```
            |---9
        |---8
        |   |---7
    |---6
    |   |   |---5
    |   |---4
    |       |---3
2---|
    |---1
```

After `RotateLeft(node2)`:
```
         |---9
     |---8
     |   |---7
6'---|
     |       |---5
     |   |---4
     |   |   |---3  
     |---2
         |---1

Orphaned: 6
```

The IAVL detects whenever a subtree has become unbalanced by 2 (after any set/remove). If this does happen, then the tree is immediately rebalanced. Thus, any unbalanced subtree can only exist in 4 states:

#### Left Left Case

1. `RotateRight(node8)`

**Before: Left Left Unbalanced**
```
    |---9
8---|
    |   |---6
    |---4
        |   |---3
        |---2
```

**After 1: Balanced**
```
         |---9
     |---8
     |   |---6
4'---|
     |   |---3
     |---2

Orphaned: 4
```

#### Left Right Case

Make tree left left unbalanced, and then balance.

1. `RotateLeft(node4)`
2. `RotateRight(node8)`

**Before: Left Right Unbalanced**
```
    |---9
8---|
    |   |---6
    |   |   |---5
    |---4
        |---2
```

**After 1: Left Left Unbalanced**
```
    |---9
8---|
    |---6'
        |   |---5
        |---4
            |---2

Orphaned: 6
```

**After 2: Balanced**
```
         |---9
     |---8
6'---|
     |   |---5
     |---4
         |---2

Orphaned: 6
```

Note: 6 got orphaned again, so omit list repitition

#### Right Right Case

1. `RotateLeft(node2)`

**Before: Right Right Unbalanced**
```
            |---9
        |---8
    |---6
    |   |---4
2---|
    |---1
```

**After: Balanced**
```
         |---9
     |---8
6'---|
     |   |---4
     |---2
         |---1

Orphaned: 6
```

#### Right Left Case

Make tree right right unbalanced, then balance.

1. `RotateRight(6)`
2. `RotateLeft(2)`

**Before: Right Left Unbalanced**
```
        |---8
    |---6
    |   |---4
    |       |---3
2---|
    |---1
```

**After 1: Right Right Unbalanced**
```
            |---8
        |---6
    |---4'
    |   |---3
2---|
    |---1

Orphaned: 4
```

**After 2: Balanced**
```
         |---8
     |---6
4'---|
     |   |---3
     |---2
         |---1

Orphaned: 4
```

### SaveVersion

SaveVersion saves the current working tree as the latest version, `tree.version+1`.

If the tree's root is empty, then there are no nodes to save. The `nodeDB` must still save any orphans since the root itself could be the node that has been removed since the last version. Then the `nodeDB` also saves the empty root for this version.

If the root is not empty. Then SaveVersion will ensure that the `nodeDB` saves the orphans, roots and any new nodes that have been created since the last version was saved.

SaveVersion also calls `nodeDB.Commit`, this ensures that any batched writes from the last save gets committed to the appropriate databases.

`tree.version` gets incremented and the versions map has `versions[tree.version] = true`.

It will set the lastSaved `ImmutableTree` to the current working tree, and clone the tree to allow for future updates on the next working tree. It also resets orphans to the empty map.

Lastly, it returns the tree's hash, the latest version, and nil for error.

SaveVersion will error if a tree at the version trying to be saved already exists.

### DeleteVersion

DeleteVersion will simply call nodeDB's `DeleteVersion` function which is documented in the [nodeDB docs](./nodedb.md) and then call `nodeDB.Commit` to flush all batched updates.

It will also delete the version from the versions map.

DeleteVersion will return an error if the version is invalid, or nonexistent. DeleteVersion will also return an error if the version trying to be deleted is the latest version of the IAVL tree since that is unallowed.
