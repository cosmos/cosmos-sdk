package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

type rekeyFixture struct {
	ctx    sdk.Context
	ak     keeper.AccountKeeper
	encCfg moduletestutil.TestEncodingConfig
}

func newRekeyFixture(t *testing.T) rekeyFixture {
	t.Helper()
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	vestingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ak := keeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		types.ProtoBaseAccount,
		getMaccPerms(),
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		types.NewModuleAddress("gov").String(),
	)
	ctx := testCtx.Ctx.WithBlockHeight(1).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	require.NoError(t, ak.Params.Set(ctx, types.DefaultParams()))
	return rekeyFixture{ctx: ctx, ak: ak, encCfg: encCfg}
}

func newAccountWithKey(t *testing.T, ctx sdk.Context, ak keeper.AccountKeeper, pk cryptotypes.PubKey) sdk.AccountI {
	t.Helper()
	addr := sdk.AccAddress(pk.Address())
	acc := ak.NewAccountWithAddress(ctx, addr)
	require.NoError(t, acc.SetPubKey(pk))
	ak.SetAccount(ctx, acc)
	return ak.GetAccount(ctx, addr)
}

func historyFor(t *testing.T, ctx sdk.Context, ak keeper.AccountKeeper, addr sdk.AccAddress) []types.PubKeyHistoryEntry {
	t.Helper()
	var out []types.PubKeyHistoryEntry
	err := ak.PubKeyHistory.Walk(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, uint64](addr),
		func(_ collections.Pair[sdk.AccAddress, uint64], v types.PubKeyHistoryEntry) (bool, error) {
			out = append(out, v)
			return false, nil
		})
	require.NoError(t, err)
	return out
}

func indexEntries(t *testing.T, ctx sdk.Context, ak keeper.AccountKeeper) []collections.Pair[sdk.AccAddress, sdk.AccAddress] {
	t.Helper()
	var out []collections.Pair[sdk.AccAddress, sdk.AccAddress]
	err := ak.RekeyIndex.Walk(ctx, nil, func(k collections.Pair[sdk.AccAddress, sdk.AccAddress]) (bool, error) {
		out = append(out, k)
		return false, nil
	})
	require.NoError(t, err)
	return out
}

func TestApplyRekey_HistoryAndIndex(t *testing.T) {
	f := newRekeyFixture(t)
	k0 := secp256k1.GenPrivKey().PubKey()
	k1 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addrA := acc.GetAddress()

	// k0 -> k1
	require.NoError(t, f.ak.ApplyRekey(f.ctx, acc, k1))

	stored := f.ak.GetAccount(f.ctx, addrA)
	require.True(t, stored.GetPubKey().Equals(k1))

	hist := historyFor(t, f.ctx, f.ak, addrA)
	require.Len(t, hist, 1)
	oldPk, ok := hist[0].PubKey.GetCachedValue().(cryptotypes.PubKey)
	require.True(t, ok)
	require.True(t, oldPk.Equals(k0))
	require.Equal(t, f.ctx.BlockHeight(), hist[0].ReplacedAtHeight)
	require.True(t, f.ctx.BlockTime().Equal(hist[0].ReplacedAtTime))
	require.Equal(t, sdk.AccAddress(k1.Address()).String(), hist[0].NewKeyAddress)

	has, err := f.ak.RekeyIndex.Has(f.ctx, collections.Join(sdk.AccAddress(k1.Address()), addrA))
	require.NoError(t, err)
	require.True(t, has)
	require.Len(t, indexEntries(t, f.ctx, f.ak), 1)

	// k1 -> k0 in a later block: the key's natural address is the account again.
	ctx2 := f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 1)
	require.NoError(t, f.ak.ApplyRekey(ctx2, stored, k0))

	hist = historyFor(t, ctx2, f.ak, addrA)
	require.Len(t, hist, 2)
	oldPk, ok = hist[1].PubKey.GetCachedValue().(cryptotypes.PubKey)
	require.True(t, ok)
	require.True(t, oldPk.Equals(k1))
	require.Equal(t, addrA.String(), hist[1].NewKeyAddress)
	require.Empty(t, indexEntries(t, ctx2, f.ak))
	require.True(t, f.ak.GetAccount(ctx2, addrA).GetPubKey().Equals(k0))
}

func historyKeys(t *testing.T, ctx sdk.Context, ak keeper.AccountKeeper, addr sdk.AccAddress) []uint64 {
	t.Helper()
	var out []uint64
	err := ak.PubKeyHistory.Walk(ctx, collections.NewPrefixedPairRange[sdk.AccAddress, uint64](addr),
		func(k collections.Pair[sdk.AccAddress, uint64], _ types.PubKeyHistoryEntry) (bool, error) {
			out = append(out, k.K2())
			return false, nil
		})
	require.NoError(t, err)
	return out
}

