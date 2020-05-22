package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	SubModuleName = "tendermint"
)

// IBC tendermint client sentinel errors
var (
	ErrInvalidTrustingPeriod  = sdkerrors.Register(SubModuleName, 2, "invalid trusting period")
	ErrInvalidUnbondingPeriod = sdkerrors.Register(SubModuleName, 3, "invalid unbonding period")
	ErrInvalidHeader          = sdkerrors.Register(SubModuleName, 4, "invalid header")
	ErrInvalidMaxClockDrift   = sdkerrors.Register(SubModuleName, 5, "invalid max clock drift")
	ErrTrustingPeriodExpired  = sdkerrors.Register(SubModuleName, 6, "time since latest trusted state has passed the trusting period")
	ErrUnbondingPeriodExpired = sdkerrors.Register(SubModuleName, 7, "time since latest trusted state has passed the unbonding period")
)
