package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
)

type Handler struct {
	ID              app.ModuleID
	DefaultGenesis  func(codec.JSONCodec) json.RawMessage
	ValidateGenesis func(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error
	ExportGenesis   func(sdk.Context, codec.JSONCodec) json.RawMessage
}
