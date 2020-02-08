package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Authorization interface {
	Msg() sdk.Msg
	MsgType() string
	Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Authorization, delete bool)
}

type AuthorizationGrant struct {
	Authorization Authorization `json:"authorization"`

	Expiration int64 `json:"expiration"`
}
