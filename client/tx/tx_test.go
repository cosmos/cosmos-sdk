package tx_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func NewTestTxConfig() client.TxConfig {
	cfg := simapp.MakeTestEncodingConfig()
	return cfg.TxConfig
}

func TestCalculateGas(t *testing.T) {
	makeQueryFunc := func(gasUsed uint64, wantErr bool) func(string, []byte) ([]byte, int64, error) {
		return func(string, []byte) ([]byte, int64, error) {
			if wantErr {
				return nil, 0, errors.New("query failed")
			}
			simRes := &txtypes.SimulateResponse{
				GasInfo: &sdk.GasInfo{GasUsed: gasUsed, GasWanted: gasUsed},
				Result:  &sdk.Result{Data: []byte("tx data"), Log: "log"},
			}

			bz, err := simRes.Marshal()
			if err != nil {
				return nil, 0, err
			}

			return bz, 0, nil
		}
	}

	type args struct {
		queryFuncGasUsed uint64
		queryFuncWantErr bool
		adjustment       float64
	}

	testCases := []struct {
		name         string
		args         args
		wantEstimate uint64
		wantAdjusted uint64
		expPass      bool
	}{
		{"error", args{0, true, 1.2}, 0, 0, false},
		{"adjusted gas", args{10, false, 1.2}, 10, 12, true},
	}

	for _, tc := range testCases {
		stc := tc
		txCfg := NewTestTxConfig()

		txf := tx.Factory{}.
			WithChainID("test-chain").
			WithTxConfig(txCfg).WithSignMode(txCfg.SignModeHandler().DefaultMode())

		t.Run(stc.name, func(t *testing.T) {
			queryFunc := makeQueryFunc(stc.args.queryFuncGasUsed, stc.args.queryFuncWantErr)
			simRes, gotAdjusted, err := tx.CalculateGas(queryFunc, txf.WithGasAdjustment(stc.args.adjustment))
			if stc.expPass {
				require.NoError(t, err)
				require.Equal(t, simRes.GasInfo.GasUsed, stc.wantEstimate)
				require.Equal(t, gotAdjusted, stc.wantAdjusted)
				require.NotNil(t, simRes.Result)
			} else {
				require.Error(t, err)
				require.Nil(t, simRes.Result)
			}
		})
	}
}

func TestBuildSimTx(t *testing.T) {
	txCfg := NewTestTxConfig()

	txf := tx.Factory{}.
		WithTxConfig(txCfg).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain").
		WithSignMode(txCfg.SignModeHandler().DefaultMode())

	msg := banktypes.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	bz, err := tx.BuildSimTx(txf, msg)
	require.NoError(t, err)
	require.NotNil(t, bz)
}

