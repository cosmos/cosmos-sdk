package rootmulti

import "cosmossdk.io/store/v2"

var (
	_ store.RootStore            = (*Store)(nil)
	_ store.UpgradeableRootStore = (*Store)(nil)
)

// Store defines a multi-tree implementation variant of a RootStore .It contains
// a single State Storage (SS) backend and a State Commitment (SC) backend per module,
// i.e. store key. This implementation is meant to be congruent with the store
// v1 RootMultiStore and support existing application that DO NOT wish to migrate
// to the SDK's default single tree RootStore variant.
type Store struct{}
