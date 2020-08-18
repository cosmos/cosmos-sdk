package tx

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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

	pubkeyCodec types.PublicKeyCodec

	txBodyHasUnknownNonCriticals bool
}

var (
	_ authsigning.Tx             = &builder{}
	_ client.TxBuilder           = &builder{}
	_ ante.HasExtensionOptionsTx = &builder{}
	_ ExtensionOptionsTxBuilder  = &builder{}
)

// ExtensionOptionsTxBuilder defines a TxBuilder that can also set extensions.
type ExtensionOptionsTxBuilder interface {
	client.TxBuilder

	SetExtensionOptions(...*codectypes.Any)
	SetNonCriticalExtensionOptions(...*codectypes.Any)
}

func newBuilder(pubkeyCodec types.PublicKeyCodec) *builder {
	return &builder{
		tx: &tx.Tx{
			Body: &tx.TxBody{},
			AuthInfo: &tx.AuthInfo{
				Fee: &tx.Fee{},
			},
		},
		pubkeyCodec: pubkeyCodec,
	}
}

func (t *builder) GetMsgs() []sdk.Msg {
	if t.tx == nil || t.tx.Body == nil {
		return nil
	}

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

func (t *builder) getBodyBytes() []byte {
	if len(t.bodyBz) == 0 {
		// if bodyBz is empty, then marshal the body. bodyBz will generally
		// be set to nil whenever SetBody is called so the result of calling
		// this method should always return the correct bytes. Note that after
		// decoding bodyBz is derived from TxRaw so that it matches what was
		// transmitted over the wire
		var err error
		t.bodyBz, err = proto.Marshal(t.tx.Body)
		if err != nil {
			panic(err)
		}
	}
	return t.bodyBz
}

func (t *builder) getAuthInfoBytes() []byte {
	if len(t.authInfoBz) == 0 {
		// if authInfoBz is empty, then marshal the body. authInfoBz will generally
		// be set to nil whenever SetAuthInfo is called so the result of calling
		// this method should always return the correct bytes. Note that after
		// decoding authInfoBz is derived from TxRaw so that it matches what was
		// transmitted over the wire
		var err error
		t.authInfoBz, err = proto.Marshal(t.tx.AuthInfo)
		if err != nil {
			panic(err)
		}
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

// GetTimeoutHeight returns the transaction's timeout height (if set).
func (t *builder) GetTimeoutHeight() uint64 {
	return t.tx.Body.TimeoutHeight
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

// SetTimeoutHeight sets the transaction's height timeout.
func (t *builder) SetTimeoutHeight(height uint64) {
	t.tx.Body.TimeoutHeight = height

	// set bodyBz to nil because the cached bodyBz no longer matches tx.Body
	t.bodyBz = nil
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

// getSignerIndex returns the index of a public key in the GetSigners array. It
// returns an error if the publicKey is not in GetSigners.
func (t *builder) getSignerIndex(pubKey crypto.PubKey) (int, error) {
	if pubKey == nil {
		return -1, sdkerrors.Wrap(
			sdkerrors.ErrInvalidPubKey,
			"public key is empty",
		)
	}

	for i, signer := range t.GetSigners() {
		if bytes.Equal(signer.Bytes(), pubKey.Address().Bytes()) {
			return i, nil
		}
	}

	return -1, sdkerrors.Wrapf(
		sdkerrors.ErrInvalidPubKey,
		"public key %s is not a signer of this tx, call SetMsgs first", pubKey,
	)
}

// SetSignerInfo implements TxBuilder.SetSignerInfo.
func (t *builder) SetSignerInfo(pubKey crypto.PubKey, modeInfo *tx.ModeInfo) error {
	signerIndex, err := t.getSignerIndex(pubKey)
	if err != nil {
		return err
	}

	pk, err := t.pubkeyCodec.Encode(pubKey)
	if err != nil {
		return err
	}

	n := len(t.GetSigners())
	// If t.tx.AuthInfo.SignerInfos is empty, we just initialize with some
	// empty data.
	if len(t.tx.AuthInfo.SignerInfos) == 0 {
		t.tx.AuthInfo.SignerInfos = make([]*tx.SignerInfo, n)
		for i := 1; i < n; i++ {
			t.tx.AuthInfo.SignerInfos[i] = &tx.SignerInfo{}
		}
	}

	t.tx.AuthInfo.SignerInfos[signerIndex] = &tx.SignerInfo{
		PublicKey: pk,
		ModeInfo:  modeInfo,
	}

	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
	// set cached pubKeys to nil because they no longer match tx.AuthInfo
	t.pubKeys = nil

	return nil
}

func (t *builder) setSignatures(sigs [][]byte) {
	t.tx.Signatures = sigs
}

func (t *builder) GetTx() authsigning.Tx {
	return t
}

// GetProtoTx returns the tx as a proto.Message.
func (t *builder) GetProtoTx() *tx.Tx {
	return t.tx
}

// WrapTxBuilder creates a TxBuilder wrapper around a tx.Tx proto message.
func WrapTxBuilder(protoTx *tx.Tx, pubkeyCodec types.PublicKeyCodec) client.TxBuilder {
	return &builder{
		tx:          protoTx,
		pubkeyCodec: pubkeyCodec,
	}
}

func (t *builder) GetExtensionOptions() []*codectypes.Any {
	return t.tx.Body.ExtensionOptions
}

func (t *builder) GetNonCriticalExtensionOptions() []*codectypes.Any {
	return t.tx.Body.NonCriticalExtensionOptions
}

func (t *builder) SetExtensionOptions(extOpts ...*codectypes.Any) {
	t.tx.Body.ExtensionOptions = extOpts
	t.bodyBz = nil
}

func (t *builder) SetNonCriticalExtensionOptions(extOpts ...*codectypes.Any) {
	t.tx.Body.NonCriticalExtensionOptions = extOpts
	t.bodyBz = nil
}
