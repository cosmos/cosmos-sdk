package signing

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type TxGenerator struct {
	Marshaler   codec.Marshaler
	PubKeyCodec cryptotypes.PublicKeyCodec
}

func NewTxGenerator(marshaler codec.Marshaler, pubKeyCodec cryptotypes.PublicKeyCodec) *TxGenerator {
	return &TxGenerator{Marshaler: marshaler, PubKeyCodec: pubKeyCodec}
}

var _ context.TxGenerator = TxGenerator{}

func (t TxGenerator) NewTx() context.TxBuilder {
	return TxBuilder{
		Tx: &types.Tx{
			Body:       &types.TxBody{},
			AuthInfo:   &types.AuthInfo{},
			Signatures: nil,
		},
		Marshaler:   t.Marshaler,
		PubKeyCodec: t.PubKeyCodec,
	}
}

func (t TxGenerator) NewFee() context.ClientFee {
	return &types.Fee{}
}

func (t TxGenerator) NewSignature() context.ClientSignature {
	return &ClientSignature{
		modeInfo: &types.ModeInfo{
			Sum: &types.ModeInfo_Single_{
				Single: &types.ModeInfo_Single{
					Mode: types.SignMode_SIGN_MODE_DIRECT,
				},
			},
		},
		codec: t.PubKeyCodec,
	}
}

func (t TxGenerator) MarshalTx(tx sdk.Tx) ([]byte, error) {
	ptx, ok := tx.(*types.Tx)
	if !ok {
		return nil, fmt.Errorf("expected protobuf Tx, got %T", tx)
	}
	return t.Marshaler.MarshalBinaryBare(ptx)
}
