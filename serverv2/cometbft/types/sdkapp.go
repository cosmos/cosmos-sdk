package types

import (
	"context"

	"cosmossdk.io/store/snapshots"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProtoApp is what we would need to implement in the new BaseApp.
type ProtoApp interface {
	ABCI // Temporary, to be replaced by our own "ABCI" as baseapp should not
	// implement Comet's ABCI interface.

	ChainID() string
	Name() string
	Version() string

	InitialHeight() int64
	SetInitialHeight(int64)

	SnapshotManager() *snapshots.Manager
	CommitMultiStore() storetypes.CommitMultiStore
	StreamingManager() storetypes.StreamingManager

	// TODO: Should this be a CometBFT specific thing?
	MinGasPrices() sdk.DecCoins

	// TODO: figure out if these below here are going to be available
	QueryMultiStore() storetypes.MultiStore
	CreateQueryContext(height int64, prove bool) (sdk.Context, error)
	AppVersion(ctx context.Context) (uint64, error)

	// TODO: Define what methods the Cosmos SDK ABCI will have
	InitChainer() sdk.InitChainer // InitChainer should not have Comet types
	BeginBlock() error
	EndBlock() error
	DeliverTx() error
}
