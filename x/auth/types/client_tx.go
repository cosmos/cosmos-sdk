package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy_global"
	"github.com/cosmos/cosmos-sdk/crypto/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

// StdTxBuilder wraps StdTx to implement to the context.TxBuilder interface
type StdTxBuilder struct {
	StdTx
}

var _ client.TxBuilder = &StdTxBuilder{}

// GetTx implements TxBuilder.GetTx
func (s *StdTxBuilder) GetTx() types.SigTx {
	return s.StdTx
}

// SetMsgs implements TxBuilder.SetMsgs
func (s *StdTxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	s.Msgs = msgs
	return nil
}

func SignatureDataToSig(data types.SignatureData) []byte {
	if data == nil {
		return nil
	}

	switch data := data.(type) {
	case *types.SingleSignatureData:
		return data.Signature
	case *types.MultiSignatureData:
		n := len(data.Signatures)
		sigs := make([][]byte, n)
		for i, s := range data.Signatures {
			sigs[i] = SignatureDataToSig(s)
		}
		msig := multisig.AminoMultisignature{
			BitArray: data.BitArray,
			Sigs:     sigs,
		}
		return legacy_global.Cdc.MustMarshalBinaryBare(msig)
	default:
		panic("unexpected case")
	}
}

// SetSignatures implements TxBuilder.SetSignatures
func (s *StdTxBuilder) SetSignatures(signatures ...client.SignatureBuilder) error {
	sigs := make([]StdSignature, len(signatures))
	for i, sig := range signatures {
		pubKey := sig.PubKey
		var pubKeyBz []byte
		if pubKey != nil {
			pubKeyBz = pubKey.Bytes()
		}
		sigs[i] = StdSignature{
			PubKey:    pubKeyBz,
			Signature: SignatureDataToSig(sig.Data),
		}
	}
	s.Signatures = sigs
	return nil
}

func (s *StdTxBuilder) SetFee(amount sdk.Coins) {
	s.Fee.Amount = amount
}

func (s *StdTxBuilder) SetGasLimit(limit uint64) {
	s.Fee.Gas = limit
}

// SetMemo implements TxBuilder.SetMemo
func (s *StdTxBuilder) SetMemo(memo string) {
	s.Memo = memo
}

// CanonicalSignBytes implements TxBuilder.CanonicalSignBytes
func (s StdTxBuilder) CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error) {
	return StdSignBytes(cid, num, seq, s.Fee, s.Msgs, s.Memo), nil
}

// StdTxGenerator is a context.TxGenerator for StdTx
type StdTxGenerator struct {
	Cdc *codec.Codec
}

func (s StdTxGenerator) SignModeHandler() types.SignModeHandler {
	return LegacyAminoJSONHandler{}
}

var _ client.TxGenerator = StdTxGenerator{}

// NewTxBuilder implements TxGenerator.NewTxBuilder
func (s StdTxGenerator) NewTxBuilder() client.TxBuilder {
	return &StdTxBuilder{}
}

// MarshalTx implements TxGenerator.MarshalTx
func (s StdTxGenerator) TxEncoder() sdk.TxEncoder {
	return DefaultTxEncoder(s.Cdc)
}

func (s StdTxGenerator) TxDecoder() sdk.TxDecoder {
	return DefaultTxDecoder(s.Cdc)
}

func (s StdTxGenerator) TxJSONEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		return s.Cdc.MarshalJSON(tx)
	}
}

func (s StdTxGenerator) TxJSONDecoder() sdk.TxDecoder {
	return DefaultJSONTxDecoder(s.Cdc)
}

type LegacyAminoJSONHandler struct{}

var _ types.SignModeHandler = LegacyAminoJSONHandler{}

func (h LegacyAminoJSONHandler) DefaultMode() types.SignMode {
	return types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (LegacyAminoJSONHandler) Modes() []types.SignMode {
	return []types.SignMode{types.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}
}

func (LegacyAminoJSONHandler) GetSignBytes(data types.SigningData, tx sdk.Tx) ([]byte, error) {
	feeTx, ok := tx.(types.FeeTx)
	if !ok {
		return nil, fmt.Errorf("expected FeeTx, got %T", tx)
	}

	memoTx, ok := tx.(types.TxWithMemo)
	if !ok {
		return nil, fmt.Errorf("expected TxWithMemo, got %T", tx)
	}

	return StdSignBytes(
		data.ChainID, data.AccountNumber, data.AccountSequence, StdFee{Amount: feeTx.GetFee(), Gas: feeTx.GetGas()}, tx.GetMsgs(), memoTx.GetMemo(),
	), nil
}
