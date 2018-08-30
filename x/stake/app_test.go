package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = ed25519.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	addr3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	priv4 = ed25519.GenPrivKey()
	addr4 = sdk.AccAddress(priv4.PubKey().Address())
	coins = sdk.Coins{{"foocoin", sdk.NewInt(10)}}
	fee   = auth.StdFee{
		sdk.Coins{{"foocoin", sdk.NewInt(0)}},
		100000,
	}
)

// getMockApp returns an initialized mock application for this module.
func getMockApp(t *testing.T) (*mock.App, Keeper) {
	mApp := mock.NewApp()

	RegisterWire(mApp.Cdc)

	keyStake := sdk.NewKVStoreKey("stake")
	coinKeeper := bank.NewKeeper(mApp.AccountMapper)
	keeper := NewKeeper(mApp.Cdc, keyStake, coinKeeper, mApp.RegisterCodespace(DefaultCodespace))

	mApp.Router().AddRoute("stake", NewHandler(keeper))
	mApp.SetEndBlocker(getEndBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper))

	require.NoError(t, mApp.CompleteSetup([]*sdk.KVStoreKey{keyStake}))
	return mApp, keeper
}

// getEndBlocker returns a stake endblocker.
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := EndBlocker(ctx, keeper)

		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	}
}

// getInitChainer initializes the chainer of the mock app and sets the genesis
// state. It returns an empty ResponseInitChain.
func getInitChainer(mapp *mock.App, keeper Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakeGenesis := DefaultGenesisState()
		stakeGenesis.Pool.LooseTokens = sdk.NewRat(100000)

		err := InitGenesis(ctx, keeper, stakeGenesis)
		if err != nil {
			panic(err)
		}

		return abci.ResponseInitChain{}
	}
}

//__________________________________________________________________________________________

func checkValidator(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.AccAddress, expFound bool) Validator {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	validator, found := keeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkDelegation(
	t *testing.T, mapp *mock.App, keeper Keeper, delegatorAddr,
	validatorAddr sdk.AccAddress, expFound bool, expShares sdk.Rat,
) {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	delegation, found := keeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		require.True(t, found)
		require.True(sdk.RatEq(t, expShares, delegation.Shares))

		return
	}

	require.False(t, found)
}

func TestStakeMsgs(t *testing.T) {
	mApp, keeper := getMockApp(t)

	genCoin := sdk.NewCoin("steak", 42)
	bondCoin := sdk.NewCoin("steak", 10)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []auth.Account{acc1, acc2}

	mock.SetGenesis(mApp, accs)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin})
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})

	// create validator
	description := NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := NewMsgCreateValidator(
		addr1, priv1.PubKey(), bondCoin, description,
	)

	mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{createValidatorMsg}, []int64{0}, []int64{0}, true, priv1)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin.Minus(bondCoin)})
	mApp.BeginBlock(abci.RequestBeginBlock{})

	validator := checkValidator(t, mApp, keeper, addr1, true)
	require.Equal(t, addr1, validator.Owner)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.RatEq(t, sdk.NewRat(10), validator.BondedTokens()))

	// addr1 create validator on behalf of addr2
	createValidatorMsgOnBehalfOf := NewMsgCreateValidatorOnBehalfOf(addr1, addr2, priv2.PubKey(), bondCoin, description)

	mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{createValidatorMsgOnBehalfOf}, []int64{0, 1}, []int64{1, 0}, true, priv1, priv2)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin.Minus(bondCoin).Minus(bondCoin)})
	mApp.BeginBlock(abci.RequestBeginBlock{})

	validator = checkValidator(t, mApp, keeper, addr2, true)
	require.Equal(t, addr2, validator.Owner)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.RatEq(t, sdk.NewRat(10), validator.Tokens))

	// check the bond that should have been created as well
	checkDelegation(t, mApp, keeper, addr1, addr1, true, sdk.NewRat(10))

	// edit the validator
	description = NewDescription("bar_moniker", "", "", "")
	editValidatorMsg := NewMsgEditValidator(addr1, description)

	mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{editValidatorMsg}, []int64{0}, []int64{2}, true, priv1)
	validator = checkValidator(t, mApp, keeper, addr1, true)
	require.Equal(t, description, validator.Description)

	// delegate
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})
	delegateMsg := NewMsgDelegate(addr2, addr1, bondCoin)

	mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{delegateMsg}, []int64{1}, []int64{1}, true, priv2)
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Minus(bondCoin)})
	checkDelegation(t, mApp, keeper, addr2, addr1, true, sdk.NewRat(10))

	// begin unbonding
	beginUnbondingMsg := NewMsgBeginUnbonding(addr2, addr1, sdk.NewRat(10))
	mock.SignCheckDeliver(t, mApp.BaseApp, []sdk.Msg{beginUnbondingMsg}, []int64{1}, []int64{2}, true, priv2)

	// delegation should exist anymore
	checkDelegation(t, mApp, keeper, addr2, addr1, false, sdk.Rat{})

	// balance should be the same because bonding not yet complete
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Minus(bondCoin)})
}
