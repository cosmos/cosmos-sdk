package mempool_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// testPubKey is a dummy implementation of PubKey used for testing.
type testPubKey struct {
	address sdk.AccAddress
}

func (t testPubKey) Reset() { panic("implement me") }

func (t testPubKey) String() string { panic("implement me") }

func (t testPubKey) ProtoMessage() { panic("implement me") }

func (t testPubKey) Address() cryptotypes.Address { return t.address.Bytes() }

func (t testPubKey) Bytes() []byte { panic("implement me") }

func (t testPubKey) VerifySignature(msg []byte, sig []byte) bool { panic("implement me") }

func (t testPubKey) Equals(key cryptotypes.PubKey) bool { panic("implement me") }

func (t testPubKey) Type() string { panic("implement me") }

// testTx is a dummy implementation of Tx used for testing.
type testTx struct {
	id       int
	priority int64
	nonce    uint64
	address  sdk.AccAddress
}

func (tx testTx) GetSigners() []sdk.AccAddress { panic("implement me") }

func (tx testTx) GetPubKeys() ([]cryptotypes.PubKey, error) { panic("GetPubKeys not implemented") }

func (tx testTx) GetSignaturesV2() (res []txsigning.SignatureV2, err error) {
	res = append(res, txsigning.SignatureV2{
		PubKey:   testPubKey{address: tx.address},
		Data:     nil,
		Sequence: tx.nonce})

	return res, nil
}

var (
	_ sdk.Tx                  = (*testTx)(nil)
	_ mempool.Tx              = (*testTx)(nil)
	_ signing.SigVerifiableTx = (*testTx)(nil)
	_ cryptotypes.PubKey      = (*testPubKey)(nil)
)

func (tx testTx) Size() int64 { return 1 }

func (tx testTx) GetMsgs() []sdk.Msg { return nil }

func (tx testTx) ValidateBasic() error { return nil }

func (tx testTx) String() string {
	return fmt.Sprintf("tx a: %s, p: %d, n: %d", tx.address, tx.priority, tx.nonce)
}

type sigErrTx struct {
	getSigs func() ([]txsigning.SignatureV2, error)
}

func (_ sigErrTx) Size() int64 { return 0 }

func (_ sigErrTx) GetMsgs() []sdk.Msg { return nil }

func (_ sigErrTx) ValidateBasic() error { return nil }

func (_ sigErrTx) GetSigners() []sdk.AccAddress { return nil }

func (_ sigErrTx) GetPubKeys() ([]cryptotypes.PubKey, error) { return nil, nil }

func (t sigErrTx) GetSignaturesV2() ([]txsigning.SignatureV2, error) { return t.getSigs() }

func (s *MempoolTestSuite) TestDefaultMempool() {
	t := s.T()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 10)
	txCount := 1000
	var txs []testTx

	for i := 0; i < txCount; i++ {
		acc := accounts[i%len(accounts)]
		tx := testTx{
			address:  acc.Address,
			priority: rand.Int63(),
		}
		txs = append(txs, tx)
	}

	// same sender-nonce just overwrites a tx
	for _, tx := range txs {
		ctx = ctx.WithPriority(tx.priority)
		err := s.mempool.Insert(ctx, tx)
		require.NoError(t, err)
	}
	require.Equal(t, len(accounts), s.mempool.CountTx())

	// distinct sender-nonce should not overwrite a tx
	s.resetMempool()
	for i, tx := range txs {
		tx.nonce = uint64(i)
		err := s.mempool.Insert(ctx, tx)
		require.NoError(t, err)
	}
	require.Equal(t, txCount, s.mempool.CountTx())

	sel, err := s.mempool.Select(nil, 13)
	require.NoError(t, err)
	require.Equal(t, 13, len(sel))

	// a tx which does not implement SigVerifiableTx should not be inserted
	tx := &sigErrTx{getSigs: func() ([]txsigning.SignatureV2, error) {
		return nil, fmt.Errorf("error")
	}}
	require.Error(t, s.mempool.Insert(ctx, tx))
	require.Error(t, s.mempool.Remove(tx))
	tx.getSigs = func() ([]txsigning.SignatureV2, error) {
		return nil, nil
	}
	require.Error(t, s.mempool.Insert(ctx, tx))
	require.Error(t, s.mempool.Remove(tx))

	// removing a tx not in the mempool should error
	s.resetMempool()
	require.NoError(t, s.mempool.Insert(ctx, txs[0]))
	require.ErrorIs(t, s.mempool.Remove(txs[1]), mempool.ErrTxNotFound)

	// inserting a tx with a different priority should overwrite the old tx
	newPriorityTx := testTx{
		address:  txs[0].address,
		priority: txs[0].priority + 1,
		nonce:    txs[0].nonce,
	}
	require.NoError(t, s.mempool.Insert(ctx, newPriorityTx))
	require.Equal(t, 1, s.mempool.CountTx())
}

type txSpec struct {
	i int
	p int
	n int
	a sdk.AccAddress
}

func (tx txSpec) String() string {
	return fmt.Sprintf("[tx i: %d, a: %s, p: %d, n: %d]", tx.i, tx.a, tx.p, tx.n)
}

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

			//mempool.DebugPrintKeys(pool)

			orderedTxs, err := pool.Select(nil, 1000)
			require.NoError(t, err)
			var txOrder []int
			for _, tx := range orderedTxs {
				txOrder = append(txOrder, tx.(testTx).id)
			}
			for _, tx := range orderedTxs {
				require.NoError(t, pool.Remove(tx))
			}

			require.Equal(t, tt.order, txOrder)

			require.NoError(t, mempool.IsEmpty(pool))
		})
	}
}

type MempoolTestSuite struct {
	suite.Suite
	numTxs      int
	numAccounts int
	iterations  int
	mempool     mempool.Mempool
}

func (s *MempoolTestSuite) resetMempool() {
	s.iterations = 0
	s.mempool = mempool.NewNonceMempool()
}

func (s *MempoolTestSuite) SetupTest() {
	s.numTxs = 1000
	s.numAccounts = 100
	s.resetMempool()
}

func TestMempoolTestSuite(t *testing.T) {
	suite.Run(t, new(MempoolTestSuite))
}
