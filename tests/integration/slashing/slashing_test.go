package slashing_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	priv1        = secp256k1.GenPrivKey()
	addr1        = sdk.AccAddress(priv1.PubKey().Address())
	addrCodec    = codecaddress.NewBech32Codec("cosmos")
	valaddrCodec = codecaddress.NewBech32Codec("cosmosvaloper")

	valKey  = ed25519.GenPrivKey()
	valAddr = sdk.AccAddress(valKey.PubKey().Address())
)

func TestSlashingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	addrStr, err := addrCodec.BytesToString(addr1)
	require.NoError(t, err)
	acc1 := &authtypes.BaseAccount{
		Address: addrStr,
	}
	accs := []sims.GenesisAccount{{GenesisAccount: acc1, Coins: sdk.Coins{genCoin}}}

	startupCfg := sims.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = accs

	var (
		stakingKeeper  *stakingkeeper.Keeper
		bankKeeper     bankkeeper.Keeper
		slashingKeeper keeper.Keeper
		txConfig       client.TxConfig
	)

	app, err := sims.SetupWithConfiguration(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AccountsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.SlashingModule(),
				configurator.TxModule(),
				configurator.ValidateModule(),
				configurator.ConsensusModule(),
				configurator.BankModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		startupCfg, &stakingKeeper, &bankKeeper, &slashingKeeper, &txConfig)
	require.NoError(t, err)

	baseApp := app.BaseApp

	ctxCheck := baseApp.NewContext(true)
	require.True(t, sdk.Coins{genCoin}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr1)))

	require.NoError(t, err)

	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())

	addrStrVal, err := valaddrCodec.BytesToString(addr1)
	require.NoError(t, err)
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		addrStrVal, valKey.PubKey(), bondCoin, description, commission, math.OneInt(),
	)
	require.NoError(t, err)

	headerInfo := header.Info{Height: app.LastBlockHeight() + 1}
	_, _, err = sims.SignCheckDeliver(t, txConfig, app.BaseApp, headerInfo, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr1)))

	ctxCheck = baseApp.NewContext(true)
	validator, err := stakingKeeper.GetValidator(ctxCheck, sdk.ValAddress(addr1))
	require.NoError(t, err)

	require.Equal(t, addrStrVal, validator.OperatorAddress)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.True(math.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := &types.MsgUnjail{ValidatorAddr: addrStrVal}

	ctxCheck = app.BaseApp.NewContext(true)
	_, err = slashingKeeper.ValidatorSigningInfo.Get(ctxCheck, sdk.ConsAddress(valAddr))
	require.NoError(t, err)

	// unjail should fail with unknown validator
	headerInfo = header.Info{Height: app.LastBlockHeight() + 1}
	_, _, err = sims.SignCheckDeliver(t, txConfig, app.BaseApp, headerInfo, []sdk.Msg{unjailMsg}, "", []uint64{0}, []uint64{1}, false, false, priv1)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrValidatorNotJailed))
}
