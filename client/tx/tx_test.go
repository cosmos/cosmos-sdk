package tx_test

import (
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/grpc/simulate"
	"github.com/cosmos/cosmos-sdk/testutil"
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func NewTestTxConfig() client.TxConfig {
	_, cdc := simapp.MakeCodecs()
	return types.StdTxConfig{Cdc: cdc}
}

func TestCalculateGas(t *testing.T) {
	makeQueryFunc := func(gasUsed uint64, wantErr bool) func(string, []byte) ([]byte, int64, error) {
		return func(string, []byte) ([]byte, int64, error) {
			if wantErr {
				return nil, 0, errors.New("query failed")
			}
			simRes := &simulate.SimulateResponse{
				GasInfo: &sdk.GasInfo{GasUsed: gasUsed, GasWanted: gasUsed},
				Result:  &sdk.Result{Data: []byte("tx data"), Log: "log"},
			}

			bz, err := codec.ProtoMarshalJSON(simRes)
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
		txf := tx.Factory{}.WithChainID("test-chain").WithTxConfig(NewTestTxConfig()).WithSignMode(signing2.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

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
	txf := tx.Factory{}.
		WithTxConfig(NewTestTxConfig()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")

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
	require.Empty(t, tx.GetTx().(signing.SigVerifiableTx).GetSignatures())
}

func TestSign(t *testing.T) {
	dir, clean := testutil.NewTestCaseDir(t)
	t.Cleanup(clean)

	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(t.Name(), "test", dir, nil)
	require.NoError(t, err)

	var from = "test_sign"

	_, seed, err := kr.NewMnemonic(from, keyring.English, path, hd.Secp256k1)
	require.NoError(t, err)
	require.NoError(t, kr.Delete(from))

	info, err := kr.NewAccount(from, seed, "", path, hd.Secp256k1)
	require.NoError(t, err)

	txf := tx.Factory{}.
		WithTxConfig(NewTestTxConfig()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")

	msg := banktypes.NewMsgSend(info.GetAddress(), sdk.AccAddress("to"), nil)
	txn, err := tx.BuildUnsignedTx(txf, msg)
	require.NoError(t, err)

	t.Log("should failed if txf without keyring")
	err = tx.Sign(txf, from, txn)
	require.Error(t, err)

	txf = tx.Factory{}.
		WithKeybase(kr).
		WithTxConfig(NewTestTxConfig()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")

	t.Log("should succeed if txf with keyring")
	err = tx.Sign(txf, from, txn)
	require.NoError(t, err)

	t.Log("should fail for non existing key")
	err = tx.Sign(txf, "non_existing_key", txn)
	require.Error(t, err)
}
