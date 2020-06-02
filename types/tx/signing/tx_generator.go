package signing

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type TxGenerator struct {
	Marshaler   codec.Marshaler
	PubKeyCodec cryptotypes.PublicKeyCodec
	ModeHandler types.SignModeHandler
}

func NewTxGenerator(marshaler codec.Marshaler, pubKeyCodec cryptotypes.PublicKeyCodec, handler types.SignModeHandler) *TxGenerator {
	return &TxGenerator{Marshaler: marshaler, PubKeyCodec: cryptotypes.CacheWrapCodec(pubKeyCodec), ModeHandler: handler}
}

var _ client.TxGenerator = TxGenerator{}

func (t TxGenerator) NewTxBuilder() client.TxBuilder {
	return TxBuilder{
		Tx:          types.NewTx(),
		Marshaler:   t.Marshaler,
		PubKeyCodec: t.PubKeyCodec,
	}
}

func (t TxGenerator) WrapTxBuilder(tx sdk.Tx) (client.TxBuilder, error) {
	stdTx, ok := tx.(*types.Tx)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", &types.Tx{}, tx)
	}
	return TxBuilder{
		Tx:          stdTx,
		Marshaler:   t.Marshaler,
		PubKeyCodec: t.PubKeyCodec,
	}, nil
}

func (t TxGenerator) TxEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		ptx, ok := tx.(*types.Tx)
		if !ok {
			return nil, fmt.Errorf("expected protobuf Tx, got %T", tx)
		}
		return t.Marshaler.MarshalBinaryBare(ptx)
	}
}

func (t TxGenerator) TxDecoder() sdk.TxDecoder {
	return DefaultTxDecoder(t.Marshaler, t.PubKeyCodec)
}

func (t TxGenerator) TxJSONEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		return t.Marshaler.MarshalJSON(tx)
	}
}

func (t TxGenerator) TxJSONDecoder() sdk.TxDecoder {
	return DefaultJSONTxDecoder(t.Marshaler, t.PubKeyCodec)
}

func (t TxGenerator) SignModeHandler() types.SignModeHandler {
	return t.ModeHandler
}
