package header

import (
	"context"
	"fmt"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/core/abci"
)

var Middleware abci.Middleware = func(handler abci.Handler) abci.Handler {
	return &middleware{
		Handler: handler,
	}
}

type middleware struct {
	abci.Handler
	initialHeight   int64
	lastBlockHeight int64
}

func (m *middleware) InitChain(ctx context.Context, req types.RequestInitChain) types.ResponseInitChain {
	// On a new chain, we consider the init chain block height as 0, even though
	// req.InitialHeight is 1 by default.
	initHeader := tmproto.Header{ChainID: req.ChainId, Time: req.Time}

	// If req.InitialHeight is > 1, then we set the initial version in the
	// stores.
	if req.InitialHeight > 1 {
		m.initialHeight = req.InitialHeight
		initHeader = tmproto.Header{ChainID: req.ChainId, Height: req.InitialHeight, Time: req.Time}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(initHeader)
	ctx = context.WithValue(ctx, sdk.SdkContextKey, sdkCtx)

	return m.Handler.InitChain(ctx, req)
}

func (m *middleware) BeginBlock(ctx context.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
	if err := m.validateHeight(req); err != nil {
		panic(err)
	}

	m.lastBlockHeight = req.Header.Height

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.
		WithBlockHeader(req.Header).
		WithBlockHeight(req.Header.Height)
	ctx = context.WithValue(ctx, sdk.SdkContextKey, sdkCtx)

	return m.Handler.BeginBlock(ctx, req)
}

func (m middleware) validateHeight(req types.RequestBeginBlock) error {
	if req.Header.Height < 1 {
		return fmt.Errorf("invalid height: %d", req.Header.Height)
	}

	// expectedHeight holds the expected height to validate.
	var expectedHeight int64
	if m.lastBlockHeight == 0 && m.initialHeight > 1 {
		// In this case, we're validating the first block of the chain (no
		// previous commit). The height we're expecting is the initial height.
		expectedHeight = m.initialHeight
	} else {
		// This case can means two things:
		// - either there was already a previous commit in the store, in which
		// case we increment the version from there,
		// - or there was no previous commit, and initial version was not set,
		// in which case we start at version 1.
		expectedHeight = m.lastBlockHeight + 1
	}

	if req.Header.Height != expectedHeight {
		return fmt.Errorf("invalid height: %d; expected: %d", req.Header.Height, expectedHeight)
	}

	return nil
}
