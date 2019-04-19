package genutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func addGenesisAccount(
	cdc *codec.Codec, appState app.GenesisState, addr sdk.AccAddress,
	coins, vestingAmt sdk.Coins, vestingStart, vestingEnd int64,
) (app.GenesisState, error) {

	for _, stateAcc := range appState.Accounts {
		if stateAcc.Address.Equals(addr) {
			return appState, fmt.Errorf("the application state already contains account %v", addr)
		}
	}

	acc := auth.NewBaseAccountWithAddress(addr)
	acc.Coins = coins

	if vestingAmt.IsZero() {
		appState.Accounts = append(appState.Accounts, app.NewGenesisAccount(&acc))
		return appState, nil
	}

	bvacc := auth.NewBaseVestingAccount(&acc, vestingAmt, sdk.NewCoins(), sdk.NewCoins(), vestingEnd)
	if bvacc.OriginalVesting.IsAllGT(acc.Coins) {
		return appState, fmt.Errorf("vesting amount cannot be greater than total amount")
	}
	if vestingStart >= vestingEnd {
		return appState, fmt.Errorf("vesting start time must before end time")
	}

	var vacc auth.VestingAccount
	if vestingStart != 0 {
		vacc = auth.NewContinuousVestingAccountRaw(bvacc, vestingStart)
	} else {
		vacc = auth.NewDelayedVestingAccountRaw(bvacc)
	}

	appState.Accounts = append(appState.Accounts, app.NewGenesisAccountI(vacc))
	return appState, nil
}
