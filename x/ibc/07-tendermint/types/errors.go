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
	ErrInvalidTrustingPeriod  = sdkerrors.Register(SubModuleName, 5, "invalid trusting period")
	ErrInvalidUnbondingPeriod = sdkerrors.Register(SubModuleName, 6, "invalid unbonding period")
	ErrInvalidMaxClockDrift   = sdkerrors.Register(SubModuleName, 7, "invalid max clock drift")
	ErrTrustingPeriodExpired  = sdkerrors.Register(SubModuleName, 8, "time since latest trusted state has passed the trusting period")
	ErrUnbondingPeriodExpired = sdkerrors.Register(SubModuleName, 9, "time since latest trusted state has passed the unbonding period")
)
