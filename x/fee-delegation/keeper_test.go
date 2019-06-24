package fee_delegation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/fee-delegation"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type testInput struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	ak     auth.AccountKeeper
	pk     params.Keeper
	bk     bank.Keeper
	dk     fee_delegation.Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	fee_delegation.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	delCapKey := sdk.NewKVStoreKey("delKey")
	fckCapKey := sdk.NewKVStoreKey("fckCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(delCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(fckCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)
	ak := auth.NewAccountKeeper(
		cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id", Time: time.Now().UTC()}, false, log.NewNopLogger())

	bk := bank.NewBaseKeeper(ak, pk.Subspace(banktypes.DefaultParamspace), banktypes.DefaultCodespace)
	bk.SetSendEnabled(ctx, true)

	dk := fee_delegation.NewKeeper(delCapKey, cdc)

	ak.SetParams(ctx, auth.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, pk: pk, bk: bk, dk: dk}
}

const (
	// some valid cosmos keys....
	sender    = "cosmos157ez5zlaq0scm9aycwphhqhmg3kws4qusmekll"
	recipient = "cosmos1rjxwm0rwyuldsg00qf5lt26wxzzppjzxs2efdw"
)

func TestKeeperFees(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	addr, err := sdk.AccAddressFromBech32(sender)
	require.NoError(t, err)
	addr2, err := sdk.AccAddressFromBech32(recipient)
	require.NoError(t, err)
	input.bk.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("tree", 10000)))

	require.True(t, input.bk.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("tree", 10000))))

	now := ctx.BlockHeader().Time
	require.NotNil(t, now)
	smallCoin := sdk.NewCoins(sdk.NewInt64Coin("tree", 2))
	someCoin := sdk.NewCoins(sdk.NewInt64Coin("tree", 123))
	lotCoin := sdk.NewCoins(sdk.NewInt64Coin("tree", 4567))

	// not allows
	ok := input.dk.AllowDelegatedFees(ctx, addr2, addr, smallCoin)
	require.False(t, ok)

	// allow it
	input.dk.DelegateFeeAllowance(ctx, addr2, addr, fee_delegation.BasicFeeAllowance{someCoin})

	// okay under threshold
	ok = input.dk.AllowDelegatedFees(ctx, addr2, addr, smallCoin)
	require.True(t, ok)

	// too high
	ok = input.dk.AllowDelegatedFees(ctx, addr2, addr, lotCoin)
	require.False(t, ok)

	// wrong grantee
	ok = input.dk.AllowDelegatedFees(ctx, addr2, addr2, smallCoin)
	require.False(t, ok)
}
