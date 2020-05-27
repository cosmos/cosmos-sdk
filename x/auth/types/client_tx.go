package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StdTxBuilder wraps StdTx to implement to the context.TxBuilder interface
type StdTxBuilder struct {
	StdTx
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

// GetSignatures implements TxBuilder.GetSignatures
func (s StdTxBuilder) GetSignatures() []sdk.Signature {
	res := make([]sdk.Signature, len(s.Signatures))
	for i, sig := range s.Signatures {
		res[i] = sig
	}
	return res
}

// SetSignatures implements TxBuilder.SetSignatures
func (s *StdTxBuilder) SetSignatures(signatures ...client.Signature) error {
	sigs := make([]StdSignature, len(signatures))
	for i, sig := range signatures {
		pubKey := sig.GetPubKey()
		var pubKeyBz []byte
		if pubKey != nil {
			pubKeyBz = pubKey.Bytes()
		}
		sigs[i] = StdSignature{
			PubKey:    pubKeyBz,
			Signature: sig.GetSignature(),
		}
	}
	s.Signatures = sigs
	return nil
}

// GetFee implements TxBuilder.GetFee
func (s StdTxBuilder) GetFee() sdk.Fee {
	return s.Fee
}

// SetFee implements TxBuilder.SetFee
func (s *StdTxBuilder) SetFee(fee client.Fee) error {
	s.Fee = StdFee{Amount: fee.GetAmount(), Gas: fee.GetGas()}
	return nil
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

var _ client.TxGenerator = StdTxGenerator{}

// NewTx implements TxGenerator.NewTx
func (s StdTxGenerator) NewTx() client.TxBuilder {
	return &StdTxBuilder{}
}

// NewFee implements TxGenerator.NewFee
func (s StdTxGenerator) NewFee() client.Fee {
	return &StdFee{}
}

// NewSignature implements TxGenerator.NewSignature
func (s StdTxGenerator) NewSignature() client.Signature {
	return &StdSignature{}
}

// MarshalTx implements TxGenerator.MarshalTx
func (s StdTxGenerator) MarshalTx(tx sdk.Tx) ([]byte, error) {
	return DefaultTxEncoder(s.Cdc)(tx)
}

var _ client.Fee = &StdFee{}

// SetGas implements Fee.SetGas
func (fee *StdFee) SetGas(gas uint64) {
	fee.Gas = gas
}

// SetAmount implements Fee.SetAmount
func (fee *StdFee) SetAmount(coins sdk.Coins) {
	fee.Amount = coins
}

var _ client.Signature = &StdSignature{}

// SetPubKey implements Signature.SetPubKey
func (ss *StdSignature) SetPubKey(key crypto.PubKey) error {
	ss.PubKey = key.Bytes()
	return nil
}

// SetSignature implements Signature.SetSignature
func (ss *StdSignature) SetSignature(bytes []byte) {
	ss.Signature = bytes
}
