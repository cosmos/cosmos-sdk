package tx

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TxWrapper defines a type that is to be used for building, signing and verifying
// protobuf transactions. The protobuf Tx type is not used directly because a) protobuf
// SIGN_MODE_DIRECT signing uses raw body and auth_info bytes and b) Tx does does not retain
// crypto.PubKey instances.
type TxWrapper interface {
	sdk.Tx

	ProtoTx

	GetPubKeys() []crypto.PubKey // If signer already has pubkey in context, this list will have nil in its place

	SetMsgs([]sdk.Msg)
	SetMemo(string)
	SetGas(uint64)
	SetFee(sdk.Coins)
	// TODO: replace SetSignerInfos with SetSignaturesV2 once SignatureV2 from #6373 is merged in
	SetSignerInfos([]*SignerInfo)
	SetSignatures([][]byte)
}

func NewTxWrapper(marshaler codec.Marshaler, pubkeyCodec types.PublicKeyCodec) TxWrapper {
	return &txWrapper{
		tx: &Tx{
			Body:     &TxBody{},
			AuthInfo: &AuthInfo{},
		},
		marshaler:   marshaler,
		pubkeyCodec: pubkeyCodec,
	}
}

type txWrapper struct {
	tx *Tx

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

var _ TxWrapper = &txWrapper{}

func (t txWrapper) GetMsgs() []sdk.Msg {
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

func (t txWrapper) ValidateBasic() error {
	tx := t.tx
	if tx == nil {
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

	sigs := tx.Signatures

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

func (t *txWrapper) GetBodyBytes() []byte {
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

func (t *txWrapper) GetAuthInfoBytes() []byte {
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

func (t txWrapper) GetSigners() []sdk.AccAddress {
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

func (t txWrapper) GetPubKeys() []crypto.PubKey {
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

func (t *txWrapper) SetMsgs(msgs []sdk.Msg) {
	anys := make([]*codectypes.Any, len(msgs))

	for i, msg := range msgs {
		var err error
		anys[i], err = codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
	}

	t.tx.Body.Messages = anys

	// set bodyBz to nil because the cached bodyBz no longer matches tx.Body
	t.bodyBz = nil
}

func (t *txWrapper) SetMemo(memo string) {
	t.tx.Body.Memo = memo

	// set bodyBz to nil because the cached bodyBz no longer matches tx.Body
	t.bodyBz = nil
}

func (t *txWrapper) SetGas(limit uint64) {
	t.tx.AuthInfo.Fee.GasLimit = limit

	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
}

func (t *txWrapper) SetFee(coins sdk.Coins) {
	t.tx.AuthInfo.Fee.Amount = coins

	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
}

func (t *txWrapper) SetSignerInfos(infos []*SignerInfo) {
	t.tx.AuthInfo.SignerInfos = infos
	// set authInfoBz to nil because the cached authInfoBz no longer matches tx.AuthInfo
	t.authInfoBz = nil
	// set cached pubKeys to nil because they no longer match tx.AuthInfo
	t.pubKeys = nil
}

func (t *txWrapper) SetSignatures(sigs [][]byte) {
	t.tx.Signatures = sigs
}
