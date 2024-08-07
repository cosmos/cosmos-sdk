package stf

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf/mock"
)

// There some field not be exported
// Helpers for cometbft test

func GetExecutionContext(ctx context.Context) *executionContext {
	executionCtx, ok := ctx.(*executionContext)
	if !ok {
		return nil
	}
	return executionCtx
}

func GetStateFromContext(ctx *executionContext) store.WriterMap {
	return ctx.state
}

func SetMsgRouter(s *STF[mock.Tx], msgRouter Router) {
	s.msgRouter = msgRouter
}

func SetQueryRouter(s *STF[mock.Tx], queryRouter Router) {
	s.queryRouter = queryRouter
}
