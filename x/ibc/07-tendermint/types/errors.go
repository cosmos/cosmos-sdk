package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	SubModuleName = "tendermint"
)

var (
	ErrInvalidTrustingPeriod  = sdkerrors.Register(SubModuleName, 1, "invalid trusting period")
	ErrInvalidUnbondingPeriod = sdkerrors.Register(SubModuleName, 2, "invalid unbonding period")
	ErrInvalidHeader          = sdkerrors.Register(SubModuleName, 3, "invalid header")
)
