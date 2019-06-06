package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestTxBuilderBuild(t *testing.T) {
	type fields struct {
		TxEncoder     sdk.TxEncoder
		AccountNumber uint64
		Sequence      uint64
		Gas           uint64
		GasAdjustment float64
		SimulateGas   bool
		ChainID       string
		Memo          string
		Fees          sdk.Coins
		GasPrices     sdk.DecCoins
	}
	defaultMsg := []sdk.Msg{sdk.NewTestMsg(addr)}
	tests := []struct {
		name    string
		fields  fields
		msgs    []sdk.Msg
		want    StdSignMsg
		wantErr bool
	}{
		{
			"builder with fees",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			false,
		},
		{
			"builder with gas prices",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 2!",
				GasPrices:     sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 2!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			false,
		},
		{
			"no chain-id supplied",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			true,
		},
		{
			"builder w/ fees and gas prices",
			fields{
				TxEncoder:     DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))),
				GasPrices:     sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.NewDecWithPrec(10000, sdk.Precision))},
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bldr := NewTxBuilder(
				tt.fields.TxEncoder, tt.fields.AccountNumber, tt.fields.Sequence,
				tt.fields.Gas, tt.fields.GasAdjustment, tt.fields.SimulateGas,
				tt.fields.ChainID, tt.fields.Memo, tt.fields.Fees, tt.fields.GasPrices,
			)
			got, err := bldr.BuildSignMsg(tt.msgs)
			require.Equal(t, tt.wantErr, (err != nil))
			if err == nil {
				require.True(t, reflect.DeepEqual(tt.want, got))
			}
		})
	}
}
