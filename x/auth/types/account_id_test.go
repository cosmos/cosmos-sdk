package types_test

import (
	"math"
	"testing"

	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

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

func TestGenerateID_TopBitAlwaysSet(t *testing.T) {
	cases := []struct {
		name     string
		height   int64
		appHash  []byte
		txIndex  int
		msgIndex int
	}{
		{"zero values", 0, nil, 0, 0},
		{"defaults (negative indices)", 0, nil, -1, -1},
		{"typical block", 100, []byte{0xab, 0xcd}, 3, 1},
		{"height 1 empty apphash", 1, []byte{}, 0, 0},
		{"max height", math.MaxInt64, nil, 0, 0},
		{"max indices", 1, nil, math.MaxInt32, math.MaxInt32},
		{"large apphash", 1, make([]byte, 1024), 0, 0},
		{"negative one indices", 0, []byte{0xff}, -1, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := types.GenerateID(makeCtx(tc.height, tc.appHash, tc.txIndex, tc.msgIndex))
			require.NotZero(t, id&(uint64(1)<<63), "top bit must be set")
			require.GreaterOrEqual(t, id, uint64(1)<<63)
		})
	}
}

func TestGenerateID_Deterministic(t *testing.T) {
	ctx := makeCtx(42, []byte("apphash"), 5, 2)
	id1 := types.GenerateID(ctx)
	id2 := types.GenerateID(ctx)
	require.Equal(t, id1, id2, "same inputs must produce same ID")
}

func TestGenerateID_DifferentHeightsProduceDifferentIDs(t *testing.T) {
	appHash := []byte("hash")
	ids := make(map[uint64]struct{})
	for h := int64(0); h < 200; h++ {
		id := types.GenerateID(makeCtx(h, appHash, 0, 0))
		require.NotContains(t, ids, id, "height %d produced duplicate ID", h)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentAppHashesProduceDifferentIDs(t *testing.T) {
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		appHash := []byte{byte(i), byte(i >> 8)}
		id := types.GenerateID(makeCtx(1, appHash, 0, 0))
		require.NotContains(t, ids, id, "appHash %v produced duplicate ID", appHash)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentTxIndicesProduceDifferentIDs(t *testing.T) {
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		id := types.GenerateID(makeCtx(1, []byte("hash"), i, 0))
		require.NotContains(t, ids, id, "txIndex %d produced duplicate ID", i)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_DifferentMsgIndicesProduceDifferentIDs(t *testing.T) {
	ids := make(map[uint64]struct{})
	for i := 0; i < 200; i++ {
		id := types.GenerateID(makeCtx(1, []byte("hash"), 0, i))
		require.NotContains(t, ids, id, "msgIndex %d produced duplicate ID", i)
		ids[id] = struct{}{}
	}
}

func TestGenerateID_NilVsEmptyAppHash(t *testing.T) {
	idNil := types.GenerateID(makeCtx(1, nil, 0, 0))
	idEmpty := types.GenerateID(makeCtx(1, []byte{}, 0, 0))
	// Both should have top bit set regardless
	require.NotZero(t, idNil&(uint64(1)<<63))
	require.NotZero(t, idEmpty&(uint64(1)<<63))
	// nil and empty both write zero bytes to the hasher, so they should be equal
	require.Equal(t, idNil, idEmpty, "nil and empty AppHash should hash identically")
}

func TestGenerateID_SingleBitInputDifferences(t *testing.T) {
	// Changing a single input bit should change the output (avalanche property of SHA-256).
	base := makeCtx(0, []byte{0x00}, 0, 0)
	baseID := types.GenerateID(base)

	flippedAppHash := makeCtx(0, []byte{0x01}, 0, 0)
	require.NotEqual(t, baseID, types.GenerateID(flippedAppHash), "flipping one AppHash bit should change ID")

	flippedHeight := makeCtx(1, []byte{0x00}, 0, 0)
	require.NotEqual(t, baseID, types.GenerateID(flippedHeight), "flipping height bit should change ID")

	flippedTxIndex := makeCtx(0, []byte{0x00}, 1, 0)
	require.NotEqual(t, baseID, types.GenerateID(flippedTxIndex), "flipping txIndex bit should change ID")

	flippedMsgIndex := makeCtx(0, []byte{0x00}, 0, 1)
	require.NotEqual(t, baseID, types.GenerateID(flippedMsgIndex), "flipping msgIndex bit should change ID")
}

func TestGenerateID_BoundaryValues(t *testing.T) {
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
			id := types.GenerateID(makeCtx(tc.height, nil, tc.txIndex, tc.msgIndex))
			require.NotZero(t, id&(uint64(1)<<63), "top bit must be set for %s", tc.name)
		})
	}
}

func TestGenerateID_NoCollisionAcrossCombinations(t *testing.T) {
	// Generate IDs across a range of (height, txIndex, msgIndex) combos and verify no collisions.
	ids := make(map[uint64]struct{})
	for h := int64(0); h < 10; h++ {
		for tx := 0; tx < 10; tx++ {
			for msg := 0; msg < 10; msg++ {
				id := types.GenerateID(makeCtx(h, []byte("app"), tx, msg))
				require.NotContains(t, ids, id, "collision at h=%d tx=%d msg=%d", h, tx, msg)
				ids[id] = struct{}{}
			}
		}
	}
	require.Len(t, ids, 1000)
}
