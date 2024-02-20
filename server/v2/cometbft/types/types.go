package types

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	ics23 "github.com/cosmos/ics23/go"
	"google.golang.org/protobuf/proto"

	corecomet "cosmossdk.io/core/comet"
	"cosmossdk.io/server/v2/core/store"
)

type VoteExtensionsHandler interface {
	ExtendVote(context.Context, *abci.RequestExtendVote) (*abci.ResponseExtendVote, error)
	VerifyVoteExtension(context.Context, *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error)
}

type Store interface {
	LatestVersion() (uint64, error)
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (store.ReaderMap, error)

	// StateCommit commits the provided changeset and returns
	// the new state root of the state.
	StateCommit(changes []store.StateChanges) (store.Hash, error)

	Query(storeKey string, version uint64, key []byte, prove bool) (QueryResult, error)
}

type QueryResult interface {
	Key() []byte
	Value() []byte
	Version() uint64
	Proof() *ics23.CommitmentProof
	ProofType() string
}

// PeerFilter responds to p2p filtering queries from Tendermint
type PeerFilter func(info string) (*abci.ResponseQuery, error)

type ConsensusInfo struct { // TODO: this is a mock, we need a proper proto.Message
	proto.Message
	corecomet.Info
}
