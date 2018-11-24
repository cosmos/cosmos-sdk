package context

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestTxBuilderBuild(t *testing.T) {
	type fields struct {
		Codec         *codec.Codec
		AccountNumber uint64
		Sequence      uint64
		Gas           uint64
		GasAdjustment float64
		SimulateGas   bool
		ChainID       string
		Memo          string
		Fee           string
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
				Codec:         codec.New(),
				AccountNumber: 1,
				Sequence:      1,
				Gas:           100,
				GasAdjustment: 1.1,
				SimulateGas:   false,
				ChainID:       "test-chain",
				Memo:          "hello",
				Fee:           "1" + stakeTypes.DefaultBondDenom,
			},
			defaultMsg,
			StdSignMsg{
				ChainID:       "test-chain",
				AccountNumber: 1,
				Sequence:      1,
				Memo:          "hello",
				Msgs:          defaultMsg,
				Fee:           auth.NewStdFee(100, sdk.NewCoin(stakeTypes.DefaultBondDenom, sdk.NewInt(1))),
			},
			false,
		},
	}
	for i, tc := range tests {
		bldr := TxBuilder{
			Codec:         tc.fields.Codec,
			AccountNumber: tc.fields.AccountNumber,
			Sequence:      tc.fields.Sequence,
			Gas:           tc.fields.Gas,
			GasAdjustment: tc.fields.GasAdjustment,
			SimulateGas:   tc.fields.SimulateGas,
			ChainID:       tc.fields.ChainID,
			Memo:          tc.fields.Memo,
			Fee:           tc.fields.Fee,
		}
		got, err := bldr.Build(tc.msgs)
		require.Equal(t, tc.wantErr, (err != nil), "TxBuilder.Build() error = %v, wantErr %v, tc %d", err, tc.wantErr, i)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TxBuilder.Build() = %v, want %v", got, tc.want)
		}
	}
}
