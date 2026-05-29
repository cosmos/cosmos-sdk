package slashing_test

import (
	"errors"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())

	valKey  = ed25519.GenPrivKey()
	valAddr = sdk.AccAddress(valKey.PubKey().Address())
)

func TestSlashingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	genAccs := []authtypes.GenesisAccount{acc1}
	balances := []banktypes.Balance{{
		Address: addr1.String(),
		Coins:   sdk.Coins{genCoin},
	}}

	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)

	ta := testapp.SetupWithGenesisValSet(t, valSet, genAccs, balances...)
	stakingKeeper := ta.StakingKeeper
	bankKeeper := ta.BankKeeper
	slashingKeeper := ta.SlashingKeeper

	baseApp := ta.BaseApp

	ctxCheck := baseApp.NewContext(true)
	require.True(t, sdk.Coins{genCoin}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr1)))

	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "")
	commission := stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())

	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(addr1).String(), valKey.PubKey(), bondCoin, description, commission, math.OneInt(),
	)
	require.NoError(t, err)

	header := cmtproto.Header{Height: ta.LastBlockHeight() + 1}
	txConfig := moduletestutil.MakeTestTxConfig()
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, ta.BaseApp, header, []sdk.Msg{createValidatorMsg}, "test-chain", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr1)))

	_, err = ta.FinalizeBlock(&abci.RequestFinalizeBlock{Height: ta.LastBlockHeight() + 1})
	require.NoError(t, err)

	ctxCheck = baseApp.NewContext(true)
	validator, err := stakingKeeper.GetValidator(ctxCheck, sdk.ValAddress(addr1))
	require.NoError(t, err)

	require.Equal(t, sdk.ValAddress(addr1).String(), validator.OperatorAddress)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.True(math.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := &types.MsgUnjail{ValidatorAddr: sdk.ValAddress(addr1).String()}

	ctxCheck = ta.NewContext(true)
	_, err = slashingKeeper.GetValidatorSigningInfo(ctxCheck, sdk.ConsAddress(valAddr))
	require.NoError(t, err)

	// unjail should fail with unknown validator
	header = cmtproto.Header{Height: ta.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, ta.BaseApp, header, []sdk.Msg{unjailMsg}, "test-chain", []uint64{0}, []uint64{1}, false, false, priv1)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrValidatorNotJailed))
}