func TestApplyRekey_TwoRotationsSameBlock(t *testing.T) {
	f := newRekeyFixture(t)
	k0 := secp256k1.GenPrivKey().PubKey()
	k1 := secp256k1.GenPrivKey().PubKey()
	k2 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addr := acc.GetAddress()

	require.NoError(t, f.ak.ApplyRekey(f.ctx, acc, k1))
	require.NoError(t, f.ak.ApplyRekey(f.ctx, f.ak.GetAccount(f.ctx, addr), k2))

	require.True(t, f.ak.GetAccount(f.ctx, addr).GetPubKey().Equals(k2))
	require.Equal(t, []uint64{0, 1}, historyKeys(t, f.ctx, f.ak, addr))
	hist := historyFor(t, f.ctx, f.ak, addr)
	require.Len(t, hist, 2)
	require.True(t, hist[0].PubKey.GetCachedValue().(cryptotypes.PubKey).Equals(k0))
	require.True(t, hist[1].PubKey.GetCachedValue().(cryptotypes.PubKey).Equals(k1))
	require.Equal(t, f.ctx.BlockHeight(), hist[0].ReplacedAtHeight)
	require.Equal(t, f.ctx.BlockHeight(), hist[1].ReplacedAtHeight)
	require.Equal(t, []collections.Pair[sdk.AccAddress, sdk.AccAddress]{
		collections.Join(sdk.AccAddress(k2.Address()), addr),
	}, indexEntries(t, f.ctx, f.ak))
}

func TestApplyRekey_NoCurrentPubKey(t *testing.T) {
	f := newRekeyFixture(t)
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := f.ak.NewAccountWithAddress(f.ctx, addr)
	f.ak.SetAccount(f.ctx, acc)
	k1 := secp256k1.GenPrivKey().PubKey()

	require.NoError(t, f.ak.ApplyRekey(f.ctx, f.ak.GetAccount(f.ctx, addr), k1))

	hist := historyFor(t, f.ctx, f.ak, addr)
	require.Len(t, hist, 1)
	require.Nil(t, hist[0].PubKey)
	require.Equal(t, []collections.Pair[sdk.AccAddress, sdk.AccAddress]{
		collections.Join(sdk.AccAddress(k1.Address()), addr),
	}, indexEntries(t, f.ctx, f.ak))
}

func TestApplyRekey_PreservesAccountType(t *testing.T) {
	f := newRekeyFixture(t)
	k0 := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(k0.Address())
	base := types.NewBaseAccount(addr, k0, 42, 7)
	original := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
	cva, err := vestingtypes.NewContinuousVestingAccount(base, original, 100, 200)
	require.NoError(t, err)
	f.ak.SetAccount(f.ctx, cva)

	_, k1 := genMlDsa65Key(t)
	require.NoError(t, f.ak.ApplyRekey(f.ctx, f.ak.GetAccount(f.ctx, addr), k1))

	got, ok := f.ak.GetAccount(f.ctx, addr).(*vestingtypes.ContinuousVestingAccount)
	require.True(t, ok, "account type changed")
	require.True(t, got.GetPubKey().Equals(k1))
	require.Equal(t, uint64(42), got.AccountNumber)
	require.Equal(t, uint64(7), got.Sequence)
	require.Equal(t, int64(100), got.StartTime)
	require.Equal(t, int64(200), got.EndTime)
	require.Equal(t, original, got.OriginalVesting)
}

func genMlDsa65Key(t *testing.T) (mldsa65.PrivKey, cryptotypes.PubKey) {
	t.Helper()
	sk, err := mldsa65.GenPrivKey()
	require.NoError(t, err)
	return sk, sk.PubKey()
}

type historyKV struct {
	key collections.Pair[sdk.AccAddress, uint64]
	bz  []byte
}

func allHistory(t *testing.T, f rekeyFixture, ctx sdk.Context, ak keeper.AccountKeeper) []historyKV {
	t.Helper()
	var out []historyKV
	err := ak.PubKeyHistory.Walk(ctx, nil, func(k collections.Pair[sdk.AccAddress, uint64], v types.PubKeyHistoryEntry) (bool, error) {
		out = append(out, historyKV{key: k, bz: f.encCfg.Codec.MustMarshal(&v)})
		return false, nil
	})
	require.NoError(t, err)
	return out
}

