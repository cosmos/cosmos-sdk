package module

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type InitGenesisHandler func(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) ([]abci.ValidatorUpdate, error)
type ExportGenesisHandler func(ctx sdk.Context, cdc codec.JSONCodec) (json.RawMessage, error)
