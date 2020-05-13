package slashing_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
)

func checkValidator(t *testing.T, app *simapp.SimApp, _ sdk.AccAddress, expFound bool) staking.Validator {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	validator, found := app.StakingKeeper.GetValidator(ctxCheck, sdk.ValAddress(addr1))
	require.Equal(t, expFound, found)
	return validator
}

func checkValidatorSigningInfo(t *testing.T, app *simapp.SimApp, addr sdk.ConsAddress, expFound bool) slashing.ValidatorSigningInfo {
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	signingInfo, found := app.SlashingKeeper.GetValidatorSigningInfo(ctxCheck, addr)
	require.Equal(t, expFound, found)
	return signingInfo
}

func TestSlashingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42)
	bondTokens := sdk.TokensFromConsensusPower(10)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc1 := &auth.BaseAccount{
		Address: addr1,
	}
	accs := authexported.GenesisAccounts{acc1}
	balances := []bank.Balance{
		{
			Address: addr1,
			Coins:   sdk.Coins{genCoin},
		},
	}

	app := simapp.SetupWithGenesisAccounts(accs, balances...)
	simapp.CheckBalance(t, app, addr1, sdk.Coins{genCoin})

	description := staking.NewDescription("foo_moniker", "", "", "", "")
	commission := staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())

	createValidatorMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addr1), priv1.PubKey(), bondCoin, description, commission, sdk.OneInt(),
	)

	header := abci.Header{Height: app.LastBlockHeight() + 1}
	_, _, err := simapp.SignCheckDeliver(t, app.Codec(), app.BaseApp, header, []sdk.Msg{createValidatorMsg}, []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	simapp.CheckBalance(t, app, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = abci.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, app, addr1, true)
	require.Equal(t, sdk.ValAddress(addr1), validator.OperatorAddress)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := slashing.MsgUnjail{ValidatorAddr: sdk.ValAddress(validator.GetConsPubKey().Address())}

	checkValidatorSigningInfo(t, app, sdk.ConsAddress(addr1), true)

	// unjail should fail with unknown validator
	header = abci.Header{Height: app.LastBlockHeight() + 1}
	_, res, err := simapp.SignCheckDeliver(t, app.Codec(), app.BaseApp, header, []sdk.Msg{unjailMsg}, []uint64{0}, []uint64{1}, false, false, priv1)
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(slashing.ErrValidatorNotJailed, err))
}