func TestGenesis_RekeyRoundTrip(t *testing.T) {
	f := newRekeyFixture(t)
	f.ak.InitGenesis(f.ctx, *types.DefaultGenesisState())

	// Account A: secp256k1 -> secp256k1 -> ML-DSA-65 across two blocks.
	ka0 := secp256k1.GenPrivKey().PubKey()
	ka1 := secp256k1.GenPrivKey().PubKey()
	_, ka2 := genMlDsa65Key(t)
	accA := newAccountWithKey(t, f.ctx, f.ak, ka0)
	require.NoError(t, f.ak.ApplyRekey(f.ctx, accA, ka1))
	ctx2 := f.ctx.WithBlockHeight(f.ctx.BlockHeight() + 5).WithBlockTime(f.ctx.BlockTime().Add(time.Minute))
	require.NoError(t, f.ak.ApplyRekey(ctx2, f.ak.GetAccount(ctx2, accA.GetAddress()), ka2))

	// Account B: rotate once, then rotate back to its natural key.
	kb0 := secp256k1.GenPrivKey().PubKey()
	kb1 := secp256k1.GenPrivKey().PubKey()
	accB := newAccountWithKey(t, ctx2, f.ak, kb0)
	require.NoError(t, f.ak.ApplyRekey(ctx2, accB, kb1))
	ctx3 := ctx2.WithBlockHeight(ctx2.BlockHeight() + 1)
	require.NoError(t, f.ak.ApplyRekey(ctx3, f.ak.GetAccount(ctx3, accB.GetAddress()), kb0))

	// Account C: rotated, stays rotated.
	kc0 := secp256k1.GenPrivKey().PubKey()
	kc1 := secp256k1.GenPrivKey().PubKey()
	accC := newAccountWithKey(t, ctx3, f.ak, kc0)
	require.NoError(t, f.ak.ApplyRekey(ctx3, accC, kc1))

	// Account D: rotates twice in the same block.
	kd0 := secp256k1.GenPrivKey().PubKey()
	kd1 := secp256k1.GenPrivKey().PubKey()
	kd2 := secp256k1.GenPrivKey().PubKey()
	accD := newAccountWithKey(t, ctx3, f.ak, kd0)
	require.NoError(t, f.ak.ApplyRekey(ctx3, accD, kd1))
	require.NoError(t, f.ak.ApplyRekey(ctx3, f.ak.GetAccount(ctx3, accD.GetAddress()), kd2))

	wantHistory := allHistory(t, f, ctx3, f.ak)
	wantIndex := indexEntries(t, ctx3, f.ak)
	require.Len(t, wantHistory, 7)
	require.Len(t, wantIndex, 3)

	exported := f.ak.ExportGenesis(ctx3)
	require.NoError(t, types.ValidateGenesis(*exported))
	require.Len(t, exported.PubKeyHistory, 4)

	// Go through JSON, as a real export/import does.
	bz, err := f.encCfg.Codec.MarshalJSON(exported)
	require.NoError(t, err)
	var imported types.GenesisState
	require.NoError(t, f.encCfg.Codec.UnmarshalJSON(bz, &imported))
	require.NoError(t, types.ValidateGenesis(imported))

	g := newRekeyFixture(t)
	g.ak.InitGenesis(g.ctx, imported)

	require.Equal(t, wantHistory, allHistory(t, g, g.ctx, g.ak))
	require.Equal(t, wantIndex, indexEntries(t, g.ctx, g.ak))
	require.True(t, g.ak.GetAccount(g.ctx, accA.GetAddress()).GetPubKey().Equals(ka2))
	require.True(t, g.ak.GetAccount(g.ctx, accB.GetAddress()).GetPubKey().Equals(kb0))
	require.True(t, g.ak.GetAccount(g.ctx, accC.GetAddress()).GetPubKey().Equals(kc1))
	require.True(t, g.ak.GetAccount(g.ctx, accD.GetAddress()).GetPubKey().Equals(kd2))

	// A second export from the imported state is identical.
	reexported := g.ak.ExportGenesis(g.ctx)
	bz2, err := g.encCfg.Codec.MarshalJSON(reexported)
	require.NoError(t, err)
	require.JSONEq(t, string(bz), string(bz2))
}

