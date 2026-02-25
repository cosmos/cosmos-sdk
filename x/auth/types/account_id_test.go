package types_test

import (
	"math"
	"testing"

	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func makeCtx(height int64, appHash []byte, txIndex, msgIndex int) sdk.Context {
	ctx := sdk.NewContext(nil, cmttypes.Header{
		Height:  height,
		AppHash: appHash,
	}, false, nil)
	return ctx.WithTxIndex(txIndex).WithMsgIndex(msgIndex)
}

func randomBaseAccount() *types.BaseAccount {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	return types.NewBaseAccountWithAddress(addr)
}

func fixedBaseAccount(addrByte byte) *types.BaseAccount {
	addr := sdk.AccAddress(make([]byte, 20))
	addr[0] = addrByte
	return types.NewBaseAccountWithAddress(addr)
}

func moduleAccount(name string) *types.ModuleAccount {
	return types.NewEmptyModuleAccount(name)
}

func TestGenerateID_TopBitAlwaysSet(t *testing.T) {
	acc := randomBaseAccount()
	cases := []struct {
		name     string
		height   int64
		appHash  []byte
		txIndex  int
		msgIndex int
		acc      sdk.AccountI
	}{
		{"zero values", 0, nil, 0, 0, acc},
		{"defaults (negative indices)", 0, nil, -1, -1, acc},
		{"typical block", 100, []byte{0xab, 0xcd}, 3, 1, acc},
		{"height 1 empty apphash", 1, []byte{}, 0, 0, acc},
		{"max height", math.MaxInt64, nil, 0, 0, acc},
		{"max indices", 1, nil, math.MaxInt32, math.MaxInt32, acc},
		{"large apphash", 1, make([]byte, 1024), 0, 0, acc},
		{"negative one indices", 0, []byte{0xff}, -1, -1, acc},
		{"module account", 100, []byte{0xab}, 3, 1, moduleAccount("staking")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := types.GenerateID(makeCtx(tc.height, tc.appHash, tc.txIndex, tc.msgIndex), tc.acc)
			require.NotZero(t, id&(uint64(1)<<63), "top bit must be set")
			require.GreaterOrEqual(t, id, uint64(1)<<63)
		})
	}
}

func TestGenerateID_Deterministic(t *testing.T) {
	ctx := makeCtx(42, []byte("apphash"), 5, 2)
	acc := randomBaseAccount()
	id1 := types.GenerateID(ctx, acc)
	id2 := types.GenerateID(ctx, acc)
	require.Equal(t, id1, id2, "same inputs must produce same ID")
}

