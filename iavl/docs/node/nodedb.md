# NodeDB

### Structure

The nodeDB is responsible for persisting nodes, orphans, and roots correctly in persistent storage.

### Saving Versions

The nodeDB saves the roothash of the IAVL tree under the key: `r|<version>`.

It marshals and saves any new node that has been created under: `n|<hash>`. For more details on how the node gets marshaled, see [node documentation](./node.md). Any old node that is still part of the latest IAVL tree will not get rewritten. Instead its parent will simply have a hash pointer with which the nodeDB can retrieve the old node if necessary.

Any old nodes that were part of the previous version IAVL but are no longer part of this one have been saved in an orphan map `orphan.hash => orphan.version`. This map will get passed into the nodeDB's `SaveVersion` function. The map maps from the orphan's hash to the version that it was added to the IAVL tree. The nodeDB iterates through this map and stores each marshalled orphan node under the key: `o|toVersion|fromVersion`. Since the toVersion is always the previous version (if we are saving version `v`, toVersion of all new orphans is `v-1`), we can save the orphans by iterating over the map and saving: `o|(latestVersion-1)|orphan.fromVersion => orphan.hash`.

(For more details on key formats see the [keyformat docs](./key_format.md))

### Deleting Versions

When a version `v` is deleted, the roothash corresponding to version `v` is deleted from nodeDB. All orphans whose `toVersion = v`, will get the `toVersion` pushed back to the highest predecessor of `v` that still exists in nodeDB. If the `toVersion <= fromVersion` then this implies that there does not exist a version of the IAVL tree in the nodeDB that still contains this node. Thus, it can be safely deleted and uncached.

##### Deleting Orphans

The deleteOrphans algorithm is shown below:

```golang
// deleteOrphans deletes orphaned nodes from disk, and the associated orphan
// entries.
func (ndb *nodeDB) deleteOrphans(version int64) {
	// Will be zero if there is no previous version.
	predecessor := ndb.getPreviousVersion(version)

	// Traverse orphans with a lifetime ending at the version specified.
	ndb.traverseOrphansVersion(version, func(key, hash []byte) {
		var fromVersion, toVersion int64

		// See comment on `orphanKeyFmt`. Note that here, `version` and
		// `toVersion` are always equal.
		orphanKeyFormat.Scan(key, &toVersion, &fromVersion)

		// Delete orphan key and reverse-lookup key.
		ndb.batch.Delete(key)

		// If there is no predecessor, or the predecessor is earlier than the
		// beginning of the lifetime (ie: negative lifetime), or the lifetime
		// spans a single version and that version is the one being deleted, we
		// can delete the orphan.  Otherwise, we shorten its lifetime, by
		// moving its endpoint to the previous version.
		if predecessor < fromVersion || fromVersion == toVersion {
			ndb.batch.Delete(ndb.nodeKey(hash))
			ndb.uncacheNode(hash)
		} else {
			ndb.saveOrphan(hash, fromVersion, predecessor)
		}
	})
}
```
