package server

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ModuleKey interface {
	InvokerConn

	ModuleID() ModuleID
	Address() sdk.AccAddress
}

type InvokerFactory func(callInfo CallInfo) (Invoker, error)

type CallInfo struct {
	Method string
	Caller ModuleID
}
