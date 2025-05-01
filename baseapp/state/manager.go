package state

import (
	"fmt"
	"sync"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/header"
	"cosmossdk.io/log"
	"cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Manager struct {
	mut                  sync.RWMutex
	checkState           *State
	prepareProposalState *State
	processProposalState *State
	finalizeBlockState   *State
}

func (m *Manager) GetState(mode sdk.ExecMode) *State {
	m.mut.RLock()
	defer m.mut.RUnlock()

	switch mode {
	case sdk.ExecModeFinalize:
		return m.finalizeBlockState

	case sdk.ExecModePrepareProposal:
		return m.prepareProposalState

	case sdk.ExecModeProcessProposal:
		return m.processProposalState

	default:
		return m.checkState
	}
}

func (m *Manager) SetState(mode sdk.ExecMode, state *State) {
	m.mut.Lock()
	defer m.mut.Unlock()

	switch mode {
	case sdk.ExecModeCheck:
		m.checkState = state

	case sdk.ExecModePrepareProposal:
		m.prepareProposalState = state

	case sdk.ExecModeProcessProposal:
		m.processProposalState = state

	case sdk.ExecModeFinalize:
		m.finalizeBlockState = state

	default:
		panic(fmt.Sprintf("invalid runTxMode for clearState: %d", mode))
	}
}

func (m *Manager) ResetState(mode sdk.ExecMode, ms types.CacheMultiStore, h cmtproto.Header, logger log.Logger, sm types.StreamingManager, minGasPrices sdk.DecCoins) {
	headerInfo := header.Info{
		Height:  h.Height,
		Time:    h.Time,
		ChainID: h.ChainID,
		AppHash: h.AppHash,
	}
	baseState := &State{
		ms: ms,
		ctx: sdk.NewContext(ms, h, false, logger).
			WithStreamingManager(sm).
			WithHeaderInfo(headerInfo),
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	switch mode {
	case sdk.ExecModeCheck:
		baseState.SetContext(baseState.Context().WithIsCheckTx(true).WithMinGasPrices(minGasPrices))
		m.checkState = baseState

	case sdk.ExecModePrepareProposal:
		m.prepareProposalState = baseState

	case sdk.ExecModeProcessProposal:
		m.processProposalState = baseState

	case sdk.ExecModeFinalize:
		m.finalizeBlockState = baseState

	default:
		panic(fmt.Sprintf("invalid runTxMode for setState: %d", mode))
	}
}

func (m *Manager) ClearState(mode sdk.ExecMode) {
	m.mut.Lock()
	defer m.mut.Unlock()

	switch mode {
	case sdk.ExecModeCheck:
		m.checkState = nil

	case sdk.ExecModePrepareProposal:
		m.prepareProposalState = nil

	case sdk.ExecModeProcessProposal:
		m.processProposalState = nil

	case sdk.ExecModeFinalize:
		m.finalizeBlockState = nil

	default:
		panic(fmt.Sprintf("invalid runTxMode for clearState: %d", mode))
	}
}
