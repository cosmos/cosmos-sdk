package types_test

import (
	"crypto/sha256"
	"fmt"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/require"
	"math/rand"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/cosmos/cosmos-sdk/types"
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

func TestInsertMemPool(t *testing.T) {
	mPool := types.NewBTreeMempool(smallSize)
	ctx := types.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	tx := simulateTx(ctx)
	err := mPool.Insert(ctx, TxWSize{tx})
	require.NoError(t, err)
}

func BenchmarkBtreeMempool_Insert(b *testing.B) {
	mPool := types.NewBTreeMempool(smallSize)
	ctx := types.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	tx := simulateTx(ctx)
	for i := 0; i < b.N; i++ {
		mPool.Insert(ctx, TxWSize{tx})
	}
}

func BenchmarkBtreeMempool_Select(b *testing.B) {
	mPool := types.NewBTreeMempool(smallSize)
	ctx := types.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000000)
	for _, t := range transactions {
		mPool.Insert(ctx, TxWSize{t})
	}
	for i := 0; i < b.N; i++ {
		mPool.Select(ctx, nil, 10000)
	}
}

func simulateManyTx(ctx types.Context, n int) []types.Tx {
	transactions := make([]types.Tx, n)
	for i := 0; i < n; i++ {
		tx := simulateTx(ctx)
		transactions[i] = tx
	}
	return transactions
}

func simulateTx(ctx types.Context) types.Tx {
	s := rand.NewSource(1)
	r := rand.New(s)
	msg := group.MsgUpdateGroupMembers{
		GroupId:       1,
		Admin:         "test",
		MemberUpdates: []group.MemberRequest{},
	}
	fees, _ := simtypes.RandomFees(r, ctx, types.NewCoins(types.NewCoin(coinDenom, types.NewInt(100000000))))

	txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
	accounts := simtypes.RandomAccounts(r, 2)

	tx, _ := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]types.Msg{&msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{acc.GetAccountNumber()},
		[]uint64{acc.GetSequence()},
		accounts[0].PrivKey,
	)
	return tx
}

type TxWSize struct {
	types.Tx
}

func (t TxWSize) Size() int {
	return 10

}

func (t TxWSize) GetHash() [32]byte {
	sha1V := sha256.Sum256([]byte(fmt.Sprintf("%#v", t.Tx)))
	return sha1V
}
