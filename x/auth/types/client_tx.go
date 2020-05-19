package types

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StdClientTx struct {
	StdTx
}

var _ context.TxBuilder = &StdClientTx{}

func (s *StdClientTx) GetTx() sdk.Tx {
	return s.StdTx
}

func (s *StdClientTx) SetMsgs(msgs ...sdk.Msg) error {
	s.Msgs = msgs
	return nil
}

func (s StdClientTx) GetSignatures() []sdk.Signature {
	res := make([]sdk.Signature, len(s.Signatures))
	for i, sig := range s.Signatures {
		res[i] = sig
	}
	return res
}

func (s *StdClientTx) SetSignatures(signatures ...context.ClientSignature) error {
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

func (s StdClientTx) GetFee() sdk.Fee {
	return s.Fee
}

func (s *StdClientTx) SetFee(fee context.ClientFee) error {
	s.Fee = StdFee{Amount: fee.GetAmount(), Gas: fee.GetGas()}
	return nil
}

func (s *StdClientTx) SetMemo(memo string) {
	s.Memo = memo
}

func (s StdClientTx) CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error) {
	return StdSignBytes(cid, num, seq, s.Fee, s.Msgs, s.Memo), nil
}

type StdTxGenerator struct {
	Cdc *codec.Codec
}

func (s StdTxGenerator) NewTx() context.TxBuilder {
	return &StdClientTx{}
}

func (s StdTxGenerator) NewFee() context.ClientFee {
	return &StdFee{}
}

func (s StdTxGenerator) NewSignature() context.ClientSignature {
	return &StdSignature{}
}

func (s StdTxGenerator) MarshalTx(tx sdk.Tx) ([]byte, error) {
	return DefaultTxEncoder(s.Cdc)(tx)
}

var _ context.TxGenerator = StdTxGenerator{}

func (fee *StdFee) SetGas(gas uint64) {
	fee.Gas = gas
}

func (fee *StdFee) SetAmount(coins sdk.Coins) {
	fee.Amount = coins
}

func (ss *StdSignature) SetPubKey(key crypto.PubKey) error {
	ss.PubKey = key.Bytes()
	return nil
}

func (ss *StdSignature) SetSignature(bytes []byte) {
	ss.Signature = bytes
}
