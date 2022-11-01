package mempool_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func (s *MempoolTestSuite) TestTxOrder() {
	t := s.T()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 5)
	sa := accounts[0].Address
	sb := accounts[1].Address
	sc := accounts[2].Address

	tests := []struct {
		txs   []txSpec
		order []int
		fail  bool
	}{
		{
			txs: []txSpec{
				{p: 21, n: 4, a: sa},
				{p: 8, n: 3, a: sa},
				{p: 6, n: 2, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 20, n: 1, a: sa},
			},
			order: []int{3, 4, 2, 1, 0},
		},
		{
			txs: []txSpec{
				{p: 3, n: 0, a: sa},
				{p: 5, n: 1, a: sa},
				{p: 9, n: 2, a: sa},
				{p: 6, n: 0, a: sb},
				{p: 5, n: 1, a: sb},
				{p: 8, n: 2, a: sb},
			},
			order: []int{3, 0, 4, 1, 5, 2},
		},
		{
			txs: []txSpec{
				{p: 21, n: 4, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 20, n: 1, a: sa},
			},
			order: []int{1, 2, 0},
		},
		{
			txs: []txSpec{
				{p: 50, n: 3, a: sa},
				{p: 30, n: 2, a: sa},
				{p: 10, n: 1, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 21, n: 2, a: sb},
			},
			order: []int{3, 2, 4, 1, 0},
		},
		{
			txs: []txSpec{
				{p: 50, n: 3, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 99, n: 1, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 8, n: 2, a: sb},
			},
			order: []int{3, 2, 4, 1, 0},
		},
		{
			txs: []txSpec{
				{p: 30, a: sa, n: 2},
				{p: 20, a: sb, n: 1},
				{p: 15, a: sa, n: 1},
				{p: 10, a: sa, n: 0},
				{p: 8, a: sb, n: 0},
				{p: 6, a: sa, n: 3},
				{p: 4, a: sb, n: 3},
			},
			order: []int{4, 3, 1, 2, 0, 6, 5},
		},
		{
			txs: []txSpec{
				{p: 30, n: 2, a: sa},
				{p: 20, a: sb, n: 1},
				{p: 15, a: sa, n: 1},
				{p: 10, a: sa, n: 0},
				{p: 8, a: sb, n: 0},
				{p: 6, a: sa, n: 3},
				{p: 4, a: sb, n: 3},
				{p: 2, a: sc, n: 0},
				{p: 7, a: sc, n: 3},
			},
			order: []int{4, 3, 7, 1, 2, 0, 6, 5, 8},
		},
		{
			txs: []txSpec{
				{p: 6, n: 1, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 5, n: 1, a: sb},
				{p: 99, n: 2, a: sb},
			},
			order: []int{2, 0, 3, 1},
		},
		{
			// if all txs have the same priority they will be ordered lexically sender address, and nonce with the
			// sender.
			txs: []txSpec{
				{p: 10, n: 7, a: sc},
				{p: 10, n: 8, a: sc},
				{p: 10, n: 9, a: sc},
				{p: 10, n: 1, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 10, n: 3, a: sa},
				{p: 10, n: 4, a: sb},
				{p: 10, n: 5, a: sb},
				{p: 10, n: 6, a: sb},
			},
			order: []int{3, 4, 5, 6, 7, 8, 0, 1, 2},
		},
		/*
			The next 4 tests are different permutations of the same set:

			  		{p: 5, n: 1, a: sa},
					{p: 10, n: 2, a: sa},
					{p: 20, n: 2, a: sb},
					{p: 5, n: 1, a: sb},
					{p: 99, n: 2, a: sc},
					{p: 5, n: 1, a: sc},

			which exercises the actions required to resolve priority ties.
		*/
		{
			txs: []txSpec{
				{p: 5, n: 1, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 5, n: 1, a: sb},
				{p: 99, n: 2, a: sb},
			},
			order: []int{2, 0, 3, 1},
		},
		{
			txs: []txSpec{
				{p: 5, n: 1, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 20, n: 2, a: sb},
				{p: 5, n: 1, a: sb},
				{p: 99, n: 2, a: sc},
				{p: 5, n: 1, a: sc},
			},
			order: []int{3, 0, 5, 2, 1, 4},
		},
		{
			txs: []txSpec{
				{p: 5, n: 1, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 5, n: 1, a: sb},
				{p: 20, n: 2, a: sb},
				{p: 5, n: 1, a: sc},
				{p: 99, n: 2, a: sc},
			},
			order: []int{2, 0, 4, 3, 1, 5},
		},
		{
			txs: []txSpec{
				{p: 5, n: 1, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 5, n: 1, a: sc},
				{p: 20, n: 2, a: sc},
				{p: 5, n: 1, a: sb},
				{p: 99, n: 2, a: sb},
			},
			order: []int{4, 0, 2, 5, 1, 3},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			pool := s.mempool

			// create test txs and insert into mempool
			for i, ts := range tt.txs {
				tx := testTx{id: i, priority: int64(ts.p), nonce: uint64(ts.n), address: ts.a}
				c := ctx.WithPriority(tx.priority)
				err := pool.Insert(c, tx)
				require.NoError(t, err)
			}

			itr := pool.Select(ctx, nil)
			orderedTxs := fetchTxs(itr, 1000)
			var txOrder []int
			for _, tx := range orderedTxs {
				txOrder = append(txOrder, tx.(testTx).id)
			}
			for _, tx := range orderedTxs {
				require.NoError(t, pool.Remove(tx))
			}

			require.Equal(t, tt.order, txOrder)
			require.Equal(t, 0, pool.CountTx())
		})
	}
}
