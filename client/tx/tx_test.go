package tx_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

func NewTestTxGenerator() context.TxGenerator {
	_, cdc := simapp.MakeCodecs()
	return types.StdTxGenerator{Cdc: cdc}
}

func TestCalculateGas(t *testing.T) {
	makeQueryFunc := func(gasUsed uint64, wantErr bool) func(string, []byte) ([]byte, int64, error) {
		return func(string, []byte) ([]byte, int64, error) {
			if wantErr {
				return nil, 0, errors.New("query failed")
			}
			simRes := &sdk.SimulationResponse{
				GasInfo: sdk.GasInfo{GasUsed: gasUsed, GasWanted: gasUsed},
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
		txf := tx.Factory{}.WithChainID("test-chain").WithTxGenerator(NewTestTxGenerator())

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
		WithTxGenerator(NewTestTxGenerator()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")

	msg := bank.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	bz, err := tx.BuildSimTx(txf, msg)
	require.NoError(t, err)
	require.NotNil(t, bz)
}

func TestBuildUnsignedTx(t *testing.T) {
	txf := tx.Factory{}.
		WithTxGenerator(NewTestTxGenerator()).
		WithAccountNumber(50).
		WithSequence(23).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID("test-chain")

	msg := bank.NewMsgSend(sdk.AccAddress("from"), sdk.AccAddress("to"), nil)
	tx, err := tx.BuildUnsignedTx(txf, msg)
	require.NoError(t, err)
	require.NotNil(t, tx)
	require.Equal(t, []sdk.Signature{}, tx.GetSignatures())
}
