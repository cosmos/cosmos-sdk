package state

import (
	"fmt"
	"sync"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	// the previous block's state. This state is never committed. In case of
	// multiple rounds, the state is always reset to the previous block's state.
	//
	// - finalizeBlockState: Used for FinalizeBlock, which is set based on the
	// previous block's state. This state is committed.
	checkState           *State
	prepareProposalState *State
	processProposalState *State
	finalizeBlockState   *State
	stateMut             sync.RWMutex

	gasConfig config.GasConfig
}

func NewManager(gasConfig config.GasConfig) *Manager {
	return &Manager{
		gasConfig: gasConfig,
	}
}

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

// SetState sets the BaseApp's state for the corresponding mode with a branched
// multi-store (i.e. a CacheMultiStore) and a new Context with the same
// multi-store branch, and provided header.
func (mgr *Manager) SetState(
	mode sdk.ExecMode,
	unbranchedStore storetypes.CommitMultiStore,
	h cmtproto.Header,
	logger log.Logger,
	streamingManager storetypes.StreamingManager,
) {
	ms := unbranchedStore.CacheMultiStore()
	headerInfo := header.Info{
		Height:  h.Height,
		Time:    h.Time,
		ChainID: h.ChainID,
		AppHash: h.AppHash,
	}
	baseState := NewState(
		sdk.NewContext(ms, h, false, logger).
			WithStreamingManager(streamingManager).
			WithHeaderInfo(headerInfo),
		ms,
	)

	mgr.stateMut.Lock()
	defer mgr.stateMut.Unlock()

	switch mode {
	case sdk.ExecModeCheck:
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
