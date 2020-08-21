package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// StdTxBuilder wraps StdTx to implement to the context.TxBuilder interface.
// Note that this type just exists for backwards compatibility with amino StdTx
// and will not work for protobuf transactions.
type StdTxBuilder struct {
	StdTx
	cdc *codec.LegacyAmino
}

var _ client.TxBuilder = &StdTxBuilder{}

// GetTx implements TxBuilder.GetTx
func (s *StdTxBuilder) GetTx() authsigning.Tx {
	return s.StdTx
}

// SetMsgs implements TxBuilder.SetMsgs
func (s *StdTxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	s.Msgs = msgs
	return nil
}

// SetSignatures implements TxBuilder.SetSignatures.
func (s *StdTxBuilder) SetSignatures(signatures ...signing.SignatureV2) error {
	sigs := make([]StdSignature, len(signatures))

	for i, sig := range signatures {
		stdSig, err := SignatureV2ToStdSignature(s.cdc, sig)
		if err != nil {
			return err
		}

		sigs[i] = stdSig
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

// SetTimeoutHeight sets the transaction's height timeout.
func (s *StdTxBuilder) SetTimeoutHeight(height uint64) {
	s.TimeoutHeight = height
}

// StdTxConfig is a context.TxConfig for StdTx
type StdTxConfig struct {
	Cdc *codec.LegacyAmino
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

func (s StdTxConfig) MarshalSignatureJSON(sigs []signing.SignatureV2) ([]byte, error) {
	stdSigs := make([]StdSignature, len(sigs))
	for i, sig := range sigs {
		stdSig, err := SignatureV2ToStdSignature(s.Cdc, sig)
		if err != nil {
			return nil, err
		}

		stdSigs[i] = stdSig
	}
	return s.Cdc.MarshalJSON(stdSigs)
}

func (s StdTxConfig) UnmarshalSignatureJSON(bz []byte) ([]signing.SignatureV2, error) {
	var stdSigs []StdSignature
	err := s.Cdc.UnmarshalJSON(bz, &stdSigs)
	if err != nil {
		return nil, err
	}

	sigs := make([]signing.SignatureV2, len(stdSigs))
	for i, stdSig := range stdSigs {
		sig, err := StdSignatureToSignatureV2(s.Cdc, stdSig)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	return sigs, nil
}

func (s StdTxConfig) SignModeHandler() authsigning.SignModeHandler {
	return stdTxSignModeHandler{}
}
