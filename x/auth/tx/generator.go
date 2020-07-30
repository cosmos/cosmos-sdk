package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type generator struct {
	pubkeyCodec types.PublicKeyCodec
	handler     signing.SignModeHandler
	decoder     sdk.TxDecoder
	encoder     sdk.TxEncoder
	jsonDecoder sdk.TxDecoder
	jsonEncoder sdk.TxEncoder
	protoCodec  *codec.ProtoCodec
}

// NewTxConfig returns a new protobuf TxConfig using the provided ProtoCodec, PublicKeyCodec and SignModeHandler.
func NewTxConfig(protoCodec *codec.ProtoCodec, pubkeyCodec types.PublicKeyCodec, signModeHandler signing.SignModeHandler) client.TxConfig {
	return &generator{
		pubkeyCodec: pubkeyCodec,
		handler:     signModeHandler,
		decoder:     DefaultTxDecoder(protoCodec, pubkeyCodec),
		encoder:     DefaultTxEncoder(),
		jsonDecoder: DefaultJSONTxDecoder(protoCodec, pubkeyCodec),
		jsonEncoder: DefaultJSONTxEncoder(),
		protoCodec:  protoCodec,
	}
}

func (g generator) NewTxBuilder() client.TxBuilder {
	return newBuilder(g.pubkeyCodec)
}

// WrapTxBuilder returns a builder from provided transaction
func (g generator) WrapTxBuilder(newTx sdk.Tx) (client.TxBuilder, error) {
	newBuilder, ok := newTx.(*builder)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", &builder{}, newTx)
	}

	return newBuilder, nil
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
