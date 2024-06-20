package tx

import (
	gogoany "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/types/known/anypb"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var (
	_ TxBuilder         = &txBuilder{}
	_ TxBuilderProvider = BuilderProvider{}
)

type ExtendedTxBuilder interface {
	SetExtensionOptions(...*gogoany.Any)
}

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() (*apitx.Tx, error)
	GetSigningTxData() (*signing.TxData, error) // TODO: check this

	SetMsgs(...transaction.Msg) error
	SetMemo(string)
	SetFeeAmount([]*base.Coin)
	SetGasLimit(uint64)
	SetTimeoutHeight(uint64)
	SetFeePayer(string) error
	SetFeeGranter(string) error
	SetUnordered(bool)
	SetSignatures(...Signature) error
}

type TxBuilderProvider interface {
	NewTxBuilder() TxBuilder
	WrapTxBuilder(*apitx.Tx) (TxBuilder, error)
}

type BuilderProvider struct {
	addressCodec address.Codec
	decoder      Decoder
	codec        codec.BinaryCodec
}

func NewBuilderProvider(addressCodec address.Codec, decoder Decoder, codec codec.BinaryCodec) *BuilderProvider {
	return &BuilderProvider{
		addressCodec: addressCodec,
		decoder:      decoder,
		codec:        codec,
	}
}

func (b BuilderProvider) NewTxBuilder() TxBuilder {
	return newTxBuilder(b.addressCodec, b.decoder, b.codec)
}

// TODO: work on this
func (b BuilderProvider) WrapTxBuilder(tx *apitx.Tx) (TxBuilder, error) {
	return &txBuilder{
		addressCodec: b.addressCodec,
		decoder:      b.decoder,
		codec:        b.codec,
	}, nil
}

type txBuilder struct {
	addressCodec address.Codec
	decoder      Decoder
	codec        codec.BinaryCodec

	msgs          []transaction.Msg
	timeoutHeight uint64
	granter       []byte
	payer         []byte
	unordered     bool
	memo          string
	gasLimit      uint64
	fees          []*base.Coin
	signerInfos   []*apitx.SignerInfo
	signatures    [][]byte

	extensionOptions            []*anypb.Any
	nonCriticalExtensionOptions []*anypb.Any
}

func newTxBuilder(addressCodec address.Codec, decoder Decoder, codec codec.BinaryCodec) *txBuilder {
	return &txBuilder{
		addressCodec: addressCodec,
		decoder:      decoder,
		codec:        codec,
	}
}

func (b *txBuilder) GetTx() (*apitx.Tx, error) {
	msgs, err := msgsV1toAnyV2(b.msgs)
	if err != nil {
		return nil, err
	}

	body := &apitx.TxBody{
		Messages:                    msgs,
		Memo:                        b.memo,
		TimeoutHeight:               b.timeoutHeight,
		Unordered:                   b.unordered,
		ExtensionOptions:            b.extensionOptions,
		NonCriticalExtensionOptions: b.nonCriticalExtensionOptions,
	}

	fee, err := b.getFee()
	if err != nil {
		return nil, err
	}

	authInfo := &apitx.AuthInfo{
		SignerInfos: b.signerInfos,
		Fee:         fee,
	}

	return &apitx.Tx{
		Body:       body,
		AuthInfo:   authInfo,
		Signatures: b.signatures,
	}, nil
}

func (b *txBuilder) getFee() (fee *apitx.Fee, err error) {
	granterStr := ""
	if b.granter != nil {
		granterStr, err = b.addressCodec.BytesToString(b.granter)
		if err != nil {
			return nil, err
		}
	}

	payerStr := ""
	if b.payer != nil {
		payerStr, err = b.addressCodec.BytesToString(b.payer)
		if err != nil {
			return nil, err
		}
	}

	fee = &apitx.Fee{
		Amount:   b.fees,
		GasLimit: b.gasLimit,
		Payer:    payerStr,
		Granter:  granterStr,
	}

	return fee, nil
}