func TestGenerateID_DifferentHeightsProduceDifferentIDs(t *testing.T) {
	appHash := []byte("hash")
	acc := randomBaseAccount()
	ids := make(map[uint64]struct{})
	for h := int64(0); h < 200; h++ {
		id := types.GenerateID(makeCtx(h, appHash, 0, 0), acc)
		require.NotContains(t, ids, id, "height %d produced duplicate ID", h)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentAppHashesProduceDifferentIDs(t *testing.T) {
	acc := randomBaseAccount()
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		appHash := []byte{byte(i), byte(i >> 8)}
		id := types.GenerateID(makeCtx(1, appHash, 0, 0), acc)
		require.NotContains(t, ids, id, "appHash %v produced duplicate ID", appHash)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentTxIndicesProduceDifferentIDs(t *testing.T) {
	acc := randomBaseAccount()
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		id := types.GenerateID(makeCtx(1, []byte("hash"), i, 0), acc)
		require.NotContains(t, ids, id, "txIndex %d produced duplicate ID", i)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentMsgIndicesProduceDifferentIDs(t *testing.T) {
	acc := randomBaseAccount()
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		id := types.GenerateID(makeCtx(1, []byte("hash"), 0, i), acc)
		require.NotContains(t, ids, id, "msgIndex %d produced duplicate ID", i)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentAddressesProduceDifferentIDs(t *testing.T) {
	ctx := makeCtx(1, []byte("hash"), 0, 0)
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		acc := randomBaseAccount()
		id := types.GenerateID(ctx, acc)
		require.NotContains(t, ids, id, "address %s produced duplicate ID", acc.GetAddress())
		ids[id] = struct{}{}
	}
}

func TestGenerateID_NilVsEmptyAppHash(t *testing.T) {
	acc := randomBaseAccount()
	idNil := types.GenerateID(makeCtx(1, nil, 0, 0), acc)
	idEmpty := types.GenerateID(makeCtx(1, []byte{}, 0, 0), acc)
	// Both should have top bit set regardless
	require.NotZero(t, idNil&(uint64(1)<<63))
	require.NotZero(t, idEmpty&(uint64(1)<<63))
	// nil and empty both write zero bytes to the hasher, so they should be equal
	require.Equal(t, idNil, idEmpty, "nil and empty AppHash should hash identically")
}

func TestGenerateID_SingleBitInputDifferences(t *testing.T) {
	// Changing a single input bit should change the output (avalanche property of SHA-256).
	acc := fixedBaseAccount(0x00)
	base := makeCtx(0, []byte{0x00}, 0, 0)
	baseID := types.GenerateID(base, acc)

	flippedAppHash := makeCtx(0, []byte{0x01}, 0, 0)
	require.NotEqual(t, baseID, types.GenerateID(flippedAppHash, acc), "flipping one AppHash bit should change ID")

	flippedHeight := makeCtx(1, []byte{0x00}, 0, 0)
	require.NotEqual(t, baseID, types.GenerateID(flippedHeight, acc), "flipping height bit should change ID")

	flippedTxIndex := makeCtx(0, []byte{0x00}, 1, 0)
	require.NotEqual(t, baseID, types.GenerateID(flippedTxIndex, acc), "flipping txIndex bit should change ID")

	flippedMsgIndex := makeCtx(0, []byte{0x00}, 0, 1)
	require.NotEqual(t, baseID, types.GenerateID(flippedMsgIndex, acc), "flipping msgIndex bit should change ID")

	flippedAddr := fixedBaseAccount(0x01)
	require.NotEqual(t, baseID, types.GenerateID(base, flippedAddr), "flipping address bit should change ID")
}

func TestGenerateID_BoundaryValues(t *testing.T) {
	acc := randomBaseAccount()
	cases := []struct {
		name     string
		height   int64
		txIndex  int
		msgIndex int
	}{
		{"zero height", 0, 0, 0},
		{"height 1", 1, 0, 0},
		{"max int64 height", math.MaxInt64, 0, 0},
		{"min int64 height", math.MinInt64, 0, 0},
		{"max int txIndex", 0, math.MaxInt, 0},
		{"min int txIndex", 0, math.MinInt, 0},
		{"max int msgIndex", 0, 0, math.MaxInt},
		{"min int msgIndex", 0, 0, math.MinInt},
		{"all max", math.MaxInt64, math.MaxInt, math.MaxInt},
		{"all min", math.MinInt64, math.MinInt, math.MinInt},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := types.GenerateID(makeCtx(tc.height, nil, tc.txIndex, tc.msgIndex), acc)
			require.NotZero(t, id&(uint64(1)<<63), "top bit must be set for %s", tc.name)
		})
	}
}

func TestGenerateID_NoCollisionAcrossCombinations(t *testing.T) {
	// Generate IDs across a range of (height, txIndex, msgIndex) combos and verify no collisions.
	acc := randomBaseAccount()
	ids := make(map[uint64]struct{})
	for h := int64(0); h < 10; h++ {
		for tx := 0; tx < 10; tx++ {
			for msg := 0; msg < 10; msg++ {
				id := types.GenerateID(makeCtx(h, []byte("app"), tx, msg), acc)
				require.NotContains(t, ids, id, "collision at h=%d tx=%d msg=%d", h, tx, msg)
				ids[id] = struct{}{}
			}
		}
	}
	require.Len(t, ids, 1000)
}

func TestGenerateID_ModuleAccountDifferentNames(t *testing.T) {
	// Different module names with the same context should produce different IDs.
	ctx := makeCtx(10, []byte("apphash"), 1, 0)
	names := []string{"staking", "distribution", "gov", "bank", "mint", "bonded_tokens_pool", "not_bonded_tokens_pool"}
	ids := make(map[uint64]struct{})
	for _, name := range names {
		id := types.GenerateID(ctx, moduleAccount(name))
		require.NotContains(t, ids, id, "module %q produced duplicate ID", name)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_ModuleAccountVsBaseAccount(t *testing.T) {
	// A module account and a base account in the same context should produce different IDs,
	// even when the base account has the same address as the module account.
	ctx := makeCtx(10, []byte("apphash"), 1, 0)
	modAcc := moduleAccount("staking")
	baseAcc := types.NewBaseAccountWithAddress(modAcc.GetAddress())
	baseID := types.GenerateID(ctx, baseAcc)
	modID := types.GenerateID(ctx, modAcc)
	require.NotEqual(t, baseID, modID, "module account and base account with same address should produce different IDs")
}

func TestGenerateID_ModuleAccountDeterministic(t *testing.T) {
	ctx := makeCtx(42, []byte("apphash"), 5, 2)
	acc := moduleAccount("distribution")
	id1 := types.GenerateID(ctx, acc)
	id2 := types.GenerateID(ctx, acc)
	require.Equal(t, id1, id2, "same module account inputs must produce same ID")
}

func TestGenerateID_ModuleAccountTopBitSet(t *testing.T) {
	ctx := makeCtx(1, []byte("hash"), 0, 0)
	names := []string{"staking", "distribution", "gov", "bank"}
	for _, name := range names {
		id := types.GenerateID(ctx, moduleAccount(name))
		require.NotZero(t, id&(uint64(1)<<63), "top bit must be set for module %q", name)
	}
}

func TestGenerateID_SameContextDifferentAccountsNeverCollide(t *testing.T) {
	// This is the key test for the address-based disambiguation:
	// multiple new accounts created in the same block/tx/msg context
	// must get different IDs because their addresses differ.
	ctx := makeCtx(1, []byte("apphash"), 0, 0)
	ids := make(map[uint64]struct{})
	for i := 0; i < 500; i++ {
		acc := randomBaseAccount()
		id := types.GenerateID(ctx, acc)
		require.NotContains(t, ids, id, "collision for account %d", i)
		ids[id] = struct{}{}
	}
}
