package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	coins = sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) (*mock.App, staking.Keeper, Keeper) {
	mapp := mock.NewApp()

	RegisterCodec(mapp.Cdc)
	staking.RegisterCodec(mapp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keySlashing := sdk.NewKVStoreKey(StoreKey)

	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	stakingKeeper := staking.NewKeeper(mapp.Cdc, keyStaking, tkeyStaking, bankKeeper, mapp.ParamsKeeper.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	keeper := NewKeeper(mapp.Cdc, keySlashing, stakingKeeper, mapp.ParamsKeeper.Subspace(DefaultParamspace), DefaultCodespace)
	mapp.Router().AddRoute(staking.RouterKey, staking.NewHandler(stakingKeeper))
	mapp.Router().AddRoute(RouterKey, NewHandler(keeper))

	mapp.SetEndBlocker(getEndBlocker(stakingKeeper))
	mapp.SetInitChainer(getInitChainer(mapp, stakingKeeper))

	require.NoError(t, mapp.CompleteSetup(keyStaking, tkeyStaking, keySlashing))

	return mapp, stakingKeeper, keeper
}

// staking endblocker
func getEndBlocker(keeper staking.Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates, tags := staking.EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
			Tags:             tags,
		}
	}
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, keeper staking.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		stakingGenesis := staking.DefaultGenesisState()
		tokens := sdk.TokensFromTendermintPower(100000)
		stakingGenesis.Pool.NotBondedTokens = tokens
		validators, err := staking.InitGenesis(ctx, keeper, stakingGenesis)
		if err != nil {
			panic(err)
		}

		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

func checkValidator(t *testing.T, mapp *mock.App, keeper staking.Keeper,
	addr sdk.AccAddress, expFound bool) staking.Validator {
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	validator, found := keeper.GetValidator(ctxCheck, sdk.ValAddress(addr1))
	require.Equal(t, expFound, found)
	return validator
}

func checkValidatorSigningInfo(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.ConsAddress, expFound bool) ValidatorSigningInfo {
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	signingInfo, found := keeper.getValidatorSigningInfo(ctxCheck, addr)
	require.Equal(t, expFound, found)
	return signingInfo
}

func TestSlashingMsgs(t *testing.T) {
	mapp, stakingKeeper, keeper := getMockApp(t)

	genTokens := sdk.TokensFromTendermintPower(42)
	bondTokens := sdk.TokensFromTendermintPower(10)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []auth.Account{acc1}
	mock.SetGenesis(mapp, accs)

	description := staking.NewDescription("foo_moniker", "", "", "")
	commission := staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())

	createValidatorMsg := staking.NewMsgCreateValidator(
		sdk.ValAddress(addr1), priv1.PubKey(), bondCoin, description, commission, sdk.OneInt(),
	)
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, []sdk.Msg{createValidatorMsg}, []uint64{0}, []uint64{0}, true, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{genCoin.Minus(bondCoin)})
	mapp.BeginBlock(abci.RequestBeginBlock{})

	validator := checkValidator(t, mapp, stakingKeeper, addr1, true)
	require.Equal(t, sdk.ValAddress(addr1), validator.OperatorAddr)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))
	unjailMsg := MsgUnjail{ValidatorAddr: sdk.ValAddress(validator.ConsPubKey.Address())}

	// no signing info yet
	checkValidatorSigningInfo(t, mapp, keeper, sdk.ConsAddress(addr1), false)

	// unjail should fail with unknown validator
	res := mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, []sdk.Msg{unjailMsg}, []uint64{0}, []uint64{1}, false, false, priv1)
	require.EqualValues(t, CodeValidatorNotJailed, res.Code)
	require.EqualValues(t, DefaultCodespace, res.Codespace)
}
