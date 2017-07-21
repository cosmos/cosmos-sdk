package ibc

import (
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

const (
	// this is the prefix for the list of chains
	// we otherwise use the chainid as prefix, so this must not be an
	// alpha-numeric byte
	prefixChains = "**"

	prefixInput  = "i"
	prefixOutput = "o"
)

// HandlerKey is used for the global permission info
func HandlerKey() []byte {
	return []byte{0x2}
}

// ChainsKey is the key to get info on all chains
func ChainsKey() []byte {
	return stack.PrefixedKey(prefixChains, state.SetKey())
}

// ChainKey is the key to get info on one chain
func ChainKey(chainID string) []byte {
	bkey := state.MakeBKey([]byte(chainID))
	return stack.PrefixedKey(prefixChains, bkey)
}

// QueueInKey is the key to get newest of the input queue from this chain
func QueueInKey(chainID string) []byte {
	return stack.PrefixedKey(chainID,
		stack.PrefixedKey(prefixInput,
			state.QueueTailKey()))
}

// QueueOutKey is the key to get v of the output queue from this chain
func QueueOutKey(chainID string) []byte {
	return stack.PrefixedKey(chainID,
		stack.PrefixedKey(prefixOutput,
			state.QueueTailKey()))
}

// QueueInPacketKey is the key to get given packet from this chain's input queue
func QueueInPacketKey(chainID string, seq uint64) []byte {
	return stack.PrefixedKey(chainID,
		stack.PrefixedKey(prefixInput,
			state.QueueItemKey(seq)))
}

// QueueOutPacketKey is the key to get given packet from this chain's output queue
func QueueOutPacketKey(chainID string, seq uint64) []byte {
	return stack.PrefixedKey(chainID,
		stack.PrefixedKey(prefixOutput,
			state.QueueItemKey(seq)))
}
