package aminojson_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-proto/anyutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/secp256k1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
)

type handlerArgumentOptions struct {
	chainID string
	memo    string
	msg     proto.Message
	accNum  uint64
	accSeq  uint64
	tip     *txv1beta1.Tip
}

func makeHandlerArguments(options handlerArgumentOptions) (signing.SignerData, signing.TxData, error) {
	pk := &secp256k1.PubKey{
		Key: make([]byte, 256),
	}
	anyPk, err := anyutil.New(pk)
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	signerInfo := []*txv1beta1.SignerInfo{
		{
			PublicKey: anyPk,
			ModeInfo: &txv1beta1.ModeInfo{
				Sum: &txv1beta1.ModeInfo_Single_{
					Single: &txv1beta1.ModeInfo_Single{
						Mode: signingv1beta1.SignMode_SIGN_MODE_DIRECT_AUX,
					},
				},
			},
			Sequence: options.accSeq,
		},
	}

	fee := &txv1beta1.Fee{
		Amount: []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
	}
	if options.tip == nil {
		options.tip = &txv1beta1.Tip{Tipper: "tipper", Amount: []*basev1beta1.Coin{{Denom: "tip-token", Amount: "10"}}}
	}

	anyMsg, err := anyutil.New(options.msg)
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	txBody := &txv1beta1.TxBody{
		Messages: []*anypb.Any{anyMsg},
		Memo:     options.memo,
	}

	authInfo := &txv1beta1.AuthInfo{
		Fee:         fee,
		Tip:         options.tip,
		SignerInfos: signerInfo,
	}

	bodyBz, err := proto.Marshal(txBody)
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}
	authInfoBz, err := proto.Marshal(authInfo)
	if err != nil {
		return signing.SignerData{}, signing.TxData{}, err
	}

	txData := signing.TxData{
		Body:          txBody,
		AuthInfo:      authInfo,
		AuthInfoBytes: authInfoBz,
		BodyBytes:     bodyBz,
	}

	signerData := signing.SignerData{
		ChainId:       options.chainID,
		AccountNumber: options.accNum,
		Sequence:      options.accSeq,
		Address:       "signerAddr",
		PubKey:        anyPk,
	}

	return signerData, txData, nil
}

func TestAminoJsonSignMode(t *testing.T) {
	handlerOptions := handlerArgumentOptions{
		chainID: "test-chain",
		memo:    "sometestmemo",
		msg: &bankv1beta1.MsgSend{
			FromAddress: "foo",
			ToAddress:   "bar",
			Amount:      []*basev1beta1.Coin{{Denom: "demon", Amount: "100"}},
		},
		accNum: 1,
		accSeq: 2,
	}

	signerData, txData, err := makeHandlerArguments(handlerOptions)
	require.NoError(t, err)

	handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
	_, err = handler.GetSignBytes(context.Background(), signerData, txData)
	require.NoError(t, err)
}
