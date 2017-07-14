package ibc

import (
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	wire "github.com/tendermint/go-wire"
)

const (
	// this is the prefix for the list of chains
	// we otherwise use the chainid as prefix, so this must not be an
	// alpha-numeric byte
	prefixChains = "**"
)

// ChainInfo is the global info we store for each registered chain,
// besides the headers, proofs, and packets
type ChainInfo struct {
	RegisteredAt uint64 `json:"registered_at"`
	RemoteBlock  int    `json:"remote_block"`
}

// ChainSet is the set of all registered chains
type ChainSet struct {
	*state.Set
}

// NewChainSet loads or initialized the ChainSet
func NewChainSet(store state.KVStore) ChainSet {
	space := stack.PrefixedStore(prefixChains, store)
	return ChainSet{
		Set: state.NewSet(space),
	}
}

// Register adds the named chain with some info
// returns error if already present
func (c ChainSet) Register(chainID string, ourHeight uint64, theirHeight int) error {
	if c.Exists([]byte(chainID)) {
		return ErrAlreadyRegistered(chainID)
	}
	info := ChainInfo{
		RegisteredAt: ourHeight,
		RemoteBlock:  theirHeight,
	}
	data := wire.BinaryBytes(info)
	c.Set.Set([]byte(chainID), data)
	return nil
}
