package state

// ChainState maintains general information for the chain
type ChainState struct {
	chainID string
}

// NewChainState creates a blank state
func NewChainState() *ChainState {
	return &ChainState{}
}

var baseChainIDKey = []byte("base/chain_id")

// SetChainID stores the chain id in the store
func (s *ChainState) SetChainID(store KVStore, chainID string) {
	s.chainID = chainID
	store.Set(baseChainIDKey, []byte(chainID))
}

// GetChainID gets the chain id from the cache or the store
func (s *ChainState) GetChainID(store KVStore) string {
	if s.chainID != "" {
		return s.chainID
	}
	s.chainID = string(store.Get(baseChainIDKey))
	return s.chainID
}
