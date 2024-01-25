package tx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/auth/ante"
	authsigning "cosmossdk.io/x/auth/signing"
	"cosmossdk.io/x/tx/decode"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func newWrapperFromDecodedTx(addrCodec address.Codec, cdc codec.BinaryCodec, decodedTx *decode.DecodedTx) (w *gogoTxWrapper, err error) {
	// set msgsv1
	msgv1, err := decodeMsgsV1(cdc, decodedTx.Tx.Body.Messages)
	if err != nil {
		return nil, fmt.Errorf("unable to convert messagev2 to messagev1: %w", err)
	}
	// set fees
	fees := make(sdk.Coins, len(decodedTx.Tx.AuthInfo.Fee.Amount))
	for i, fee := range decodedTx.Tx.AuthInfo.Fee.Amount {
		amtInt, ok := math.NewIntFromString(fee.Amount)
		if !ok {
			return nil, fmt.Errorf("invalid fee coin amount at index %d: %s", i, fee.Amount)
		}
		if err = sdk.ValidateDenom(fee.Denom); err != nil {
			return nil, fmt.Errorf("invalid fee coin denom at index %d: %w", i, err)
		}
		fees[i] = sdk.Coin{
			Denom:  fee.Denom,
			Amount: amtInt,
		}
	}
	if !fees.IsSorted() {
		return nil, fmt.Errorf("invalid not sorted tx fees: %s", fees.String())
	}
	// set fee payer
	var feePayer []byte
	if len(decodedTx.Signers) != 0 {
		feePayer = decodedTx.Signers[0]
		if decodedTx.Tx.AuthInfo.Fee.Payer != "" {
			feePayer, err = addrCodec.StringToBytes(decodedTx.Tx.AuthInfo.Fee.Payer)
			if err != nil {
				return nil, err
			}
		}
	}

	// fee granter
	var feeGranter []byte
	if decodedTx.Tx.AuthInfo.Fee.Granter != "" {
		feeGranter, err = addrCodec.StringToBytes(decodedTx.Tx.AuthInfo.Fee.Granter)
		if err != nil {
			return nil, err
		}
	}
	return &gogoTxWrapper{
		cdc:        cdc,
		decodedTx:  decodedTx,
		msgsV1:     msgv1,
		fees:       fees,
		feePayer:   feePayer,
		feeGranter: feeGranter,
	}, nil
}

// gogoTxWrapper is a gogoTxWrapper around the tx.Tx proto.Message which retain the raw
// body and auth_info bytes.
type gogoTxWrapper struct {
	decodedTx *decode.DecodedTx
	cdc       codec.BinaryCodec

	msgsV1     []proto.Message
	fees       sdk.Coins
	feePayer   []byte
	feeGranter []byte
}

func (w *gogoTxWrapper) String() string { return w.decodedTx.Tx.String() }

var (
	_ authsigning.Tx             = &gogoTxWrapper{}
	_ ante.HasExtensionOptionsTx = &gogoTxWrapper{}
)

// ExtensionOptionsTxBuilder defines a TxBuilder that can also set extensions.
type ExtensionOptionsTxBuilder interface {
	client.TxBuilder

	SetExtensionOptions(...*codectypes.Any)
	SetNonCriticalExtensionOptions(...*codectypes.Any)
}

func (w *gogoTxWrapper) GetMsgs() []sdk.Msg {
	if w.msgsV1 == nil {
		panic("fill in msgs")
	}
	return w.msgsV1
}

func (w *gogoTxWrapper) GetMsgsV2() ([]protov2.Message, error) {
	return w.decodedTx.Messages, nil
}

func (w *gogoTxWrapper) ValidateBasic() error {
	if len(w.decodedTx.Tx.Signatures) == 0 {
		return sdkerrors.ErrNoSignatures.Wrapf("empty signatures")
	}
	if len(w.decodedTx.Signers) != len(w.decodedTx.Tx.Signatures) {
		return sdkerrors.ErrUnauthorized.Wrapf("invalid number of signatures: got %d signatures and %d signers", len(w.decodedTx.Tx.Signatures), len(w.decodedTx.Signers))
	}
	return nil
}

func (w *gogoTxWrapper) GetSigners() ([][]byte, error) {
	return w.decodedTx.Signers, nil
}

func (w *gogoTxWrapper) GetPubKeys() ([]cryptotypes.PubKey, error) {
	signerInfos := w.decodedTx.Tx.AuthInfo.SignerInfos
	pks := make([]cryptotypes.PubKey, len(signerInfos))

	for i, si := range signerInfos {
		// NOTE: it is okay to leave this nil if there is no PubKey in the SignerInfo.
		// PubKey's can be left unset in SignerInfo.
		if si.PublicKey == nil {
			continue
		}
		maybePK, err := decodeFromAny(w.cdc, si.PublicKey)
		if err != nil {
			return nil, err
		}
		pk, ok := maybePK.(cryptotypes.PubKey)
		if ok {
			pks[i] = pk
		} else {
			return nil, errorsmod.Wrapf(sdkerrors.ErrLogic, "expecting pubkey, got: %T", maybePK)
		}
	}

	return pks, nil
}

func (w *gogoTxWrapper) GetGas() uint64 {
	return w.decodedTx.Tx.AuthInfo.Fee.GasLimit
}

func (w *gogoTxWrapper) GetFee() sdk.Coins { return w.fees }

