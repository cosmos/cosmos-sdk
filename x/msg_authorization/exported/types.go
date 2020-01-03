package exported

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

type Authorization interface {
	MsgType() sdk.Msg
	Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Authorization, delete bool)
}
