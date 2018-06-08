package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

var (
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	priv2 = crypto.GenPrivKeyEd25519()
	addr2 = priv2.PubKey().Address()
	addr3 = crypto.GenPrivKeyEd25519().PubKey().Address()
	priv4 = crypto.GenPrivKeyEd25519()
	addr4 = priv4.PubKey().Address()
	coins = sdk.Coins{{"foocoin", 10}}
	fee   = auth.StdFee{
		sdk.Coins{{"foocoin", 0}},
		100000,
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) (*mock.App, Keeper) {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	keyStake := sdk.NewKVStoreKey("stake")
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	keeper := NewKeeper(mapp.Cdc, keyStake, coinKeeper, mapp.RegisterCodespace(DefaultCodespace))
	mapp.Router().AddRoute("stake", NewHandler(keeper))

	mapp.SetEndBlocker(getEndBlocker(keeper))
	mapp.SetInitChainer(getInitChainer(mapp, keeper))

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{keyStake})
	return mapp, keeper
}

// stake endblocker
func getEndBlocker(keeper Keeper) sdk.EndBlocker {
	return func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := EndBlocker(ctx, keeper)
		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	}
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, keeper Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		InitGenesis(ctx, keeper, DefaultGenesisState())

		return abci.ResponseInitChain{}
	}
}

func TestStakeMsgs(t *testing.T) {
	mapp, keeper := getMockApp(t)

	genCoin := sdk.Coin{"steak", 42}
	bondCoin := sdk.Coin{"steak", 10}

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{genCoin},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{genCoin},
	}
	accs := []auth.Account{acc1, acc2}

	mock.SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	res2 := mapp.AccountMapper.GetAccount(ctxCheck, addr2)
	require.Equal(t, acc1, res1)
	require.Equal(t, acc2, res2)

	// Create Validator

	description := NewDescription("foo_moniker", "", "", "")
	createValidatorMsg := NewMsgCreateValidator(
		addr1, priv1.PubKey(), bondCoin, description,
	)
	mock.SignCheckDeliver(t, mapp.BaseApp, createValidatorMsg, []int64{0}, true, priv1)

	ctxDeliver := mapp.BaseApp.NewContext(false, abci.Header{})
	res1 = mapp.AccountMapper.GetAccount(ctxDeliver, addr1)
	require.Equal(t, sdk.Coins{genCoin.Minus(bondCoin)}, res1.GetCoins())
	validator, found := keeper.GetValidator(ctxDeliver, addr1)
	require.True(t, found)
	require.Equal(t, addr1, validator.Owner)
	require.Equal(t, sdk.Bonded, validator.Status())
	require.True(sdk.RatEq(t, sdk.NewRat(10), validator.PoolShares.Bonded()))

	// check the bond that should have been created as well
	bond, found := keeper.GetDelegation(ctxDeliver, addr1, addr1)
	require.True(sdk.RatEq(t, sdk.NewRat(10), bond.Shares))

	// Edit Validator

	description = NewDescription("bar_moniker", "", "", "")
	editValidatorMsg := NewMsgEditValidator(
		addr1, description,
	)
	mock.SignDeliver(t, mapp.BaseApp, editValidatorMsg, []int64{1}, true, priv1)

	validator, found = keeper.GetValidator(ctxDeliver, addr1)
	require.True(t, found)
	require.Equal(t, description, validator.Description)

	// Delegate

	delegateMsg := NewMsgDelegate(
		addr2, addr1, bondCoin,
	)
	mock.SignDeliver(t, mapp.BaseApp, delegateMsg, []int64{0}, true, priv2)

	res2 = mapp.AccountMapper.GetAccount(ctxDeliver, addr2)
	require.Equal(t, sdk.Coins{genCoin.Minus(bondCoin)}, res2.GetCoins())
	bond, found = keeper.GetDelegation(ctxDeliver, addr2, addr1)
	require.True(t, found)
	require.Equal(t, addr2, bond.DelegatorAddr)
	require.Equal(t, addr1, bond.ValidatorAddr)
	require.True(sdk.RatEq(t, sdk.NewRat(10), bond.Shares))

	// Unbond

	unbondMsg := NewMsgUnbond(
		addr2, addr1, "MAX",
	)
	mock.SignDeliver(t, mapp.BaseApp, unbondMsg, []int64{1}, true, priv2)

	res2 = mapp.AccountMapper.GetAccount(ctxDeliver, addr2)
	require.Equal(t, sdk.Coins{genCoin}, res2.GetCoins())
	_, found = keeper.GetDelegation(ctxDeliver, addr2, addr1)
	require.False(t, found)
}
