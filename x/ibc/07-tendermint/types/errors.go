package types

import (
	sdkerrors "github.com/KiraCore/cosmos-sdk/types/errors"
)

const (
	SubModuleName = "tendermint"
)

// IBC tendermint client sentinel errors
var (
	ErrInvalidTrustingPeriod  = sdkerrors.Register(SubModuleName, 2, "invalid trusting period")
	ErrInvalidUnbondingPeriod = sdkerrors.Register(SubModuleName, 3, "invalid unbonding period")
	ErrInvalidHeader          = sdkerrors.Register(SubModuleName, 4, "invalid header")
)
