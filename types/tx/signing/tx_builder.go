package signing

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type TxBuilder struct {
	*types.Tx
	Marshaler   codec.Marshaler
	PubKeyCodec cryptotypes.PublicKeyCodec
}

var _ context.TxBuilder = TxBuilder{}

func (t TxBuilder) GetTx() sdk.Tx {
	return t.Tx
}

func (t TxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	anys := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		pmsg, ok := msg.(proto.Message)
		if !ok {
			return fmt.Errorf("cannot proto marshal %T", msg)
		}
		any, err := codectypes.NewAnyWithValue(pmsg)
		if err != nil {
			return err
		}
		anys[i] = any
	}
	t.Body.Messages = anys
	return nil
}

func (t TxBuilder) GetSignatures() []sdk.Signature {
	signerInfos := t.Tx.AuthInfo.SignerInfos
	rawSigs := t.Tx.Signatures
	n := len(signerInfos)
	res := make([]sdk.Signature, n)
	for i, si := range signerInfos {
		res[i] = ClientSignature{
			pubKey:    si.PublicKey,
			signature: rawSigs[i],
			modeInfo:  si.ModeInfo,
		}
	}
	return res
}

func (t TxBuilder) SetSignatures(signatures ...context.ClientSignature) error {
	n := len(signatures)
	signerInfos := make([]*types.SignerInfo, n)
	rawSigs := make([][]byte, n)
	for i, sig := range signatures {
		csig, ok := sig.(*ClientSignature)
		if !ok {
			return fmt.Errorf("expected ClientSignature, got %T", sig)
		}
		rawSigs[i] = csig.signature
		signerInfos[i] = &types.SignerInfo{
			PublicKey: csig.pubKey,
			ModeInfo:  csig.modeInfo,
		}
	}
	t.Tx.AuthInfo.SignerInfos = signerInfos
	t.Tx.Signatures = rawSigs
	return nil
}

func (t TxBuilder) GetFee() sdk.Fee {
	return t.Tx.AuthInfo.Fee
}

func (t TxBuilder) SetFee(fee context.ClientFee) error {
	t.Tx.AuthInfo.Fee = &types.Fee{
		Amount:   fee.GetAmount(),
		GasLimit: fee.GetGas(),
	}
	return nil
}

func (t TxBuilder) GetMemo() string {
	return t.Tx.Body.Memo
}

func (t TxBuilder) SetMemo(s string) {
	t.Tx.Body.Memo = s
}

func (t TxBuilder) CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error) {
	bodyBz, err := t.Marshaler.MarshalBinaryBare(t.Body)
	if err != nil {
		return nil, err
	}
	aiBz, err := t.Marshaler.MarshalBinaryBare(t.AuthInfo)
	if err != nil {

		return nil, err
	}
	return DirectSignBytes(bodyBz, aiBz, cid, num, seq)
}
