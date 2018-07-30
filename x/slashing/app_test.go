package slashing

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	coins = sdk.Coins{sdk.NewCoin("foocoin", 10)}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) (*mock.App, stake.Keeper, Keeper) {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	keyStake := sdk.NewKVStoreKey("stake")
	keySlashing := sdk.NewKVStoreKey("slashing")
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	stakeKeeper := stake.NewKeeper(mapp.Cdc, keyStake, coinKeeper, mapp.RegisterCodespace(stake.DefaultCodespace))
	keeper := NewKeeper(mapp.Cdc, keySlashing, stakeKeeper, mapp.RegisterCodespace(DefaultCodespace))
	mapp.Router().AddRoute("stake", stake.NewHandler(stakeKeeper))
	mapp.Router().AddRoute("slashing", NewHandler(keeper))

	mapp.SetEndBlocker(getEndBlocker(stakeKeeper))
	mapp.SetInitChainer(getInitChainer(mapp, stakeKeeper))
	require.NoError(t, mapp.CompleteSetup([]*sdk.KVStoreKey{keyStake, keySlashing}))

	return mapp, stakeKeeper, keeper
}

// stake endblocker
func getEndBlocker(keeper stake.Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := stake.EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	}
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, keeper stake.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		stakeGenesis := stake.DefaultGenesisState()
		stakeGenesis.Pool.LooseTokens = sdk.NewRat(100000)
		err := stake.InitGenesis(ctx, keeper, stakeGenesis)
		if err != nil {
			panic(err)
		}
		return abci.ResponseInitChain{}
	}
}

func checkValidator(t *testing.T, mapp *mock.App, keeper stake.Keeper,
	addr sdk.AccAddress, expFound bool) stake.Validator {
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	validator, found := keeper.GetValidator(ctxCheck, addr1)
	require.Equal(t, expFound, found)
	return validator
}

func checkValidatorSigningInfo(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.ValAddress, expFound bool) ValidatorSigningInfo {
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	signingInfo, found := keeper.getValidatorSigningInfo(ctxCheck, addr)
	require.Equal(t, expFound, found)
	return signingInfo
}

func TestSlashingMsgs(t *testing.T) {
	mapp, stakeKeeper, keeper := getMockApp(t)

	genCoin := sdk.NewCoin("steak", 42)
	bondCoin := sdk.NewCoin("steak", 10)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []auth.Account{acc1}
	mock.SetGenesis(mapp, accs)
	description := stake.NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := stake.NewMsgCreateValidator(
		addr1, priv1.PubKey(), bondCoin, description,
	)
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{createValidatorMsg}, []int64{0}, []int64{0}, true, priv1)
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{genCoin.Minus(bondCoin)})
	mapp.BeginBlock(abci.RequestBeginBlock{})

	validator := checkValidator(t, mapp, stakeKeeper, addr1, true)
	require.Equal(t, addr1, validator.Owner)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.RatEq(t, sdk.NewRat(10), validator.BondedTokens()))
	unrevokeMsg := MsgUnrevoke{ValidatorAddr: sdk.AccAddress(validator.PubKey.Address())}

	// no signing info yet
	checkValidatorSigningInfo(t, mapp, keeper, sdk.ValAddress(addr1), false)

	// unrevoke should fail with unknown validator
	res := mock.CheckGenTx(t, mapp.BaseApp, []sdk.Msg{unrevokeMsg}, []int64{0}, []int64{1}, false, priv1)
	require.Equal(t, sdk.ToABCICode(DefaultCodespace, CodeValidatorNotRevoked), res.Code)
}
