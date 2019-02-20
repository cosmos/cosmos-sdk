package context

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
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
		fields  fields
		msgs    []sdk.Msg
		want    StdSignMsg
		wantErr bool
	}{
		{
			fields{
				TxEncoder:     auth.DefaultTxEncoder(codec.New()),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           200000,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello from Voyager 1!",
				Fees:          sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))},
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello from Voyager 1!",
				Msgs:          defaultMsg,
				Fee:           auth.NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			false,
		},
		{
			fields{
				TxEncoder:     auth.DefaultTxEncoder(codec.New()),
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
				Fee:           auth.NewStdFee(200000, sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1))}),
			},
			false,
		},
	}

	for i, tc := range tests {
		bldr := NewTxBuilder(
			tc.fields.TxEncoder, tc.fields.AccountNumber, tc.fields.Sequence,
			tc.fields.Gas, tc.fields.GasAdjustment, tc.fields.SimulateGas,
			tc.fields.ChainID, tc.fields.Memo, tc.fields.Fees, tc.fields.GasPrices,
		)

		got, err := bldr.BuildSignMsg(tc.msgs)
		require.Equal(t, tc.wantErr, (err != nil), "TxBuilder.Build() error = %v, wantErr %v, tc %d", err, tc.wantErr, i)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TxBuilder.Build() = %v, want %v", got, tc.want)
		}
	}
}
