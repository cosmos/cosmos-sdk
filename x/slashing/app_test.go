package slashing_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())

	valKey  = ed25519.GenPrivKey()
	valAddr = sdk.AccAddress(valKey.PubKey().Address())
)

func checkValidator(t *testing.T, ctxCheck sdk.Context, stakingKeeper *stakingkeeper.Keeper, addr sdk.ValAddress, expFound bool) stakingtypes.Validator {
	validator, found := stakingKeeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkValidatorSigningInfo(t *testing.T, ctxCheck sdk.Context, slashingKeeper keeper.Keeper, addr sdk.ConsAddress, expFound bool) types.ValidatorSigningInfo {
	signingInfo, found := slashingKeeper.GetValidatorSigningInfo(ctxCheck, addr)
	require.Equal(t, expFound, found)
	return signingInfo
}

// checkBalance checks the balance of an account.
func checkBalance(t *testing.T, ctxCheck sdk.Context, bankKeeper bankkeeper.Keeper, addr sdk.AccAddress, balances sdk.Coins) {
	require.True(t, balances.IsEqual(bankKeeper.GetAllBalances(ctxCheck, addr)))
}

func TestSlashingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	var (
		txConfig       client.TxConfig
		bankKeeper     bankkeeper.Keeper
		stakingKeeper  *stakingkeeper.Keeper
		slashingKeeper keeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&txConfig,
		&bankKeeper,
		&stakingKeeper,
		&slashingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	banktestutil.FundAccount(bankKeeper, ctx, addr1, sdk.Coins{genCoin})
	checkBalance(t, ctx, bankKeeper, addr1, sdk.Coins{genCoin})

	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())

	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr1), valKey.PubKey(), bondCoin, description, commission, sdk.OneInt(),
	)
	require.NoError(t, err)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, app.BaseApp, header, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	checkBalance(t, ctx, bankKeeper, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, ctx, stakingKeeper, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1).String(), validator.OperatorAddress)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := &types.MsgUnjail{ValidatorAddr: sdk.ValAddress(addr1).String()}

	checkValidatorSigningInfo(t, ctx, slashingKeeper, sdk.ConsAddress(valAddr), true)

	// unjail should fail with unknown validator
	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, res, err := simtestutil.SignCheckDeliver(t, txConfig, app.BaseApp, header, []sdk.Msg{unjailMsg}, "", []uint64{0}, []uint64{1}, false, false, priv1)
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, errors.Is(types.ErrValidatorNotJailed, err))
}
