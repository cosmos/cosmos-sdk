package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
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

// NewTxConfig returns a new protobuf TxConfig using the provided Marshaler, PublicKeyCodec and SignModeHandler.
func NewTxConfig(marshaler codec.Marshaler, pubkeyCodec types.PublicKeyCodec, signModeHandler signing.SignModeHandler) client.TxConfig {
	return &generator{
		marshaler:   marshaler,
		pubkeyCodec: pubkeyCodec,
		handler:     signModeHandler,
		decoder:     DefaultTxDecoder(marshaler, pubkeyCodec),
		encoder:     DefaultTxEncoder(marshaler),
		jsonDecoder: DefaultJSONTxDecoder(marshaler, pubkeyCodec),
		jsonEncoder: DefaultJSONTxEncoder(marshaler),
	}
}

func (g generator) NewTxBuilder() client.TxBuilder {
	return newBuilder(g.marshaler, g.pubkeyCodec)
}

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
