package types

import (
	"context"
	"time"

	"cosmossdk.io/store/snapshots"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProtoApp is what we would need to implement in the new BaseApp.
type ProtoApp interface {
	ChainID() string
	Name() string
	Version() string

	InitialHeight() int64
	SetInitialHeight(int64)

	SnapshotManager() *snapshots.Manager
	CommitMultiStore() storetypes.CommitMultiStore
	StreamingManager() storetypes.StreamingManager

	// TODO: Should these be a CometBFT specific thing?
	MinGasPrices() sdk.DecCoins
	CheckHalt(height int64, time time.Time) error
	GetBlockRetentionHeight(commitHeight int64) int64

	// TODO: figure out if these below here are going to be available
	QueryMultiStore() storetypes.MultiStore
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
	AppVersion(ctx context.Context) (uint64, error)

	// TODO: Define what methods the Cosmos SDK ABCI will have
	ValidateTX([]byte) (gInfo sdk.GasInfo, result *sdk.Result, anteEvents []abci.Event, err error) // TODO: I'm just replicating what runTx replies here

	InitChainer() sdk.InitChainer                       // InitChainer should not have Comet types
	PreBlock(req *abci.RequestFinalizeBlock) error      // TODO: Should preblock be only a Comet thing?
	BeginBlock(context.Context) (sdk.BeginBlock, error) // TODO: create a new response type for this, we might not need it at all as we can access the EventManager()
	EndBlock(context.Context) (sdk.EndBlock, error)     // TODO: sdk.EndBlock should not have Comet types
	DeliverTx([]byte) *abci.ExecTxResult                // TODO: *abci.ExecTxResult should be an SDK type instead
}

type HasProposal interface {
	PrepareProposal(context.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error)
	ProcessProposal(context.Context, *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error)
}
