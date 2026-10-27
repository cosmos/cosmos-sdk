package tx

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestDecodeMultisignatures(t *testing.T) {
	testSigs := [][]byte{
		[]byte("dummy1"),
		[]byte("dummy2"),
		[]byte("dummy3"),
	}

	badMultisig := testdata.BadMultiSignature{
		Signatures:     testSigs,
		MaliciousField: []byte("bad stuff..."),
	}
	bz, err := badMultisig.Marshal()
	require.NoError(t, err)

	_, err = decodeMultisignatures(bz)
	require.Error(t, err)

	goodMultisig := types.MultiSignature{
		Signatures: testSigs,
	}
	bz, err = goodMultisig.Marshal()
	require.NoError(t, err)

	decodedSigs, err := decodeMultisignatures(bz)
	require.NoError(t, err)

	require.Equal(t, testSigs, decodedSigs)
}

func TestModeInfoAndSigToSignatureDataMultisigCountMismatch(t *testing.T) {
	single := &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Single_{Single: &txtypes.ModeInfo_Single{Mode: signingtypes.SignMode_SIGN_MODE_DIRECT}}}
	multi := &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Multi_{Multi: &txtypes.ModeInfo_Multi{
		ModeInfos: []*txtypes.ModeInfo{single, single}, // two nested mode infos
	}}}

	// MultiSignature carries only one sub-signature, fewer than ModeInfos.
	msig := types.MultiSignature{Signatures: [][]byte{[]byte("onesig")}}
	bz, err := msig.Marshal()
	require.NoError(t, err)

	_, err = ModeInfoAndSigToSignatureData(multi, bz)
	require.Error(t, err)
}

func TestGetSignaturesV2SignerSignatureCountMismatch(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	decoder := DefaultTxDecoder(cdc)

	single := &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Single_{Single: &txtypes.ModeInfo_Single{Mode: signingtypes.SignMode_SIGN_MODE_DIRECT}}}
	authInfo := &txtypes.AuthInfo{
		SignerInfos: []*txtypes.SignerInfo{
			{ModeInfo: single},
			{ModeInfo: single},
		},
		Fee: &txtypes.Fee{GasLimit: 1},
	}
	authInfoBz, err := authInfo.Marshal()
	require.NoError(t, err)

	bodyBz, err := (&txtypes.TxBody{}).Marshal()
	require.NoError(t, err)

	raw := &txtypes.TxRaw{
		BodyBytes:     bodyBz,
		AuthInfoBytes: authInfoBz,
		Signatures:    [][]byte{[]byte("onlyonesig")}, // one signature, two signer infos
	}
	txBz, err := raw.Marshal()
	require.NoError(t, err)

	decoded, err := decoder(txBz)
	require.NoError(t, err)

	_, err = decoded.(interface {
		GetSignaturesV2() ([]signingtypes.SignatureV2, error)
	}).GetSignaturesV2()
	require.Error(t, err)
}
