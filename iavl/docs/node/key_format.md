# Key Format

Nodes, orphans, and roots are stored under the database with different key formats to ensure there are no key collisions and a structured key from which we can extract useful information.

### Nodes

Node KeyFormat: `n|<node.hash>`

Nodes are marshalled and stored under nodekey with prefix `n` to prevent collisions and then appended with the node's hash.

### Orphans

Orphan KeyFormat: `o|toVersion|fromVersion|hash`

Orphans are marshalled nodes stored with prefix `o` to prevent collisions. You can extract the toVersion, fromVersion and hash from the orphan key by using:

```golang
// orphanKey: o|50|30|0xABCD
var toVersion, fromVersion int64
var hash []byte
orphanKeyFormat.Scan(orphanKey, &toVersion, &fromVersion, hash)

/*
toVersion = 50
fromVersion = 30
hash = 0xABCD
*/
```

The order of the orphan KeyFormat matters. Since deleting a version `v` will delete all orphans whose `toVersion = v`, we can easily retrieve all orphans from nodeDb by iterating over the key prefix: `o|v`.

### Roots

Root KeyFormat: `r|<version>`

Root hash of the IAVL tree at version `v` is stored under the key `r|v` (prefixed with `r` to avoid collision).
