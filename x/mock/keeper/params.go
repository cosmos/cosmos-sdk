package keeper

// MockParamKeeper of the global paramstore
type MockParamKeeper struct {
	cdc       *codec.Codec
	key       sdk.StoreKey
	tkey      sdk.StoreKey
	codespace sdk.CodespaceType
	spaces    map[string]*Subspace
}

// Allocate subspace used for keepers
func (k MockParamKeeper) Subspace(s string) Subspace {
	_, ok := k.spaces[s]
	if ok {
		panic("subspace already occupied")
	}

	if s == "" {
		panic("cannot use empty string for subspace")
	}

	space := subspace.NewSubspace(k.cdc, k.key, k.tkey, s)
	k.spaces[s] = &space

	return space
}

// Get existing substore from keeper
func (k MockParamKeeper) GetSubspace(s string) (Subspace, bool) {
	space, ok := k.spaces[s]
	if !ok {
		return Subspace{}, false
	}
	return *space, ok
}
