package types

import (
	"context"

	coreheader "cosmossdk.io/core/header"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProtoApp is what we would need to implement in the new BaseApp.
type ProtoApp interface {
	ChainID() string
	Name() string
	Version() string

	InitialHeight() int64
	LastBlockHeight() int64
	AppVersion() (uint64, error)

	// TODO: Replace this with Marko's TX validation
	ValidateTX([]byte) (gInfo sdk.GasInfo, result *sdk.Result, anteEvents []abci.Event, err error) // TODO: I'm just replicating what runTx replies here

	// New interface
	InitChain(RequestInitChain) (ResponseInitChain, error)
	DeliverTxs(coreheader.Info, [][]byte) ([]interface{}, error)
	Commit() error

	// COMET BFT specific stuff below (tbd where to put them)
	Validators() []abci.ValidatorUpdate
	ConsensusParams() *tmtypes.ConsensusParams
	AppHash() []byte
	GetBlockRetentionHeight(commitHeight int64) int64
}

type HasProposal interface {
	PrepareProposal(context.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)
}

type RequestInitChain struct {
	StateBytes []byte
}

type ResponseInitChain struct{}
