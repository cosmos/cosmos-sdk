package tx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestAuxTxBuilder(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	testdata.RegisterInterfaces(encCfg.InterfaceRegistry)

	var b tx.AuxTxBuilder
	_, pk, addr := testdata.KeyTestPubAddr()
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)
	msg := testdata.NewTestMsg(addr)
	msgAny, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)
	tip := &typestx.Tip{Tipper: addr.String(), Amount: testdata.NewTestFeeAmount()}
	sig := []byte{42}

	testcases := []struct {
		name      string
		malleate  func() error
		expErr    bool
		expErrStr string
	}{
		{
			"cannot set SIGN_MODE_DIRECT",
			func() error {
				return b.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT)
			},
			true, "AuxTxBuilder can only sign with SIGN_MODE_DIRECT_AUX or SIGN_MODE_LEGACY_AMINO_JSON",
		},
		{
			"cannot set invalid pubkey",
			func() error {
				return b.SetPubKey(cryptotypes.PubKey(nil))
			},
			true, "failed packing protobuf message to Any",
		},
		{
			"cannot set invalid Msg",
			func() error {
				return b.SetMsgs(sdk.Msg(nil))
			},
			true, "failed packing protobuf message to Any",
		},
		{
			"GetSignBytes body should not be nil",
			func() error {
				_, err := b.GetSignBytes()
				return err
			},
			true, "aux tx is nil, call setters on AuxTxBuilder first",
		},
		{
			"GetSignBytes pubkey should not be nil",
			func() error {
				b.SetMsgs(msg)

				_, err := b.GetSignBytes()
				return err
			},
			true, "public key cannot be empty: invalid pubkey",
		},
		{
			"GetSignBytes invalid sign mode",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)

				_, err := b.GetSignBytes()
				return err
			},
			true, "got unknown sign mode SIGN_MODE_UNSPECIFIED",
		},
		{
			"GetSignBytes tipper should not be nil (if tip is set)",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(&typestx.Tip{Tipper: addr.String()})
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				return err
			},
			true, "tip amount cannot be empty",
		},
		{
			"GetSignBytes works for DIRECT_AUX",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(tip)
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				return err
			},
			false, "",
		},
		{
			"GetAuxSignerData address should not be empty",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(tip)
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				require.NoError(t, err)

				_, err = b.GetAuxSignerData()
				return err
			},
			true, "address cannot be empty: invalid request",
		},
		{
			"GetAuxSignerData signature should not be empty",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(tip)
				b.SetAddress(addr.String())
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				require.NoError(t, err)

				_, err = b.GetAuxSignerData()
				return err
			},
			true, "signature cannot be empty: no signatures supplied",
		},
		{
			"GetAuxSignerData works for DIRECT_AUX",
			func() error {
				memo := "test-memo"
				chainID := "test-chain"

				b.SetAccountNumber(1)
				b.SetSequence(2)
				b.SetTimeoutHeight(3)
				b.SetMemo(memo)
				b.SetChainID(chainID)
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(tip)
				b.SetAddress(addr.String())
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				require.NoError(t, err)
				b.SetSignature(sig)

				auxSignerData, err := b.GetAuxSignerData()
				require.NoError(t, err)

				// Make sure auxSignerData is correctly populated
				var body typestx.TxBody
				err = encCfg.Codec.Unmarshal(auxSignerData.SignDoc.BodyBytes, &body)
				require.NoError(t, err)

				require.Equal(t, uint64(1), auxSignerData.SignDoc.AccountNumber)
				require.Equal(t, uint64(2), auxSignerData.SignDoc.Sequence)
				require.Equal(t, uint64(3), body.TimeoutHeight)
				require.Equal(t, memo, body.Memo)
				require.Equal(t, chainID, auxSignerData.SignDoc.ChainId)
				require.Equal(t, msgAny, body.GetMessages()[0])
				require.Equal(t, pkAny, auxSignerData.SignDoc.PublicKey)
				require.Equal(t, tip, auxSignerData.SignDoc.Tip)
				require.Equal(t, signing.SignMode_SIGN_MODE_DIRECT_AUX, auxSignerData.Mode)
				require.Equal(t, sig, auxSignerData.Sig)

				return err
			},
			false, "",
		},
		{
			"GetSignBytes works for LEGACY_AMINO_JSON",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(tip)
				b.SetAddress(addr.String())
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				return err
			},
			false, "",
		},
		{
			"GetAuxSignerData works for LEGACY_AMINO_JSON",
			func() error {
				b.SetMsgs(msg)
				b.SetPubKey(pk)
				b.SetTip(tip)
				b.SetAddress(addr.String())
				err := b.SetSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
				require.NoError(t, err)

				_, err = b.GetSignBytes()
				require.NoError(t, err)
				b.SetSignature(sig)

				_, err = b.GetAuxSignerData()
				return err
			},
			false, "",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b = tx.NewAuxTxBuilder()
			err := tc.malleate()

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrStr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
