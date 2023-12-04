package offChain

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-proto/anyutil"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/math"
	authsigning "cosmossdk.io/x/auth/signing"
	authtx "cosmossdk.io/x/auth/tx"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type builder struct {
	cdc codec.Codec

	tx *apitx.Tx
}

func newBuilder(cdc codec.Codec) *builder {
	return &builder{
		cdc: cdc,
		tx: &apitx.Tx{
			Body: &apitx.TxBody{},
			AuthInfo: &apitx.AuthInfo{
				Fee: &apitx.Fee{
					Amount:   nil,
					GasLimit: 0,
					Payer:    "",
					Granter:  "",
				},
			},
			Signatures: nil,
		},
	}
}

func (b *builder) GetMsgs() []types.Msg {
	msgs := make([]types.Msg, len(b.tx.Body.Messages))
	for i, v := range b.tx.Body.Messages {
		msgs[i] = types.Msg(v)
	}
	return msgs
}

func (b *builder) GetMsgsV2() ([]protov2.Message, error) {
	_, msgs, err := b.getSigners()
	return msgs, err
}

func (b *builder) GetMemo() string {
	return b.tx.Body.Memo
}

func (b *builder) GetGas() uint64 {
	return b.tx.AuthInfo.Fee.GasLimit
}

func (b *builder) GetFee() types.Coins {
	coins := make(types.Coins, len(b.tx.AuthInfo.Fee.Amount))
	for i, v := range b.tx.AuthInfo.Fee.Amount {
		res, ok := math.NewIntFromString(v.Amount)
		if !ok {
			panic("could not convert amount")
		}
		coins[i] = types.Coin{
			Denom:  v.Denom,
			Amount: res,
		}
	}
	return coins
}

func (b *builder) FeePayer() []byte {
	feePayer := b.tx.AuthInfo.Fee.Payer
	if feePayer != "" {
		feePayerAddr, err := b.cdc.InterfaceRegistry().SigningContext().AddressCodec().StringToBytes(feePayer)
		if err != nil {
			panic(err)
		}
		return feePayerAddr
	}
	// use first signer as default if no payer specified
	signers, err := b.GetSigners()
	if err != nil {
		panic(err)
	}

	return signers[0]
}

func (b *builder) FeeGranter() []byte {
	feeGranter := b.tx.AuthInfo.Fee.Granter
	if feeGranter != "" {
		feeGranterAddr, err := b.cdc.InterfaceRegistry().SigningContext().AddressCodec().StringToBytes(feeGranter)
		if err != nil {
			panic(err)
		}

		return feeGranterAddr
	}
	return nil
}

func (b *builder) GetTimeoutHeight() uint64 {
	return b.tx.Body.TimeoutHeight
}

func (b *builder) ValidateBasic() error {
	return nil
}

func (b *builder) GetProtoTx() *apitx.Tx {
	return b.tx
}

func (b *builder) GetSigningTxData() txsigning.TxData {
	body := b.tx.Body
	authInfo := b.tx.AuthInfo

	msgs := make([]*anypb.Any, len(body.Messages))
	for i, msg := range body.Messages {
		msgs[i] = &anypb.Any{
			TypeUrl: msg.TypeUrl,
			Value:   msg.Value,
		}
	}

	extOptions := make([]*anypb.Any, len(body.ExtensionOptions))
	for i, extOption := range body.ExtensionOptions {
		extOptions[i] = &anypb.Any{
			TypeUrl: extOption.TypeUrl,
			Value:   extOption.Value,
		}
	}

	nonCriticalExtOptions := make([]*anypb.Any, len(body.NonCriticalExtensionOptions))
	for i, extOption := range body.NonCriticalExtensionOptions {
		nonCriticalExtOptions[i] = &anypb.Any{
			TypeUrl: extOption.TypeUrl,
			Value:   extOption.Value,
		}
	}

	feeCoins := authInfo.Fee.Amount
	feeAmount := make([]*basev1beta1.Coin, len(feeCoins))
	for i, coin := range feeCoins {
		feeAmount[i] = &basev1beta1.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount,
		}
	}

	txSignerInfos := make([]*apitx.SignerInfo, len(authInfo.SignerInfos))
	for i, signerInfo := range authInfo.SignerInfos {
		txSignerInfo := &apitx.SignerInfo{
			PublicKey: &anypb.Any{
				TypeUrl: signerInfo.PublicKey.TypeUrl,
				Value:   signerInfo.PublicKey.Value,
			},
			Sequence: signerInfo.Sequence,
			ModeInfo: signerInfo.ModeInfo,
		}
		txSignerInfos[i] = txSignerInfo
	}

	txAuthInfo := &apitx.AuthInfo{
		SignerInfos: txSignerInfos,
		Fee: &apitx.Fee{
			Amount:   feeAmount,
			GasLimit: authInfo.Fee.GasLimit,
			Payer:    authInfo.Fee.Payer,
			Granter:  authInfo.Fee.Granter,
		},
	}

	txBody := &apitx.TxBody{
		Messages:                    msgs,
		Memo:                        body.Memo,
		TimeoutHeight:               body.TimeoutHeight,
		ExtensionOptions:            extOptions,
		NonCriticalExtensionOptions: nonCriticalExtOptions,
	}
	authInfoBz, err := protov2.Marshal(b.tx.AuthInfo)
	if err != nil {
		panic(err)
	}
	bodyBz, err := protov2.Marshal(b.tx.Body)
	if err != nil {
		panic(err)
	}
	txData := txsigning.TxData{
		AuthInfo:      txAuthInfo,
		AuthInfoBytes: authInfoBz,
		Body:          txBody,
		BodyBytes:     bodyBz,
	}
	return txData
}