func (b *txBuilder) GetSigningTxData() (*signing.TxData, error) {
	tx, err := b.GetTx()
	if err != nil {
		return nil, err
	}

	bodyBytes, err := marshalOption.Marshal(tx.Body)
	if err != nil {
		return nil, err
	}
	authBytes, err := marshalOption.Marshal(tx.AuthInfo)
	if err != nil {
		return nil, err
	}

	rawTx, err := marshalOption.Marshal(&apitx.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authBytes,
		Signatures:    b.signatures,
	})
	if err != nil {
		return nil, err
	}

	decodedTx, err := b.decoder.Decode(rawTx)
	if err != nil {
		return nil, err
	}

	return &signing.TxData{
		Body:                       decodedTx.Tx.Body,
		AuthInfo:                   decodedTx.Tx.AuthInfo,
		BodyBytes:                  decodedTx.TxRaw.BodyBytes,
		AuthInfoBytes:              decodedTx.TxRaw.AuthInfoBytes,
		BodyHasUnknownNonCriticals: decodedTx.TxBodyHasUnknownNonCriticals,
	}, nil
}

func (b *txBuilder) SetMsgs(msgs ...transaction.Msg) error {
	b.msgs = msgs
	return nil
}

func (b *txBuilder) SetMemo(memo string) {
	b.memo = memo
}

func (b *txBuilder) SetFeeAmount(coins []*base.Coin) {
	b.fees = coins
}

func (b *txBuilder) SetGasLimit(gasLimit uint64) {
	b.gasLimit = gasLimit
}

func (b *txBuilder) SetTimeoutHeight(timeoutHeight uint64) {
	b.timeoutHeight = timeoutHeight
}

func (b *txBuilder) SetFeePayer(feePayer string) error {
	if feePayer == "" {
		return nil
	}

	addr, err := b.addressCodec.StringToBytes(feePayer)
	if err != nil {
		return err
	}
	b.payer = addr
	return nil
}

func (b *txBuilder) SetFeeGranter(feeGranter string) error {
	if feeGranter == "" {
		return nil
	}

	addr, err := b.addressCodec.StringToBytes(feeGranter)
	if err != nil {
		return err
	}
	b.granter = addr

	return nil
}

func (b *txBuilder) SetUnordered(unordered bool) {
	b.unordered = unordered
}

func (b *txBuilder) SetSignatures(signatures ...Signature) error {
	n := len(signatures)
	signerInfos := make([]*apitx.SignerInfo, n)
	rawSignatures := make([][]byte, n)

	for i, sig := range signatures {
		var (
			modeInfo *apitx.ModeInfo
			pubKey   *codectypes.Any
			err      error
			anyPk    *anypb.Any
		)

		modeInfo, rawSignatures[i] = SignatureDataToModeInfoAndSig(sig.Data)
		if sig.PubKey != nil {
			pubKey, err = codectypes.NewAnyWithValue(sig.PubKey)
			if err != nil {
				return err
			}
			anyPk = &anypb.Any{
				TypeUrl: pubKey.TypeUrl,
				Value:   pubKey.Value,
			}
		}

		signerInfos[i] = &apitx.SignerInfo{
			PublicKey: anyPk,
			ModeInfo:  modeInfo,
			Sequence:  sig.Sequence,
		}
	}

	b.signerInfos = signerInfos
	b.signatures = rawSignatures

	return nil
}

func msgsV1toAnyV2(msgs []transaction.Msg) ([]*anypb.Any, error) {
	anys := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		anys[i] = anyMsg
	}

	return intoAnyV2(anys), nil
}

func intoAnyV2(v1s []*codectypes.Any) []*anypb.Any {
	v2s := make([]*anypb.Any, len(v1s))
	for i, v1 := range v1s {
		v2s[i] = &anypb.Any{
			TypeUrl: v1.TypeUrl,
			Value:   v1.Value,
		}
	}
	return v2s
}
