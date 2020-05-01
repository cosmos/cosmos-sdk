package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	SubModuleName = "tendermint"
)

var (
	ErrInvalidTrustingPeriod  = sdkerrors.Register(SubModuleName, 2, "invalid trusting period")
	ErrInvalidUnbondingPeriod = sdkerrors.Register(SubModuleName, 3, "invalid unbonding period")
	ErrInvalidHeader          = sdkerrors.Register(SubModuleName, 4, "invalid header")
)
