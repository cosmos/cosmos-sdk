package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/signing/direct"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type builder struct {
	tx *tx.Tx

	// bodyBz represents the protobuf encoding of TxBody. This should be encoding
	// from the client using TxRaw if the tx was decoded from the wire
	bodyBz []byte

	// authInfoBz represents the protobuf encoding of TxBody. This should be encoding
	// from the client using TxRaw if the tx was decoded from the wire
	authInfoBz []byte

	// pubKeys represents the cached crypto.PubKey's that were set either from tx decoding
	// or decoded from AuthInfo when GetPubKey's was called
	pubKeys []crypto.PubKey

	marshaler   codec.Marshaler
	pubkeyCodec types.PublicKeyCodec
}

var (
	_ authsigning.SigFeeMemoTx = &builder{}
	_ client.TxBuilder         = &builder{}
	_ direct.ProtoTx           = &builder{}
)

func newBuilder(marshaler codec.Marshaler, pubkeyCodec types.PublicKeyCodec) *builder {
	return &builder{
		tx: &tx.Tx{
			Body: &tx.TxBody{},
			AuthInfo: &tx.AuthInfo{
				Fee: &tx.Fee{},
			},
		},
		marshaler:   marshaler,
		pubkeyCodec: pubkeyCodec,
	}
}

func (t *builder) GetMsgs() []sdk.Msg {
	anys := t.tx.Body.Messages
	res := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		msg := any.GetCachedValue().(sdk.Msg)
		res[i] = msg
	}
	return res
}

// MaxGasWanted defines the max gas allowed.
const MaxGasWanted = uint64((1 << 63) - 1)

func (t *builder) ValidateBasic() error {
	theTx := t.tx
	if theTx == nil {
		return fmt.Errorf("bad Tx")
	}

	body := t.tx.Body
	if body == nil {
		return fmt.Errorf("missing TxBody")
	}

	authInfo := t.tx.AuthInfo
	if authInfo == nil {
		return fmt.Errorf("missing AuthInfo")
	}

	fee := authInfo.Fee
	if fee == nil {
		return fmt.Errorf("missing fee")
	}

	if fee.GasLimit > MaxGasWanted {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"invalid gas supplied; %d > %d", fee.GasLimit, MaxGasWanted,
		)
	}

	if fee.Amount.IsAnyNegative() {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"invalid fee provided: %s", fee.Amount,
		)
	}

	sigs := theTx.Signatures

	if len(sigs) == 0 {
		return sdkerrors.ErrNoSignatures
	}

	if len(sigs) != len(t.GetSigners()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrUnauthorized,
			"wrong number of signers; expected %d, got %d", t.GetSigners(), len(sigs),
		)
	}

	return nil
}

func (t *builder) GetBodyBytes() []byte {
	if len(t.bodyBz) == 0 {
		// if bodyBz is empty, then marshal the body. bodyBz will generally
		// be set to nil whenever SetBody is called so the result of calling
		// this method should always return the correct bytes. Note that after
		// decoding bodyBz is derived from TxRaw so that it matches what was
		// transmitted over the wire
		t.bodyBz = t.marshaler.MustMarshalBinaryBare(t.tx.Body)
	}
	return t.bodyBz
}

func (t *builder) GetAuthInfoBytes() []byte {
	if len(t.authInfoBz) == 0 {
		// if authInfoBz is empty, then marshal the body. authInfoBz will generally
		// be set to nil whenever SetAuthInfo is called so the result of calling
		// this method should always return the correct bytes. Note that after
		// decoding authInfoBz is derived from TxRaw so that it matches what was
		// transmitted over the wire
		t.authInfoBz = t.marshaler.MustMarshalBinaryBare(t.tx.AuthInfo)
	}
	return t.authInfoBz
}

func (t *builder) GetSigners() []sdk.AccAddress {
	var signers []sdk.AccAddress
	seen := map[string]bool{}

	for _, msg := range t.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}

	return signers
}

