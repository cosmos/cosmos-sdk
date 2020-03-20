package staking_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func checkValidator(t *testing.T, app *simapp.SimApp, addr sdk.ValAddress, expFound bool) staking.Validator {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	validator, found := app.StakingKeeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkDelegation(
	t *testing.T, app *simapp.SimApp, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress, expFound bool, expShares sdk.Dec,
) {

	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	delegation, found := app.StakingKeeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		require.True(t, found)
		require.True(sdk.DecEq(t, expShares, delegation.Shares))

		return
	}

	require.False(t, found)
}

func TestStakingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42)
	bondTokens := sdk.TokensFromConsensusPower(10)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc1 := &auth.BaseAccount{Address: addr1}
	acc2 := &auth.BaseAccount{Address: addr2}
	accs := authexported.GenesisAccounts{acc1, acc2}
	balances := []bank.Balance{
		{
			Address: addr1,
			Coins:   sdk.Coins{genCoin},
		},
		{
			Address: addr2,
			Coins:   sdk.Coins{genCoin},
		},
	}

	app := simapp.SetupWithGenesisAccounts(accs, balances...)
	simapp.CheckBalance(t, app, addr1, sdk.Coins{genCoin})
	simapp.CheckBalance(t, app, addr2, sdk.Coins{genCoin})

	// create validator
	description := staking.NewDescription("foo_moniker", "", "", "", "")
	createValidatorMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addr1), priv1.PubKey(), bondCoin, description, commissionRates, sdk.OneInt(),
	)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	_, _, err := simapp.SignCheckDeliver(t, app.Codec(), app.BaseApp, header, []sdk.Msg{createValidatorMsg}, []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	simapp.CheckBalance(t, app, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, app, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1), validator.OperatorAddress)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))

	header = abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// edit the validator
	description = staking.NewDescription("bar_moniker", "", "", "", "")
	editValidatorMsg := staking.NewMsgEditValidator(sdk.ValAddress(addr1), description, nil, nil)

	header = abci.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simapp.SignCheckDeliver(t, app.Codec(), app.BaseApp, header, []sdk.Msg{editValidatorMsg}, []uint64{0}, []uint64{1}, true, true, priv1)
	require.NoError(t, err)

	validator = checkValidator(t, app, sdk.ValAddress(addr1), true)
	require.Equal(t, description, validator.Description)

	// delegate
	simapp.CheckBalance(t, app, addr2, sdk.Coins{genCoin})
	delegateMsg := staking.NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)

	header = abci.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simapp.SignCheckDeliver(t, app.Codec(), app.BaseApp, header, []sdk.Msg{delegateMsg}, []uint64{1}, []uint64{0}, true, true, priv2)
	require.NoError(t, err)

	simapp.CheckBalance(t, app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
	checkDelegation(t, app, addr2, sdk.ValAddress(addr1), true, bondTokens.ToDec())

	// begin unbonding
	beginUnbondingMsg := staking.NewMsgUndelegate(addr2, sdk.ValAddress(addr1), bondCoin)
	header = abci.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simapp.SignCheckDeliver(t, app.Codec(), app.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, []uint64{1}, []uint64{1}, true, true, priv2)
	require.NoError(t, err)

	// delegation should exist anymore
	checkDelegation(t, app, addr2, sdk.ValAddress(addr1), false, sdk.Dec{})

	// balance should be the same because bonding not yet complete
	simapp.CheckBalance(t, app, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
}
