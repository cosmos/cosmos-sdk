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

func SignatureDataToSignerInfoSig(data types.SignatureData) (*types.ModeInfo, []byte) {
	if data == nil {
		return nil, nil
	}

	switch data := data.(type) {
	case *types.SingleSignatureData:
		return &types.ModeInfo{
			Sum: &types.ModeInfo_Single_{
				Single: &types.ModeInfo_Single{Mode: data.SignMode},
			},
		}, data.Signature
	case *types.MultiSignatureData:
		n := len(data.Signatures)
		modeInfos := make([]*types.ModeInfo, n)
		sigs := make([][]byte, n)

		for i, d := range data.Signatures {
			modeInfos[i], sigs[i] = SignatureDataToSignerInfoSig(d)
		}

		multisig := cryptotypes.MultiSignature{
			Signatures: sigs,
		}
		sig, err := multisig.Marshal()
		if err != nil {
			panic(err)
		}

		return &types.ModeInfo{
			Sum: &types.ModeInfo_Multi_{
				Multi: &types.ModeInfo_Multi{
					Bitarray:  data.BitArray,
					ModeInfos: modeInfos,
				},
			},
		}, sig
	default:
		panic("unexpected case")
	}
}

func (t TxBuilder) SetSignatures(signatures ...context.SignatureBuilder) error {
	n := len(signatures)
	signerInfos := make([]*types.SignerInfo, n)
	rawSigs := make([][]byte, n)
	for i, sig := range signatures {
		var modeInfo *types.ModeInfo
		modeInfo, rawSigs[i] = SignatureDataToSignerInfoSig(sig.Data)
		pk, err := t.PubKeyCodec.Encode(sig.PubKey)
		if err != nil {
			return err
		}
		signerInfos[i] = &types.SignerInfo{
			PublicKey: pk,
			ModeInfo:  modeInfo,
		}
	}
	t.Tx.AuthInfo.SignerInfos = signerInfos
	t.Tx.Signatures = rawSigs
	return nil
}

func (t TxBuilder) SetFee(amount sdk.Coins) {
	t.Tx.AuthInfo.Fee.Amount = amount
}

func (t TxBuilder) SetGasLimit(limit uint64) {
	t.Tx.AuthInfo.Fee.GasLimit = limit
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