func (w *gogoTxWrapper) FeePayer() []byte { return w.feePayer }

func (w *gogoTxWrapper) FeeGranter() []byte { return w.feeGranter }

func (w *gogoTxWrapper) GetMemo() string { return w.decodedTx.Tx.Body.Memo }

// GetTimeoutHeight returns the transaction's timeout height (if set).
func (w *gogoTxWrapper) GetTimeoutHeight() uint64 { return w.decodedTx.Tx.Body.TimeoutHeight }

// GetUnordered returns the transaction's unordered field (if set).
func (w *gogoTxWrapper) GetUnordered() bool { return w.decodedTx.Tx.Body.Unordered }

// GetSignaturesV2 returns the signatures of the Tx.
func (w *gogoTxWrapper) GetSignaturesV2() ([]signing.SignatureV2, error) {
	signerInfos := w.decodedTx.Tx.AuthInfo.SignerInfos
	sigs := w.decodedTx.Tx.Signatures
	pubKeys, err := w.GetPubKeys()
	if err != nil {
		return nil, err
	}
	n := len(signerInfos)
	res := make([]signing.SignatureV2, n)

	for i, si := range signerInfos {
		// handle nil signatures (in case of simulation)
		if si.ModeInfo == nil || si.ModeInfo.Sum == nil {
			res[i] = signing.SignatureV2{
				PubKey: pubKeys[i],
			}
		} else {
			var err error
			sigData, err := ModeInfoAndSigToSignatureData(si.ModeInfo, sigs[i])
			if err != nil {
				return nil, err
			}
			// sequence number is functionally a transaction nonce and referred to as such in the SDK
			nonce := si.GetSequence()
			res[i] = signing.SignatureV2{
				PubKey:   pubKeys[i],
				Data:     sigData,
				Sequence: nonce,
			}
		}
	}

	return res, nil
}

// GetSigningTxData returns an x/tx/signing.TxData representation of a transaction for use in the signing
// TODO: evaluate if this is even needed considering we have decoded tx.
func (w *gogoTxWrapper) GetSigningTxData() txsigning.TxData {
	return txsigning.TxData{
		Body:                       w.decodedTx.Tx.Body,
		AuthInfo:                   w.decodedTx.Tx.AuthInfo,
		BodyBytes:                  w.decodedTx.TxRaw.BodyBytes,
		AuthInfoBytes:              w.decodedTx.TxRaw.AuthInfoBytes,
		BodyHasUnknownNonCriticals: w.decodedTx.TxBodyHasUnknownNonCriticals,
	}
}

func (w *gogoTxWrapper) GetExtensionOptions() []*codectypes.Any {
	return intoAnyV1(w.decodedTx.Tx.Body.ExtensionOptions)
}

func (w *gogoTxWrapper) GetNonCriticalExtensionOptions() []*codectypes.Any {
	return intoAnyV1(w.decodedTx.Tx.Body.NonCriticalExtensionOptions)
}

func (w *gogoTxWrapper) AsTx() (*txtypes.Tx, error) {
	body := new(txtypes.TxBody)
	authInfo := new(txtypes.AuthInfo)

	err := w.cdc.Unmarshal(w.decodedTx.TxRaw.BodyBytes, body)
	if err != nil {
		return nil, err
	}
	err = w.cdc.Unmarshal(w.decodedTx.TxRaw.AuthInfoBytes, authInfo)
	if err != nil {
		return nil, err
	}
	return &txtypes.Tx{
		Body:       body,
		AuthInfo:   authInfo,
		Signatures: w.decodedTx.TxRaw.Signatures,
	}, nil
}

func (w *gogoTxWrapper) AsTxRaw() (*txtypes.TxRaw, error) {
	return &txtypes.TxRaw{
		BodyBytes:     w.decodedTx.TxRaw.BodyBytes,
		AuthInfoBytes: w.decodedTx.TxRaw.AuthInfoBytes,
		Signatures:    w.decodedTx.TxRaw.Signatures,
	}, nil
}

func intoAnyV1(v2s []*anypb.Any) []*codectypes.Any {
	v1s := make([]*codectypes.Any, len(v2s))
	for i, v2 := range v2s {
		v1s[i] = &codectypes.Any{
			TypeUrl: v2.TypeUrl,
			Value:   v2.Value,
		}
	}
	return v1s
}

// decodeMsgsV1 will decode the given messages into
func decodeMsgsV1(cdc codec.BinaryCodec, anyPBs []*anypb.Any) ([]proto.Message, error) {
	v1s := make([]proto.Message, len(anyPBs))

	for i, anyPB := range anyPBs {
		v1, err := decodeFromAny(cdc, anyPB)
		if err != nil {
			return nil, err
		}
		v1s[i] = v1
	}
	return v1s, nil
}

func decodeFromAny(cdc codec.BinaryCodec, anyPB *anypb.Any) (proto.Message, error) {
	messageName := anyPB.TypeUrl
	if i := strings.LastIndexByte(anyPB.TypeUrl, '/'); i >= 0 {
		messageName = messageName[i+len("/"):]
	}
	typ := proto.MessageType(messageName)
	if typ == nil {
		return nil, fmt.Errorf("cannot find type: %s", anyPB.TypeUrl)
	}
	v1 := reflect.New(typ.Elem()).Interface().(proto.Message)
	err := cdc.Unmarshal(anyPB.Value, v1)
	if err != nil {
		return nil, err
	}
	return v1, nil
}
