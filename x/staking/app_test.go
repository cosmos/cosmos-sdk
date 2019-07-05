package staking

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// getMockApp returns an initialized mock application for this module.
func getMockApp(t *testing.T) (*mock.App, Keeper) {
	mApp := mock.NewApp()

	RegisterCodec(mApp.Cdc)
	supply.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(TStoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)

	bankKeeper := bank.NewBaseKeeper(mApp.AccountKeeper, mApp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	supplyKeeper := supply.NewKeeper(mApp.Cdc, keySupply, mApp.AccountKeeper, bankKeeper, supply.DefaultCodespace,
		[]string{auth.FeeCollectorName}, []string{}, []string{types.NotBondedPoolName, types.BondedPoolName})
	keeper := NewKeeper(mApp.Cdc, keyStaking, tkeyStaking, supplyKeeper, mApp.ParamsKeeper.Subspace(DefaultParamspace), DefaultCodespace)

	mApp.Router().AddRoute(RouterKey, NewHandler(keeper))
	mApp.SetEndBlocker(getEndBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper, mApp.AccountKeeper, supplyKeeper))

	require.NoError(t, mApp.CompleteSetup(keyStaking, tkeyStaking, keySupply))
	return mApp, keeper
}

// getEndBlocker returns a staking endblocker.
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
func getInitChainer(mapp *mock.App, keeper Keeper, accountKeeper types.AccountKeeper, supplyKeeper types.SupplyKeeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		// set module accounts
		feeCollector := supply.NewEmptyModuleAccount(auth.FeeCollectorName, supply.Basic)
		notBondedPool := supply.NewEmptyModuleAccount(types.NotBondedPoolName, supply.Burner)
		bondPool := supply.NewEmptyModuleAccount(types.BondedPoolName, supply.Burner)

		supplyKeeper.SetModuleAccount(ctx, feeCollector)
		supplyKeeper.SetModuleAccount(ctx, bondPool)
		supplyKeeper.SetModuleAccount(ctx, notBondedPool)

		stakingGenesis := DefaultGenesisState()
		validators := InitGenesis(ctx, keeper, accountKeeper, supplyKeeper, stakingGenesis)
		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}

//__________________________________________________________________________________________

func checkValidator(t *testing.T, mapp *mock.App, keeper Keeper,
	addr sdk.ValAddress, expFound bool) Validator {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	validator, found := keeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkDelegation(
	t *testing.T, mapp *mock.App, keeper Keeper, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress, expFound bool, expShares sdk.Dec,
) {

	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	delegation, found := keeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		require.True(t, found)
		require.True(sdk.DecEq(t, expShares, delegation.Shares))

		return
	}

	require.False(t, found)
}

func TestStakingMsgs(t *testing.T) {
	mApp, keeper := getMockApp(t)

	genTokens := sdk.TokensFromConsensusPower(42)
	bondTokens := sdk.TokensFromConsensusPower(10)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

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
		sdk.ValAddress(addr1), priv1.PubKey(), bondCoin, description, commissionRates, sdk.OneInt(),
	)

	header := abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{createValidatorMsg}, []uint64{0}, []uint64{0}, true, true, priv1)
	mock.CheckBalance(t, mApp, addr1, sdk.Coins{genCoin.Sub(bondCoin)})

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, mApp, keeper, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1), validator.OperatorAddress)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	// edit the validator
	description = NewDescription("bar_moniker", "", "", "")
	editValidatorMsg := NewMsgEditValidator(sdk.ValAddress(addr1), description, nil, nil)

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{editValidatorMsg}, []uint64{0}, []uint64{1}, true, true, priv1)

	validator = checkValidator(t, mApp, keeper, sdk.ValAddress(addr1), true)
	require.Equal(t, description, validator.Description)

	// delegate
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin})
	delegateMsg := NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)

	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{delegateMsg}, []uint64{1}, []uint64{0}, true, true, priv2)
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
	checkDelegation(t, mApp, keeper, addr2, sdk.ValAddress(addr1), true, bondTokens.ToDec())

	// begin unbonding
	beginUnbondingMsg := NewMsgUndelegate(addr2, sdk.ValAddress(addr1), bondCoin)
	header = abci.Header{Height: mApp.LastBlockHeight() + 1}
	mock.SignCheckDeliver(t, mApp.Cdc, mApp.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, []uint64{1}, []uint64{1}, true, true, priv2)

	// delegation should exist anymore
	checkDelegation(t, mApp, keeper, addr2, sdk.ValAddress(addr1), false, sdk.Dec{})

	// balance should be the same because bonding not yet complete
	mock.CheckBalance(t, mApp, addr2, sdk.Coins{genCoin.Sub(bondCoin)})
}
