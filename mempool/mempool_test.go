package mempool_test

import (
	mempool2 "github.com/cosmos/cosmos-sdk/mempool"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/require"
	"math/rand"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

var (
	smallSize = 10
	coinDenom = "aCoin"
	acc       = authtypes.NewEmptyModuleAccount("anaccount")
)

type ConfigurableBenchmarker struct {
	Name     string
	GetBench func(input []byte) func(b *testing.B)
}

func TestNewBTreeMempool(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	require.Equal(t, 1000, len(transactions))
	mempool := mempool2.NewBTreeMempool(1000)

	for _, tx := range transactions {
		ctx.WithPriority(rand.Int63())
		err := mempool.Insert(ctx, tx.(mempool2.MempoolTx))
		require.NoError(t, err)
	}
}

func TestInsertMemPool(t *testing.T) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	tx := simulateTx(ctx)
	err := mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	require.NoError(t, err)
}

func TestSelectMempool(t *testing.T) {
	maxBytes := 10
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
	selectedTx, err := mPool.Select(ctx, nil, maxBytes)
	require.NoError(t, err)
	actualBytes := 0
	for _, selectedTx := range selectedTx {
		actualBytes += selectedTx.Size()
	}
	require.LessOrEqual(t, maxBytes, actualBytes)

}

func TestNewStatefulMempool(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	require.Equal(t, 1000, len(transactions))
	mempool := mempool2.NewStatefulMempool()

	for _, tx := range transactions {
		ctx.WithPriority(rand.Int63())
		err := mempool.Insert(ctx, tx.(mempool2.MempoolTx))
		require.NoError(t, err)
	}
}

func BenchmarkBtreeMempool_Insert(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	tx := simulateTx(ctx)
	for i := 0; i < b.N; i++ {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
}

func BenchmarkBtreeMempool_Insert_100(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 100)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
}

func BenchmarkBtreeMempool_Insert_1000(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
}

func BenchmarkBtreeMempool_Insert_100000(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 100000)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
}

func BenchmarkBtreeMempool_Select_1000(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
	for i := 0; i < b.N; i++ {
		mPool.Select(ctx, nil, 1000)
	}
}

func BenchmarkBtreeMempool_Select_100000(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 100000)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
	for i := 0; i < b.N; i++ {
		mPool.Select(ctx, nil, 10000)
	}
}

func BenchmarkBtreeMempool_Select(b *testing.B) {
	mPool := mempool2.NewBTreeMempool(smallSize)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000000)
	for _, tx := range transactions {
		mPool.Insert(ctx, tx.(mempool2.MempoolTx))
	}
	for i := 0; i < b.N; i++ {
		mPool.Select(ctx, nil, 10000)
	}
}

func simulateManyTx(ctx sdk.Context, n int) []sdk.Tx {
	transactions := make([]sdk.Tx, n)
	for i := 0; i < n; i++ {
		tx := simulateTx(ctx)
		transactions[i] = tx
	}
	return transactions
}

func simulateTx(ctx sdk.Context) sdk.Tx {
	acc := authtypes.NewEmptyModuleAccount("anaccount")

	s := rand.NewSource(1)
	r := rand.New(s)
	msg := group.MsgUpdateGroupMembers{
		GroupId:       1,
		Admin:         acc.Address,
		MemberUpdates: []group.MemberRequest{},
	}
	fees, _ := simtypes.RandomFees(r, ctx, sdk.NewCoins(sdk.NewCoin("coin", sdk.NewInt(100000000))))

	txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
	accounts := simtypes.RandomAccounts(r, 2)

	tx, _ := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{&msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{acc.GetAccountNumber()},
		[]uint64{acc.GetSequence()},
		accounts[0].PrivKey,
	)
	return tx
}