func TestBuildUnsignedTx(t *testing.T) {
	txf := tx.Factory{}.
		WithTxConfig(NewTestTxConfig()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")

	msg := banktypes.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	tx, err := tx.BuildUnsignedTx(txf, msg)
	require.NoError(t, err)
	require.NotNil(t, tx)

	sigs, err := tx.GetTx().(signing.SigVerifiableTx).GetSignaturesV2()
	require.NoError(t, err)
	require.Empty(t, sigs)
}

func TestSign(t *testing.T) {
	requireT := require.New(t)
	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(t.Name(), "test", t.TempDir(), nil)
	requireT.NoError(err)

	var from1 = "test_key1"
	var from2 = "test_key2"

	// create a new key using a mnemonic generator and test if we can reuse seed to recreate that account
	_, seed, err := kr.NewMnemonic(from1, keyring.English, path, hd.Secp256k1)
	requireT.NoError(err)
	requireT.NoError(kr.Delete(from1))
	info1, _, err := kr.NewMnemonic(from1, keyring.English, path, hd.Secp256k1)
	requireT.NoError(err)

	info2, err := kr.NewAccount(from2, seed, "", path, hd.Secp256k1)
	requireT.NoError(err)

	pubKey1 := info1.GetPubKey()
	pubKey2 := info2.GetPubKey()
	requireT.NotEqual(pubKey1.Bytes(), pubKey2.Bytes())
	t.Log("Pub keys:", pubKey1, pubKey2)

	txfNoKeybase := tx.Factory{}.
		WithTxConfig(NewTestTxConfig()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")
	txfDirect := txfNoKeybase.
		WithKeybase(kr).
		WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)
	txfAmino := txfDirect.
		WithSignMode(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	msg1 := banktypes.NewMsgSend(info1.GetAddress(), sdk.AccAddress("to"), nil)
	msg2 := banktypes.NewMsgSend(info2.GetAddress(), sdk.AccAddress("to"), nil)
	txb, err := tx.BuildUnsignedTx(txfNoKeybase, msg1, msg2)
	requireT.NoError(err)
	txb2, err := tx.BuildUnsignedTx(txfNoKeybase, msg1, msg2)
	requireT.NoError(err)
	txbSimple, err := tx.BuildUnsignedTx(txfNoKeybase, msg2)
	requireT.NoError(err)

	testCases := []struct {
		name         string
		txf          tx.Factory
		txb          client.TxBuilder
		from         string
		overwrite    bool
		expectedPKs  []cryptotypes.PubKey
		matchingSigs []int // if not nil, check matching signature against old ones.
	}{
		{"should fail if txf without keyring",
			txfNoKeybase, txb, from1, true, nil, nil},
		{"should fail for non existing key",
			txfAmino, txb, "unknown", true, nil, nil},
		{"amino: should succeed with keyring",
			txfAmino, txbSimple, from1, true, []cryptotypes.PubKey{pubKey1}, nil},
		{"direct: should succeed with keyring",
			txfDirect, txbSimple, from1, true, []cryptotypes.PubKey{pubKey1}, nil},

		/**** test double sign Amino mode ****/
		{"amino: should sign multi-signers tx",
			txfAmino, txb, from1, true, []cryptotypes.PubKey{pubKey1}, nil},
		{"amino: should append a second signature and not overwrite",
			txfAmino, txb, from2, false, []cryptotypes.PubKey{pubKey1, pubKey2}, []int{0, 0}},
		{"amino: should overwrite a signature",
			txfAmino, txb, from2, true, []cryptotypes.PubKey{pubKey2}, []int{1, 0}},

		/**** test double sign Direct mode
		  signing transaction with more than 2 signers should fail in DIRECT mode ****/
		{"direct: should fail to append a signature with different mode",
			txfDirect, txb, from1, false, []cryptotypes.PubKey{}, nil},
		{"direct: should fail to sign multi-signers tx",
			txfDirect, txb2, from1, false, []cryptotypes.PubKey{}, nil},
		{"direct: should fail to overwrite multi-signers tx",
			txfDirect, txb2, from1, true, []cryptotypes.PubKey{}, nil},
	}
	var prevSigs []signingtypes.SignatureV2
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err = tx.Sign(tc.txf, tc.from, tc.txb, tc.overwrite)
			if len(tc.expectedPKs) == 0 {
				requireT.Error(err)
			} else {
				requireT.NoError(err)
				sigs := testSigners(requireT, tc.txb.GetTx(), tc.expectedPKs...)
				if tc.matchingSigs != nil {
					requireT.Equal(prevSigs[tc.matchingSigs[0]], sigs[tc.matchingSigs[1]])
				}
				prevSigs = sigs
			}
		})
	}
}

func testSigners(require *require.Assertions, tr signing.Tx, pks ...cryptotypes.PubKey) []signingtypes.SignatureV2 {
	sigs, err := tr.GetSignaturesV2()
	require.Len(sigs, len(pks))
	require.NoError(err)
	require.Len(sigs, len(pks))
	for i := range pks {
		require.True(sigs[i].PubKey.Equals(pks[i]), "Signature is signed with a wrong pubkey. Got: %s, expected: %s", sigs[i].PubKey, pks[i])
	}
	return sigs
}
