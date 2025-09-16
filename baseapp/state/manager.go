package state

import (
	"fmt"
	"sync"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Manager manages the different execution states of the BaseApp.
// It maintains separate states for different ABCI execution modes and provides
// thread-safe access to these states.
type Manager struct {
	// volatile states:
	//
	// - checkState is set on InitChain and reset on Commit
	// - finalizeBlockState is set on InitChain and FinalizeBlock and set to nil
	// on Commit.
	//
	// - checkState: Used for CheckTx, which is set based on the previous block's
	// state. This state is never committed.
	//
	// - prepareProposalState: Used for PrepareProposal, which is set based on the
	// previous block's state. This state is never committed. In case of multiple
	// consensus rounds, the state is always reset to the previous block's state.
	//
	// - processProposalState: Used for ProcessProposal, which is set based on the
	// previous block's state. This state is never committed. In case of
	// multiple rounds, the state is always reset to the previous block's state.
	//
	// - finalizeBlockState: Used for FinalizeBlock, which is set based on the
	// previous block's state. This state is committed.
	checkState           *State       // State for CheckTx execution
	prepareProposalState *State       // State for PrepareProposal execution
	processProposalState *State       // State for ProcessProposal execution
	finalizeBlockState   *State       // State for FinalizeBlock execution
	stateMut             sync.RWMutex // Protects concurrent access to states

	gasConfig config.GasConfig // Gas configuration for transaction execution
}

// NewManager creates a new state manager with the given gas configuration.
func NewManager(gasConfig config.GasConfig) *Manager {
	return &Manager{
		gasConfig: gasConfig,
	}
}

// GetState returns the state for the specified execution mode.
// Returns nil if no state has been set for the given mode.
func (mgr *Manager) GetState(mode sdk.ExecMode) *State {
	mgr.stateMut.RLock()
	defer mgr.stateMut.RUnlock()

	switch mode {
	case sdk.ExecModeFinalize:
		return mgr.finalizeBlockState

	case sdk.ExecModePrepareProposal:
		return mgr.prepareProposalState

	case sdk.ExecModeProcessProposal:
		return mgr.processProposalState

	default:
		return mgr.checkState
	}
}

// SetState initializes the BaseApp's state for the specified execution mode.
// It creates a branched multi-store (CacheMultiStore) and a new Context with
// the provided header information. The state is configured based on the execution mode.
func (mgr *Manager) SetState(
	mode sdk.ExecMode,
	unbranchedStore storetypes.CommitMultiStore,
	h cmtproto.Header,
	logger log.Logger,
	streamingManager storetypes.StreamingManager,
) {
	// Create a cached multi-store for the state
	ms := unbranchedStore.CacheMultiStore()
	// Extract header information for the context
	headerInfo := header.Info{
		Height:  h.Height,
		Time:    h.Time,
		ChainID: h.ChainID,
		AppHash: h.AppHash,
	}
	// Create a new state with the branched store and context
	baseState := NewState(
		sdk.NewContext(ms, h, false, logger).
			WithStreamingManager(streamingManager).
			WithHeaderInfo(headerInfo),
		ms,
	)

	mgr.stateMut.Lock()
	defer mgr.stateMut.Unlock()

	// Configure and assign the state based on execution mode
	switch mode {
	case sdk.ExecModeCheck:
		// CheckTx mode requires special gas price configuration
		baseState.SetContext(baseState.Context().WithIsCheckTx(true).WithMinGasPrices(mgr.gasConfig.MinGasPrices))
		mgr.checkState = baseState

	case sdk.ExecModePrepareProposal:
		mgr.prepareProposalState = baseState

	case sdk.ExecModeProcessProposal:
		mgr.processProposalState = baseState

	case sdk.ExecModeFinalize:
		mgr.finalizeBlockState = baseState

	default:
		panic(fmt.Sprintf("invalid runTxMode for setState: %d", mode))
	}
}

// ClearState removes the state for the specified execution mode.
// This is typically called when resetting or cleaning up states.
func (mgr *Manager) ClearState(mode sdk.ExecMode) {
	mgr.stateMut.Lock()
	defer mgr.stateMut.Unlock()

	switch mode {
	case sdk.ExecModeCheck:
		mgr.checkState = nil

	case sdk.ExecModePrepareProposal:
		mgr.prepareProposalState = nil

	case sdk.ExecModeProcessProposal:
		mgr.processProposalState = nil

	case sdk.ExecModeFinalize:
		mgr.finalizeBlockState = nil

	default:
		panic(fmt.Sprintf("invalid runTxMode for clearState: %d", mode))
	}
}
