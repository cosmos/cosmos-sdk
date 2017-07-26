package state

// ChainState maintains general information for the chain
type ChainState struct {
	chainID string
}

// NewChainState creates a blank state
func NewChainState() *ChainState {
	return &ChainState{}
}

// SetChainID stores the chain id in the store
func (s *ChainState) SetChainID(store KVStore, chainID string) {
	s.chainID = chainID
	store.Set([]byte("base/chain_id"), []byte(chainID))
}

// GetChainID gets the chain id from the cache or the store
func (s *ChainState) GetChainID(store KVStore) string {
	if s.chainID != "" {
		return s.chainID
	}
	s.chainID = string(store.Get([]byte("base/chain_id")))
	return s.chainID
}