func (b *builder) GetTx() authsigning.Tx {
	return b
}

func (b *builder) GetPubKeys() ([]cryptotypes.PubKey, error) { // If signer already has pubkey in context, this list will have nil in its place
	signerInfos := b.tx.AuthInfo.SignerInfos
	pks := make([]cryptotypes.PubKey, len(signerInfos))

	for i, si := range signerInfos {
		// NOTE: it is okay to leave this nil if there is no PubKey in the SignerInfo.
		// PubKey's can be left unset in SignerInfo.
		if si.PublicKey == nil {
			continue
		}
		var pk cryptotypes.PubKey
		anyPk := &codectypes.Any{
			TypeUrl: si.PublicKey.TypeUrl,
			Value:   si.PublicKey.Value,
		}
		err := b.cdc.UnpackAny(anyPk, &pk)
		if err != nil {
			return nil, err
		}
		pks[i] = pk
	}

	return pks, nil
}

func (b *builder) GetSignaturesV2() ([]signing.SignatureV2, error) {
	signerInfos := b.tx.AuthInfo.SignerInfos
	sigs := b.tx.Signatures
	pubKeys, err := b.GetPubKeys()
	if err != nil {
		return nil, err
	}
	n := len(signerInfos)
	res := make([]signing.SignatureV2, n)

	for i, si := range signerInfos {
		// handle nil signatures (in case of simulation)
		if si.ModeInfo == nil {
			res[i] = signing.SignatureV2{
				PubKey: pubKeys[i],
			}
		} else {
			var err error
			sigData, err := modeInfoAndSigToSignatureData(si.ModeInfo, sigs[i])
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

func (b *builder) GetSigners() ([][]byte, error) {
	signers, _, err := b.getSigners()
	return signers, err
}

func (b *builder) getSigners() ([][]byte, []protov2.Message, error) {
	var signers [][]byte
	seen := map[string]bool{}

	var msgsv2 []protov2.Message
	for _, msg := range b.tx.Body.Messages {
		msgv2, err := anyutil.Unpack(msg, b.cdc.InterfaceRegistry(), nil)
		if err != nil {
			return nil, nil, err
		}
		xs, err := b.cdc.InterfaceRegistry().SigningContext().GetSigners(msgv2)
		if err != nil {
			return nil, nil, err
		}

		msgsv2 = append(msgsv2, msg)

		for _, signer := range xs {
			if !seen[string(signer)] {
				signers = append(signers, signer)
				seen[string(signer)] = true
			}
		}
	}

	// ensure any specified fee payer is included in the required signers (at the end)
	feePayer := b.tx.AuthInfo.Fee.Payer
	var feePayerAddr []byte
	if feePayer != "" {
		var err error
		if err != nil {
			return nil, nil, err
		}
	}
	if feePayerAddr != nil && !seen[string(feePayerAddr)] {
		signers = append(signers, feePayerAddr)
		seen[string(feePayerAddr)] = true
	}

	return signers, msgsv2, nil
}

func (b *builder) setMsgs(msgs ...types.Msg) error {
	anys := make([]*anypb.Any, len(msgs))
	for i, msg := range msgs {
		protoMsg, ok := msg.(protov2.Message)
		if !ok {
			return errors.New("")
		}
		protov2MarshalOpts := protov2.MarshalOptions{Deterministic: true}
		bz, err := protov2MarshalOpts.Marshal(protoMsg)
		if err != nil {
			return err
		}
		anys[i] = &anypb.Any{
			TypeUrl: types.MsgTypeURL(msg),
			Value:   bz,
		}
	}
	b.tx.Body.Messages = anys
	return nil
}

func (b *builder) SetSignatures(signatures ...signing.SignatureV2) error {
	n := len(signatures)
	signerInfos := make([]*apitx.SignerInfo, n)
	rawSigs := make([][]byte, n)

	for i, sig := range signatures {
		var mi *typestx.ModeInfo
		mi, rawSigs[i] = authtx.SignatureDataToModeInfoAndSig(sig.Data)
		modeinfo, err := castModeInfo(mi)
		if err != nil {
			return err
		}
		pubKey, err := codectypes.NewAnyWithValue(sig.PubKey)
		if err != nil {
			return err
		}
		signerInfos[i] = &apitx.SignerInfo{
			PublicKey: &anypb.Any{
				TypeUrl: pubKey.TypeUrl,
				Value:   pubKey.Value,
			},
			ModeInfo: modeinfo,
			Sequence: sig.Sequence,
		}
	}

	b.tx.AuthInfo.SignerInfos = signerInfos
	b.tx.Signatures = rawSigs

	return nil
}

// TODO: cast multisig
func castModeInfo(modeinfo *typestx.ModeInfo) (*apitx.ModeInfo, error) {
	mi, ok := modeinfo.GetSum().(*typestx.ModeInfo_Single_)
	if !ok {
		return nil, errors.New("")
	}
	return &apitx.ModeInfo{
		Sum: &apitx.ModeInfo_Single_{
			Single: &apitx.ModeInfo_Single{
				Mode: apitxsigning.SignMode(mi.Single.Mode),
			},
		},
	}, nil
}

// modeInfoAndSigToSignatureData converts a ModeInfo and raw bytes signature to a SignatureData
func modeInfoAndSigToSignatureData(modeInfo *apitx.ModeInfo, sig []byte) (signing.SignatureData, error) {
	switch modeInfoType := modeInfo.Sum.(type) {
	case *apitx.ModeInfo_Single_:
		return &signing.SingleSignatureData{
			SignMode:  signing.SignMode(modeInfoType.Single.Mode),
			Signature: sig,
		}, nil

	default:
		panic(fmt.Errorf("unexpected ModeInfo data type %T", modeInfo))
	}
}
