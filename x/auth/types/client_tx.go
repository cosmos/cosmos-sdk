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

// StdTxBuilder wraps StdTx to implement to the context.TxBuilder interface
type StdTxBuilder struct {
	StdTx
	cdc *codec.Codec
}

var _ client.TxBuilder = &StdTxBuilder{}

// GetTx implements TxBuilder.GetTx
func (s *StdTxBuilder) GetTx() sdk.Tx {
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
		pubKey := sig.PubKey
		var pubKeyBz []byte
		if pubKey != nil {
			pubKeyBz = pubKey.Bytes()
		}
		sigBz, err := SignatureDataToAminoSignature(legacy.Cdc, sig.Data)
		if err != nil {
			return err
		}
		sigs[i] = StdSignature{
			PubKey:    pubKeyBz,
			Signature: sigBz,
		}
	}
	s.Signatures = sigs
	return nil
}

func (s *StdTxBuilder) SetFee(amount sdk.Coins) {
	s.StdTx.Fee.Amount = amount
}

func (s *StdTxBuilder) SetGasLimit(limit uint64) {
	s.StdTx.Fee.Gas = limit
}

// SetMemo implements TxBuilder.SetMemo
func (s *StdTxBuilder) SetMemo(memo string) {
	s.Memo = memo
}

// StdTxGenerator is a context.TxGenerator for StdTx
type StdTxGenerator struct {
	Cdc *codec.Codec
}

var _ client.TxGenerator = StdTxGenerator{}

// NewTxBuilder implements TxGenerator.NewTxBuilder
func (s StdTxGenerator) NewTxBuilder() client.TxBuilder {
	return &StdTxBuilder{
		StdTx: StdTx{},
		cdc:   s.Cdc,
	}
}

// MarshalTx implements TxGenerator.MarshalTx
func (s StdTxGenerator) MarshalTx(tx sdk.Tx) ([]byte, error) {
	return DefaultTxEncoder(s.Cdc)(tx)
}

func (s StdTxGenerator) WrapTxBuilder(tx sdk.Tx) (client.TxBuilder, error) {
	stdTx, ok := tx.(StdTx)
	if !ok {
		return nil, fmt.Errorf("expected %T, got %T", StdTx{}, tx)
	}
	return &StdTxBuilder{StdTx: stdTx, cdc: s.Cdc}, nil
}

func (s StdTxGenerator) SignModeHandler() authsigning.SignModeHandler {
	return LegacyAminoJSONHandler{}
}
