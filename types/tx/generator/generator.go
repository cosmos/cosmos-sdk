package generator

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing/direct"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type generator struct {
	marshaler   codec.Marshaler
	pubkeyCodec types.PublicKeyCodec
	handler     signing.SignModeHandler
	decoder     sdk.TxDecoder
	encoder     sdk.TxEncoder
	jsonDecoder sdk.TxDecoder
	jsonEncoder sdk.TxEncoder
}

func New(marshaler codec.Marshaler, pubkeyCodec types.PublicKeyCodec) client.TxGenerator {
	return &generator{
		marshaler:   marshaler,
		pubkeyCodec: pubkeyCodec,
		handler:     DefaultSignModeHandler(),
		decoder:     DefaultTxDecoder(marshaler, pubkeyCodec),
		encoder:     DefaultTxEncoder(marshaler),
		jsonDecoder: DefaultJSONTxDecoder(marshaler, pubkeyCodec),
		jsonEncoder: DefaultJSONTxEncoder(marshaler),
	}
}

func (g generator) NewTxBuilder() client.TxBuilder {
	return NewBuilder(g.marshaler, g.pubkeyCodec)
}

func (g generator) SignModeHandler() signing.SignModeHandler {
	return g.handler
}

func (g generator) TxEncoder() sdk.TxEncoder {
	return g.encoder
}

func (g generator) TxDecoder() sdk.TxDecoder {
	return g.decoder
}

func (g generator) TxJSONEncoder() sdk.TxEncoder {
	return g.jsonEncoder
}

func (g generator) TxJSONDecoder() sdk.TxDecoder {
	return g.jsonDecoder
}

func DefaultSignModeHandler() signing.SignModeHandler {
	return signing.NewSignModeHandlerMap(
		signing2.SignMode_SIGN_MODE_DIRECT,
		[]signing.SignModeHandler{
			types2.LegacyAminoJSONHandler{},
			direct.ModeHandler{},
		},
	)
}
