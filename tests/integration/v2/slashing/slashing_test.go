package slashing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	"cosmossdk.io/x/slashing/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	banktestutil "cosmossdk.io/x/bank/testutil"
	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	f := initFixture(t)

	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc := f.accountKeeper.NewAccountWithAddress(f.ctx, addr1)
	f.accountKeeper.SetAccount(f.ctx, acc)
	require.NoError(t, banktestutil.FundAccount(f.ctx, f.bankKeeper, addr1, sdk.NewCoins(genCoin)))

	description := stakingtypes.NewDescription("foo_moniker", "", "", "", "", &stakingtypes.Metadata{})
	commission := stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())

	addrStrVal, err := valaddrCodec.BytesToString(addr1)
	require.NoError(t, err)
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		addrStrVal, valKey.PubKey(), bondCoin, description, commission, math.OneInt(),
	)

	stakingMsgServer := stakingkeeper.NewMsgServerImpl(f.stakingKeeper)
	_, err = f.app.RunMsg(
		t,
		f.ctx,
		func(ctx context.Context) (transaction.Msg, error) {
			res, err := stakingMsgServer.CreateValidator(ctx, createValidatorMsg)
			return res, err
		},
		integration.WithAutomaticCommit(),
	)

	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(f.bankKeeper.GetAllBalances(f.ctx, addr1)))
	_, err = f.stakingKeeper.EndBlocker(f.ctx)
	require.NoError(t, err)

	validator, err := f.stakingKeeper.GetValidator(f.ctx, sdk.ValAddress(addr1))
	require.NoError(t, err)

	require.Equal(t, addrStrVal, validator.OperatorAddress)
	require.Equal(t, stakingtypes.Bonded, validator.Status)
	require.True(math.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := &types.MsgUnjail{ValidatorAddr: addrStrVal}

	_, err = f.slashingKeeper.ValidatorSigningInfo.Get(f.ctx, sdk.ConsAddress(valAddr))
	require.NoError(t, err)

	// unjail should fail with validator not jailed
	_, err = f.app.RunMsg(
		t,
		f.ctx,
		func(ctx context.Context) (transaction.Msg, error) {
			res, err := f.slashingMsgServer.Unjail(ctx, unjailMsg)
			return res, err
		},
		integration.WithAutomaticCommit(),
	)
	require.ErrorIs(t, err, types.ErrValidatorNotJailed)
}
