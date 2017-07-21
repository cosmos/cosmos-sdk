package ibc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	wire "github.com/tendermint/go-wire"
)

const (
	// this is the prefix for the list of chains
	// we otherwise use the chainid as prefix, so this must not be an
	// alpha-numeric byte
	prefixChains = "**"

	prefixInput  = "i"
	prefixOutput = "o"
)

// this is used for the global handler info
var (
	handlerKey = []byte{0x2}
)

// HandlerInfo is the global state of the ibc.Handler
type HandlerInfo struct {
	Registrar basecoin.Actor `json:"registrar"`
}

// Save the HandlerInfo to the store
func (h HandlerInfo) Save(store state.KVStore) {
	b := wire.BinaryBytes(h)
	store.Set(handlerKey, b)
}

// LoadInfo loads the HandlerInfo from the data store
func LoadInfo(store state.KVStore) (h HandlerInfo) {
	b := store.Get(handlerKey)
	if len(b) > 0 {
		wire.ReadBinaryBytes(b, &h)
	}
	return
}

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

// Packet is a wrapped transaction and permission that we want to
// send off to another chain.
type Packet struct {
	DestChain   string          `json:"dest_chain"`
	Sequence    uint64          `json:"sequence"`
	Permissions basecoin.Actors `json:"permissions"`
	Tx          basecoin.Tx     `json:"tx"`
}

// NewPacket creates a new outgoing packet
func NewPacket(tx basecoin.Tx, dest string, seq uint64, perm ...basecoin.Actor) Packet {
	return Packet{
		DestChain:   dest,
		Sequence:    seq,
		Permissions: perm,
		Tx:          tx,
	}
}

// Bytes returns a serialization of the Packet
func (p Packet) Bytes() []byte {
	return wire.BinaryBytes(p)
}

// InputQueue returns the queue of input packets from this chain
func InputQueue(store state.KVStore, chainID string) *state.Queue {
	ch := stack.PrefixedStore(chainID, store)
	space := stack.PrefixedStore(prefixInput, ch)
	return state.NewQueue(space)
}

// OutputQueue returns the queue of output packets destined for this chain
func OutputQueue(store state.KVStore, chainID string) *state.Queue {
	ch := stack.PrefixedStore(chainID, store)
	space := stack.PrefixedStore(prefixOutput, ch)
	return state.NewQueue(space)
}
