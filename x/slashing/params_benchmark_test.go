package slashing_test

import (
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtestutil "github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type paramsCountingCodec struct {
	codec.BinaryCodec
	reads uint64
}

func (c *paramsCountingCodec) Unmarshal(bz []byte, ptr proto.Message) error {
	if _, ok := ptr.(*slashingtypes.Params); ok {
		c.reads++
	}
	return c.BinaryCodec.Unmarshal(bz, ptr)
}

func setupParamsBenchmark(tb testing.TB, n int) (sdk.Context, slashingkeeper.Keeper, *paramsCountingCodec, *storetypes.KVStoreKey) {
	tb.Helper()
	key := storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(tb, key, storetypes.NewTransientStoreKey("transient_test"))
	tb.Cleanup(func() { require.NoError(tb, testCtx.DB.Close()) })
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1, Time: time.Unix(1, 0).UTC()})
	enc := moduletestutil.MakeTestEncodingConfig()
	cdc := &paramsCountingCodec{BinaryCodec: enc.Codec}
	staking := slashingtestutil.NewMockStakingKeeper(gomock.NewController(tb))
	staking.EXPECT().IsValidatorJailed(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	k := slashingkeeper.NewKeeper(cdc, enc.Amino, runtime.NewKVStoreService(key), staking, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	require.NoError(tb, k.SetParams(ctx, slashingtestutil.TestParams()))
	votes := make([]abci.VoteInfo, n)
	for i := range votes {
		addr := make(sdk.ConsAddress, 20)
		binary.BigEndian.PutUint64(addr[12:], uint64(i+1))
		info := slashingtypes.NewValidatorSigningInfo(addr, 0, 0, time.Unix(0, 0).UTC(), false, 0)
		require.NoError(tb, k.SetValidatorSigningInfo(ctx, addr, info))
		votes[i] = abci.VoteInfo{Validator: abci.Validator{Address: addr, Power: 100}, BlockIdFlag: cmtproto.BlockIDFlagCommit}
	}
	return ctx.WithVoteInfos(votes), k, cdc, key
}

// The benchmark executes the production BeginBlocker and slashing store path.
// Only the unrelated staking lookup is mocked; all validators signed the block.
func BenchmarkBeginBlockerParams(b *testing.B) {
	for _, n := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			ctx, k, cdc, _ := setupParamsBenchmark(b, n)
			require.NoError(b, slashing.BeginBlocker(ctx, k))
			cdc.reads = 0
			gasStart := ctx.GasMeter().GasConsumed()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := slashing.BeginBlocker(ctx, k); err != nil {
					b.Fatal(err)
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(cdc.reads)/float64(b.N), "params/op")
			b.ReportMetric(float64(ctx.GasMeter().GasConsumed()-gasStart)/float64(b.N), "gas/op")
		})
	}
}

func TestBeginBlockerSigningInfoForVotes(t *testing.T) {
	for _, n := range []int{0, 1, 10, 100} {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			ctx, k, _, _ := setupParamsBenchmark(t, n)
			require.NoError(t, slashing.BeginBlocker(ctx, k))
			for _, vote := range ctx.VoteInfos() {
				info, err := k.GetValidatorSigningInfo(ctx, sdk.ConsAddress(vote.Validator.Address))
				require.NoError(t, err)
				require.Equal(t, int64(1), info.IndexOffset)
				require.Zero(t, info.MissedBlocksCounter)
			}
		})
	}
}

func TestBeginBlockerInvalidParams(t *testing.T) {
	ctx, k, _, key := setupParamsBenchmark(t, 1)
	addr := sdk.ConsAddress(ctx.VoteInfos()[0].Validator.Address)
	before, err := k.GetValidatorSigningInfo(ctx, addr)
	require.NoError(t, err)
	ctx.KVStore(key).Set(slashingtypes.ParamsKey, []byte{0xff})

	require.Error(t, slashing.BeginBlocker(ctx, k))
	after, err := k.GetValidatorSigningInfo(ctx, addr)
	require.NoError(t, err)
	require.Equal(t, before, after)
	require.Empty(t, ctx.EventManager().Events())
}

func TestBeginBlockerMissingSigningInfo(t *testing.T) {
	ctx, k, _, _ := setupParamsBenchmark(t, 1)
	addr := sdk.ConsAddress(ctx.VoteInfos()[0].Validator.Address)
	require.NoError(t, k.DeleteValidatorSigningInfo(ctx, addr))

	require.ErrorIs(t, slashing.BeginBlocker(ctx, k), slashingtypes.ErrNoSigningInfoFound)
	require.Empty(t, ctx.EventManager().Events())
}