func (t *builder) GetPubKeys() []crypto.PubKey {
	if t.pubKeys == nil {
		signerInfos := t.tx.AuthInfo.SignerInfos
		pubKeys := make([]crypto.PubKey, len(signerInfos))

		for i, si := range signerInfos {
			var err error
			pk := si.PublicKey
			if pk != nil {
				pubKeys[i], err = t.pubkeyCodec.Decode(si.PublicKey)
				if err != nil {
					panic(err)
				}
			}
		}

		t.pubKeys = pubKeys
	}

	return t.pubKeys
}

func (t *builder) GetGas() uint64 {
	return t.tx.AuthInfo.Fee.GasLimit
}

func (t *builder) GetFee() sdk.Coins {
	return t.tx.AuthInfo.Fee.Amount
}

func (t *builder) FeePayer() sdk.AccAddress {
	return t.GetSigners()[0]
}

func (t *builder) GetMemo() string {
	return t.tx.Body.Memo
}

func (t *builder) GetSignatures() [][]byte {
	return t.tx.Signatures
}

func (t *builder) GetSignaturesV2() ([]signing.SignatureV2, error) {
	signerInfos := t.tx.AuthInfo.SignerInfos
	sigs := t.tx.Signatures
	pubKeys := t.GetPubKeys()
	n := len(signerInfos)
	res := make([]signing.SignatureV2, n)

	for i, si := range signerInfos {
		var err error
		sigData, err := ModeInfoAndSigToSignatureData(si.ModeInfo, sigs[i])
		if err != nil {
			return nil, err
		}
		res[i] = signing.SignatureV2{
			PubKey: pubKeys[i],
			Data:   sigData,
		}
	}

	return res, nil
}

func (t *builder) SetMsgs(msgs ...sdk.Msg) error {
	anys := make([]*codectypes.Any, len(msgs))

	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
	}

	t.tx.Body.Messages = anys

	// set bodyBz to nil because the cached bodyBz no longer matches tx.Body
	t.bodyBz = nil

	return nil
}

func (t *builder) SetMemo(memo string) {
	t.tx.Body.Memo = memo

	// set bodyBz to nil because the cached bodyBz no longer matches tx.Body
	t.bodyBz = nil
}

func (t *builder) SetGasLimit(limit uint64) {
	if t.tx.AuthInfo.Fee == nil {
		t.tx.AuthInfo.Fee = &tx.Fee{}
	}

	t.tx.AuthInfo.Fee.GasLimit = limit

	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
}

func (t *builder) SetFeeAmount(coins sdk.Coins) {
	if t.tx.AuthInfo.Fee == nil {
		t.tx.AuthInfo.Fee = &tx.Fee{}
	}

	t.tx.AuthInfo.Fee.Amount = coins

	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
}

func (t *builder) SetSignatures(signatures ...signing.SignatureV2) error {
	n := len(signatures)
	signerInfos := make([]*tx.SignerInfo, n)
	rawSigs := make([][]byte, n)

	for i, sig := range signatures {
		var modeInfo *tx.ModeInfo
		modeInfo, rawSigs[i] = SignatureDataToModeInfoAndSig(sig.Data)
		pk, err := t.pubkeyCodec.Encode(sig.PubKey)
		if err != nil {
			return err
		}
		signerInfos[i] = &tx.SignerInfo{
			PublicKey: pk,
			ModeInfo:  modeInfo,
		}
	}

	t.setSignerInfos(signerInfos)
	t.setSignatures(rawSigs)

	return nil
}

func (t *builder) setSignerInfos(infos []*tx.SignerInfo) {
	t.tx.AuthInfo.SignerInfos = infos
	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
	// set cached pubKeys to nil because they no longer match tx.AuthInfo
	t.pubKeys = nil
}

func (t *builder) setSignatures(sigs [][]byte) {
	t.tx.Signatures = sigs
}

func (t *builder) GetTx() authsigning.SigFeeMemoTx {
	return t
}