// TestGenesis_RekeyZeroHeightRestart imports an export into a chain that
// restarts at height 0, as a zero-height export does, and rotates again at a
// height that an imported entry already has.
func TestGenesis_RekeyZeroHeightRestart(t *testing.T) {
	f := newRekeyFixture(t)
	f.ak.InitGenesis(f.ctx, *types.DefaultGenesisState())

	k0 := secp256k1.GenPrivKey().PubKey()
	k1 := secp256k1.GenPrivKey().PubKey()
	k2 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addr := acc.GetAddress()
	oldCtx := f.ctx.WithBlockHeight(5)
	require.NoError(t, f.ak.ApplyRekey(oldCtx, acc, k1))

	exported := f.ak.ExportGenesis(oldCtx)
	bz, err := f.encCfg.Codec.MarshalJSON(exported)
	require.NoError(t, err)
	var imported types.GenesisState
	require.NoError(t, f.encCfg.Codec.UnmarshalJSON(bz, &imported))
	require.NoError(t, types.ValidateGenesis(imported))

	g := newRekeyFixture(t)
	genesisCtx := g.ctx.WithBlockHeight(0)
	g.ak.InitGenesis(genesisCtx, imported)
	require.Equal(t, allHistory(t, f, oldCtx, f.ak), allHistory(t, g, genesisCtx, g.ak))
	require.Equal(t, indexEntries(t, oldCtx, f.ak), indexEntries(t, genesisCtx, g.ak))

	// The new chain reaches height 5 again and the account rotates.
	newCtx := g.ctx.WithBlockHeight(5)
	require.NoError(t, g.ak.ApplyRekey(newCtx, g.ak.GetAccount(newCtx, addr), k2))

	require.Equal(t, []uint64{0, 1}, historyKeys(t, newCtx, g.ak, addr))
	hist := historyFor(t, newCtx, g.ak, addr)
	require.Len(t, hist, 2)
	require.True(t, hist[0].PubKey.GetCachedValue().(cryptotypes.PubKey).Equals(k0))
	require.True(t, hist[1].PubKey.GetCachedValue().(cryptotypes.PubKey).Equals(k1))
	require.Equal(t, int64(5), hist[0].ReplacedAtHeight)
	require.Equal(t, int64(5), hist[1].ReplacedAtHeight)
	require.Equal(t, []collections.Pair[sdk.AccAddress, sdk.AccAddress]{
		collections.Join(sdk.AccAddress(k2.Address()), addr),
	}, indexEntries(t, newCtx, g.ak))
}

// TestRemoveAccount_RekeyedClearsIndex checks that removing a rekeyed account
// removes its RekeyIndex pair and its pubkey history, so RekeyedAccounts does
// not return a deleted address, the index matches what InitGenesis rebuilds
// from accounts, and the export passes ValidateGenesis, which rejects history
// that has no account.
func TestRemoveAccount_RekeyedClearsIndex(t *testing.T) {
	f := newRekeyFixture(t)
	f.ak.InitGenesis(f.ctx, *types.DefaultGenesisState())
	k0 := secp256k1.GenPrivKey().PubKey()
	k1 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addr := acc.GetAddress()
	require.NoError(t, f.ak.ApplyRekey(f.ctx, acc, k1))
	require.Len(t, indexEntries(t, f.ctx, f.ak), 1)

	f.ak.RemoveAccount(f.ctx, f.ak.GetAccount(f.ctx, addr))

	require.Nil(t, f.ak.GetAccount(f.ctx, addr))
	require.Empty(t, indexEntries(t, f.ctx, f.ak))
	require.Empty(t, historyFor(t, f.ctx, f.ak, addr))
	require.NoError(t, types.ValidateGenesis(*f.ak.ExportGenesis(f.ctx)))

	// Recreating the address with its natural key also exports cleanly.
	newAccountWithKey(t, f.ctx, f.ak, k0)
	require.NoError(t, types.ValidateGenesis(*f.ak.ExportGenesis(f.ctx)))
}

// TestRemoveAccount_RotatedBackClearsHistory checks that removing an account
// that rotated back to its natural key, so it is not in RekeyIndex, still
// removes its pubkey history.
func TestRemoveAccount_RotatedBackClearsHistory(t *testing.T) {
	f := newRekeyFixture(t)
	f.ak.InitGenesis(f.ctx, *types.DefaultGenesisState())
	k0 := secp256k1.GenPrivKey().PubKey()
	k1 := secp256k1.GenPrivKey().PubKey()
	acc := newAccountWithKey(t, f.ctx, f.ak, k0)
	addr := acc.GetAddress()
	require.NoError(t, f.ak.ApplyRekey(f.ctx, acc, k1))
	require.NoError(t, f.ak.ApplyRekey(f.ctx, f.ak.GetAccount(f.ctx, addr), k0))
	require.Len(t, historyFor(t, f.ctx, f.ak, addr), 2)

	f.ak.RemoveAccount(f.ctx, f.ak.GetAccount(f.ctx, addr))

	require.Empty(t, historyFor(t, f.ctx, f.ak, addr))
	require.NoError(t, types.ValidateGenesis(*f.ak.ExportGenesis(f.ctx)))
}
