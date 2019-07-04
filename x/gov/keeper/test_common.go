package keeper // noalias

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// dummy addresses used for testing
// nolint: unused deadcode
var (
	Addrs = createTestAddrs(500)
	PKs   = createTestPubKeys(500)

	addrDels = []sdk.AccAddress{
		Addrs[0],
		Addrs[1],
	}
	addrVals = []sdk.ValAddress{
		sdk.ValAddress(Addrs[2]),
		sdk.ValAddress(Addrs[3]),
		sdk.ValAddress(Addrs[4]),
		sdk.ValAddress(Addrs[5]),
		sdk.ValAddress(Addrs[6]),
	}

	TestAddrs = []sdk.AccAddress{
		addrDels[0], addrDels[1], addrDels[2],
		addrVals[0], addrVals[1], addrVals[2],
	}
)

// nolint: deadcode unused
func makeTestCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

// nolint: deadcode unused
func createTestInput(t *testing.T, isCheckTx bool, initPower int64) (sdk.Context, auth.AccountKeeper, Keeper, types.SupplyKeeper) {

// FIXME: update tests 

	initTokens := sdk.TokensFromConsensusPower(initPower)

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyGov:= sdk.NewKVStoreKey(types.StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	require.Nil(t, ms.LoadLatestVersion())


	ctx := sdk.NewContext(ms, abci.Header{ChainID: "gov-chain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	cdc := makeTestCodec()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	supplyKeeper := supply.NewKeeper(cdc, keySupply, accountKeeper, bankKeeper, supply.DefaultCodespace,
		[]string{auth.FeeCollectorName}, []string{}, []string{types.ModuleName})

	sk := staking.NewKeeper(cdc, keyStaking, tkeyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	sk.SetParams(ctx, staking.DefaultParams())

	rtr := types.NewRouter().
		AddRoute(types.RouterKey, types.ProposalHandler)

	keeper := keep.NewKeeper(mApp.Cdc, keyGov, pk.Subspace(types.DefaultParamspace), supplyKeeper, sk, types.DefaultCodespace, rtr)

	initCoins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens.MulRaw(int64(len(TestAddrs)))))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	for _, addr := range TestAddrs {
		_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		require.Nil(t, err)
	}

	// create module accounts
	feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName, supply.Basic)
	govAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Burner)

	keeper.supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	keeper.supplyKeeper.SetModuleAccount(ctx, govAcc)


	return ctx, accountKeeper, keeper, supplyKeeper
}

