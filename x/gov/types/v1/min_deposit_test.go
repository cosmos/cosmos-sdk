package v1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetNewMinDeposit(t *testing.T) {
	type args struct {
		minDepositFloor sdk.Coins
		lastMinDeposit  sdk.Coins
		percChange      math.LegacyDec
	}
	tests := []struct {
		name string
		args args
		want sdk.Coins
	}{
		{
			"minDepositFloor = lastMinDeposit, percChange = 1, no change",
			args{
				minDepositFloor: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				lastMinDeposit:  sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				percChange:      math.LegacyOneDec(),
			},
			sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		},
		{
			"minDepositFloor > lastMinDeposit, percChange = 0.5, minDepositFloor",
			args{
				minDepositFloor: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				lastMinDeposit:  sdk.NewCoins(sdk.NewInt64Coin("stake", 500)),
				percChange:      math.LegacyNewDecWithPrec(5, 1),
			},
			sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		},
		{
			"minDepositFloor*2 = lastMinDeposit, percChange = 0.5, minDepositFloor",
			args{
				minDepositFloor: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				lastMinDeposit:  sdk.NewCoins(sdk.NewInt64Coin("stake", 2000)),
				percChange:      math.LegacyNewDecWithPrec(5, 1),
			},
			sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		},
		{
			"minDepositFloor != lastMinDeposit, percChange = 2, minDepositFloor*2",
			args{
				minDepositFloor: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				lastMinDeposit:  sdk.NewCoins(sdk.NewInt64Coin("xxx", 1000)),
				percChange:      math.LegacyNewDec(2),
			},
			sdk.NewCoins(sdk.NewInt64Coin("stake", 2000)),
		},
		{
			"minDepositFloor denoms in lastMinDeposit, percChange = 0.5, minDepositFloor",
			args{
				minDepositFloor: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				lastMinDeposit:  sdk.NewCoins(sdk.NewInt64Coin("stake", 2000), sdk.NewInt64Coin("xxx", 1000)),
				percChange:      math.LegacyNewDecWithPrec(5, 1),
			},
			sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		},
		{
			"minDepositFloor not in lastMinDeposit, percChange = 2, minDepositFloor*2",
			args{
				minDepositFloor: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000), sdk.NewInt64Coin("xxx", 1000)),
				lastMinDeposit:  sdk.NewCoins(sdk.NewInt64Coin("yyy", 2000)),
				percChange:      math.LegacyNewDec(2),
			},
			sdk.NewCoins(sdk.NewInt64Coin("stake", 2000), sdk.NewInt64Coin("xxx", 2000)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNewMinDeposit(tt.args.minDepositFloor, tt.args.lastMinDeposit, tt.args.percChange)
			require.Equal(t, tt.want, got)
		})
	}
}
