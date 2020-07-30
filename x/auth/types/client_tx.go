package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// StdTxBuilder wraps StdTx to implement to the context.TxBuilder interface.
// Note that this type just exists for backwards compatibility with amino StdTx
// and will not work for protobuf transactions.
type StdTxBuilder struct {
	StdTx
	cdc *codec.Codec
}

var _ client.TxBuilder = &StdTxBuilder{}

// GetTx implements TxBuilder.GetTx
func (s *StdTxBuilder) GetTx() authsigning.SigFeeMemoTx {
	return s.StdTx
}

// SetMsgs implements TxBuilder.SetMsgs
func (s *StdTxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	s.Msgs = msgs
	return nil
}

// SetSignatures implements TxBuilder.SetSignatures
func (s *StdTxBuilder) SetSignatures(signatures ...signing.SignatureV2) error {
	sigs := make([]StdSignature, len(signatures))

	for i, sig := range signatures {
		var pubKeyBz []byte

		pubKey := sig.PubKey
		if pubKey != nil {
			pubKeyBz = pubKey.Bytes()
		}

		var (
			sigBz []byte
			err   error
		)

		if sig.Data != nil {
			sigBz, err = SignatureDataToAminoSignature(legacy.Cdc, sig.Data)
			if err != nil {
				return err
			}
		}

		sigs[i] = StdSignature{
			PubKey:    pubKeyBz,
			Signature: sigBz,
		}
	}

	s.Signatures = sigs
	return nil
}

func (s *StdTxBuilder) SetFeeAmount(amount sdk.Coins) {
	s.StdTx.Fee.Amount = amount
}

func (s *StdTxBuilder) SetGasLimit(limit uint64) {
	s.StdTx.Fee.Gas = limit
}

// SetMemo implements TxBuilder.SetMemo
func (s *StdTxBuilder) SetMemo(memo string) {
	s.Memo = memo
}

// StdTxConfig is a context.TxConfig for StdTx
type StdTxConfig struct {
	Cdc *codec.Codec
}

var _ client.TxConfig = StdTxConfig{}

// NewTxBuilder implements TxConfig.NewTxBuilder
func (s StdTxConfig) NewTxBuilder() client.TxBuilder {
	return &StdTxBuilder{
		StdTx: StdTx{},
		cdc:   s.Cdc,
	}
}

// WrapTxBuilder returns a StdTxBuilder from provided transaction
func (s StdTxConfig) WrapTxBuilder(newTx sdk.Tx) (client.TxBuilder, error) {
	stdTx, ok := newTx.(StdTx)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", StdTx{}, newTx)
	}
	return &StdTxBuilder{StdTx: stdTx, cdc: s.Cdc}, nil
}

// MarshalTx implements TxConfig.MarshalTx
func (s StdTxConfig) TxEncoder() sdk.TxEncoder {
	return DefaultTxEncoder(s.Cdc)
}

func (s StdTxConfig) TxDecoder() sdk.TxDecoder {
	return DefaultTxDecoder(s.Cdc)
}

func (s StdTxConfig) TxJSONEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		return s.Cdc.MarshalJSON(tx)
	}
}

func (s StdTxConfig) TxJSONDecoder() sdk.TxDecoder {
	return DefaultJSONTxDecoder(s.Cdc)
}

func (s StdTxConfig) SignModeHandler() authsigning.SignModeHandler {
	return legacyAminoJSONHandler{}
}
