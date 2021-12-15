// This package provides concrete implementations of the store/v2 "MultiStore" types, including
// CommitMultiStore, CacheMultiStore, and BasicMultiStore (as read-only stores at past versions).
//
// Substores are declared as part of a schema within StoreOptions.
// The schema cannot be changed once a CommitMultiStore is initialized, and changes to the schema must be done
// by migrating via StoreOptions.Upgrades. If a past version is accessed, it will be loaded with the past schema.
// Stores may be declared as StoreTypePersistent, StoreTypeMemory (not persisted after close), or
// StoreTypeTransient (not persisted across commits). Non-persistent substores cannot be migrated or accessed
// in past versions.
//
// A declared persistent substore is initially empty and stores nothing in the backing DB until a value is set.
// A non-empty store is stored within a prefixed subdomain of the backing DB (using db/prefix).
// If the MultiStore is configured to use a separate DBConnection for StateCommitmentDB, it will store the
// state commitment (SC) store (as an SMT) in subdomains there, and the "flat" state is stored in the main DB.
// Each substore's SC is allocated as an independent SMT, and query proofs contain two components: a proof
// of a key's (non)existence within the substore SMT, and a proof of the substore's existence within the
// MultiStore (using the Merkle map proof spec (TendermintSpec)).

package root
