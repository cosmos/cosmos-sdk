package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSupply(t *testing.T) {
	initialPower := int64(100)
	initTokens := sdk.TokensFromTendermintPower(initialPower)
	nAccs := int64(4)

	ctx, _, keeper := createTestInput(t, false, initialPower, nAccs)
	balance, _, _ := keeper.AccountsSupply(ctx)
	require.Equal(t, initTokens.MulRaw(nAccs).Int64(), balance.AmountOf(keeper.sk.BondDenom(ctx)).Int64())

	total := keeper.TotalSupply(ctx)

	expectedTotal := balance.Add(keeper.EscrowedSupply(ctx))
	require.Equal(t, expectedTotal, total)
}

// create a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()

	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	distribution.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	params.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func createTestInput(t *testing.T, isCheckTx bool, initPower int64, nAccs int64) (sdk.Context, auth.AccountKeeper, Keeper) {

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyFeeCollection := sdk.NewKVStoreKey(auth.FeeStoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyDistr := sdk.NewKVStoreKey(distribution.StoreKey)
	tkeyDistr := sdk.NewTransientStoreKey(distribution.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(tkeyStaking, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyFeeCollection, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyDistr, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "supply-chain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	cdc := MakeTestCodec()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	fck := auth.NewFeeCollectionKeeper(cdc, keyFeeCollection)
	bk := bank.NewBaseKeeper(ak, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := staking.NewKeeper(cdc, keyStaking, tkeyStaking, bk, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	dk := distribution.NewKeeper(cdc, keyDistr, pk.Subspace(distribution.DefaultParamspace), bk, sk, fck, distribution.DefaultCodespace)

	valTokens := sdk.TokensFromTendermintPower(initPower)

	pool := staking.InitialPool()
	pool.NotBondedTokens = pool.NotBondedTokens.Add(valTokens.MulRaw(10))
	sk.SetPool(ctx, pool)
	sk.SetParams(ctx, staking.DefaultParams())
	dk.SetFeePool(ctx, distribution.InitialFeePool())

	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	createTestAccs(ctx, int(nAccs), initialCoins, &ak)

	keeper := NewKeeper(cdc, ak, dk, fck, sk)

	return ctx, ak, keeper
}

// nolint: unparam
func createTestAccs(ctx sdk.Context, numAccs int, initialCoins sdk.Coins, ak *auth.AccountKeeper) (accs []auth.Account) {
	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())
		acc := auth.NewBaseAccountWithAddress(addr)
		acc.Coins = initialCoins
		acc.PubKey = pubKey
		acc.AccountNumber = uint64(i)
		ak.SetAccount(ctx, &acc)
	}
	return
}
