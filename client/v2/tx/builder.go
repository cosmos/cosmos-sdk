package tx

import (
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

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

// TxBuilder defines an interface which an application-defined concrete transaction
// type must implement. Namely, it must be able to set messages, generate
// signatures, and provide canonical bytes to sign over. The transaction must
// also know how to encode itself.
type TxBuilder interface {
	GetTx() (Tx, error)
	GetSigningTxData() (*signing.TxData, error)

	SetMsgs(...transaction.Msg) error
	SetMemo(string)
	SetFeeAmount([]*base.Coin)
	SetGasLimit(uint64)
	SetTimeoutTimestamp(time.Time)
	SetFeePayer(string) error
	SetFeeGranter(string) error
	SetUnordered(bool)
	SetSignatures(...Signature) error
}

// TxBuilderProvider provides a TxBuilder.
type TxBuilderProvider interface {
	NewTxBuilder() TxBuilder
}

// BuilderProvider implements TxBuilderProvider.
type BuilderProvider struct {
	addressCodec address.Codec
	decoder      Decoder
	codec        codec.BinaryCodec
}

// NewBuilderProvider BuilderProvider constructor.
func NewBuilderProvider(addressCodec address.Codec, decoder Decoder, codec codec.BinaryCodec) *BuilderProvider {
	return &BuilderProvider{
		addressCodec: addressCodec,
		decoder:      decoder,
		codec:        codec,
	}
}

// NewTxBuilder TxBuilder constructor.
func (b BuilderProvider) NewTxBuilder() TxBuilder {
	return newTxBuilder(b.addressCodec, b.decoder, b.codec)
}

type txBuilder struct {
	addressCodec address.Codec
	decoder      Decoder
	codec        codec.BinaryCodec

	msgs             []transaction.Msg
	timeoutHeight    uint64
	timeoutTimestamp time.Time
	granter          []byte
	payer            []byte
	unordered        bool
	memo             string
	gasLimit         uint64
	fees             []*base.Coin
	signerInfos      []*apitx.SignerInfo
	signatures       [][]byte

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

// GetTx converts txBuilder messages to V2 and returns a Tx.
func (b *txBuilder) GetTx() (Tx, error) {
	return b.getTx()
}

func (b *txBuilder) getTx() (*wrappedTx, error) {
	msgs, err := msgsV1toAnyV2(b.msgs)
	if err != nil {
		return nil, err
	}

	body := &apitx.TxBody{
		Messages:                    msgs,
		Memo:                        b.memo,
		TimeoutHeight:               b.timeoutHeight,
		TimeoutTimestamp:            timestamppb.New(b.timeoutTimestamp),
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

	bodyBytes, err := marshalOption.Marshal(body)
	if err != nil {
		return nil, err
	}

	authInfoBytes, err := marshalOption.Marshal(authInfo)
	if err != nil {
		return nil, err
	}

	txRawBytes, err := marshalOption.Marshal(&apitx.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    b.signatures,
	})
	if err != nil {
		return nil, err
	}

	decodedTx, err := b.decoder.Decode(txRawBytes)
	if err != nil {
		return nil, err
	}

	return newWrapperTx(b.codec, decodedTx), nil
}

// getFee computes the transaction fee information for the txBuilder.
// It returns a pointer to an apitx.Fee struct containing the fee amount, gas limit, payer, and granter information.
// If the granter or payer addresses are set, it converts them from bytes to string using the addressCodec.
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

// GetSigningTxData returns a TxData with the txBuilder info.
func (b *txBuilder) GetSigningTxData() (*signing.TxData, error) {
	tx, err := b.getTx()
	if err != nil {
		return nil, err
	}

	return &signing.TxData{
		Body:                       tx.Tx.Body,
		AuthInfo:                   tx.Tx.AuthInfo,
		BodyBytes:                  tx.TxRaw.BodyBytes,
		AuthInfoBytes:              tx.TxRaw.AuthInfoBytes,
		BodyHasUnknownNonCriticals: tx.TxBodyHasUnknownNonCriticals,
	}, nil
}

// SetMsgs sets the messages for the transaction.
func (b *txBuilder) SetMsgs(msgs ...transaction.Msg) error {
	b.msgs = msgs
	return nil
}

// SetMemo sets the memo for the transaction.
func (b *txBuilder) SetMemo(memo string) {
	b.memo = memo
}

// SetFeeAmount sets the fee amount for the transaction.
func (b *txBuilder) SetFeeAmount(coins []*base.Coin) {
	b.fees = coins
}

// SetGasLimit sets the gas limit for the transaction.
func (b *txBuilder) SetGasLimit(gasLimit uint64) {
	b.gasLimit = gasLimit
}

// SetTimeoutTimestamp sets the timeout timestamp for the transaction.
func (b *txBuilder) SetTimeoutTimestamp(timeoutHeight time.Time) {
	b.timeoutTimestamp = timeoutHeight
}

// SetFeePayer sets the fee payer for the transaction.
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

// SetFeeGranter sets the fee granter's address in the transaction builder.
// If the feeGranter string is empty, the function returns nil without setting an address.
// It converts the feeGranter string to bytes using the address codec and sets it as the granter address.
// Returns an error if the conversion fails.
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

// SetUnordered sets the unordered flag of the transaction builder.
func (b *txBuilder) SetUnordered(unordered bool) {
	b.unordered = unordered
}

// SetSignatures sets the signatures for the transaction builder.
// It takes a variable number of Signature arguments and processes each one to extract the mode information and raw signature.
// It also converts the public key to the appropriate format and sets the signer information.
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

// msgsV1toAnyV2 converts a slice of transaction.Msg (v1) to a slice of anypb.Any (v2).
// It first converts each transaction.Msg into a codectypes.Any and then converts
// these into anypb.Any.
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

// intoAnyV2 converts a slice of codectypes.Any (v1) to a slice of anypb.Any (v2).
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
