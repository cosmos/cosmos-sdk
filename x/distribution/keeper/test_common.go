package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delPk2   = ed25519.GenPrivKey().PubKey()
	delPk3   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())

	valPk1      = ed25519.GenPrivKey().PubKey()
	valPk2      = ed25519.GenPrivKey().PubKey()
	valPk3      = ed25519.GenPrivKey().PubKey()
	valAddr1    = sdk.ValAddress(valPk1.Address())
	valAddr2    = sdk.ValAddress(valPk2.Address())
	valAddr3    = sdk.ValAddress(valPk3.Address())
	valAccAddr1 = sdk.AccAddress(valPk1.Address()) // generate acc addresses for these validator keys too
	valAccAddr2 = sdk.AccAddress(valPk2.Address())
	valAccAddr3 = sdk.AccAddress(valPk3.Address())

	addrs = []sdk.AccAddress{
		delAddr1, delAddr2, delAddr3,
		valAccAddr1, valAccAddr2, valAccAddr3,
	}

	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
	emptyPubkey  crypto.PubKey
)

// create a codec used only for testing
func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	bank.RegisterCodec(cdc)
	stake.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	types.RegisterCodec(cdc) // distr
	return cdc
}

// hogpodge of all sorts of input required for testing
func CreateTestInput(t *testing.T, isCheckTx bool, initCoins int64) (
	sdk.Context, auth.AccountMapper, Keeper, stake.Keeper) {

	keyDistr := sdk.NewKVStoreKey("distr")
	tkeyDistr := sdk.NewTransientStoreKey("transient_distr")
	keyStake := sdk.NewKVStoreKey("stake")
	tkeyStake := sdk.NewTransientStoreKey("transient_stake")
	keyAcc := sdk.NewKVStoreKey("acc")
	keyFeeCollection := sdk.NewKVStoreKey("fee")
	keyParams := sdk.NewKVStoreKey("params")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(tkeyDistr, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyDistr, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyStake, sdk.StoreTypeTransient, nil)
	ms.MountStoreWithDB(keyStake, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyFeeCollection, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foochainid"}, isCheckTx, log.NewNopLogger())
	cdc := MakeTestCodec()
	accountMapper := auth.NewAccountMapper(cdc, keyAcc, auth.ProtoBaseAccount)
	ck := bank.NewBaseKeeper(accountMapper)
	sk := stake.NewKeeper(cdc, keyStake, tkeyStake, ck, stake.DefaultCodespace)
	sk.SetPool(ctx, stake.InitialPool())
	sk.SetParams(ctx, stake.DefaultParams())
	sk.InitIntraTxCounter(ctx)

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range addrs {
		pool := sk.GetPool(ctx)
		_, _, err := ck.AddCoins(ctx, addr, sdk.Coins{
			{sk.GetParams(ctx).BondDenom, sdk.NewInt(initCoins)},
		})
		require.Nil(t, err)
		pool.LooseTokens = pool.LooseTokens.Add(sdk.NewDec(initCoins))
		sk.SetPool(ctx, pool)
	}

	fck := auth.NewFeeCollectionKeeper(cdc, keyFeeCollection)
	pk := params.NewKeeper(cdc, keyParams)
	keeper := NewKeeper(cdc, keyDistr, tkeyDistr, pk.Setter(), ck, sk, fck, types.DefaultCodespace)

	// set the distribution hooks on staking
	sk = sk.WithHooks(keeper.Hooks())

	// set genesis items required for distribution
	keeper.SetFeePool(ctx, types.InitialFeePool())
	keeper.SetCommunityTax(ctx, sdk.NewDecWithPrec(2, 2))

	return ctx, accountMapper, keeper, sk
}
