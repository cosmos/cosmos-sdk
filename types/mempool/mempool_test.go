package mempool_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

type testTx struct {
	hash     [32]byte
	priority int64
	nonce    uint64
	sender   string
}

func (tx testTx) GetSigners() []sdk.AccAddress {
	// TODO multi sender
	return []sdk.AccAddress{sdk.AccAddress(tx.sender)}
}

func (tx testTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	panic("GetPubkeys not implemented")
}

func (tx testTx) GetSignaturesV2() ([]signing2.SignatureV2, error) {
	// TODO multi sender
	return []signing2.SignatureV2{
		{
			PubKey:   nil,
			Data:     nil,
			Sequence: tx.nonce,
		},
	}, nil
}

func newTestTx(priority int64, nonce uint64, sender string) testTx {
	hash := make([]byte, 32)
	rand.Read(hash)
	return testTx{
		hash:     *(*[32]byte)(hash),
		priority: priority,
		nonce:    nonce,
		sender:   sender,
	}
}

var (
	_ sdk.Tx                  = (*testTx)(nil)
	_ mempool.Tx              = (*testTx)(nil)
	_ signing.SigVerifiableTx = (*testTx)(nil)
)

func (tx testTx) GetHash() [32]byte {
	return tx.hash
}

func (tx testTx) Size() int {
	return 10
}

func (tx testTx) GetMsgs() []sdk.Msg {
	return nil
}

func (tx testTx) ValidateBasic() error {
	return nil
}

func TestNewBTreeMempool(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	transactions := simulateManyTx(ctx, 1000)
	require.Equal(t, 1000, len(transactions))
	mp := mempool.NewBTreeMempool(1000)

	for _, tx := range transactions {
		ctx.WithPriority(rand.Int63())
		err := mp.Insert(ctx, tx.(mempool.Tx))
		require.NoError(t, err)
	}
	require.Equal(t, 1000, mp.CountTx())
}

func TestNewStatefulMempool(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	// general test
	transactions := simulateManyTx(ctx, 1000)
	require.Equal(t, 1000, len(transactions))
	mp := mempool.NewBTreeMempool(1000)

	for _, tx := range transactions {
		ctx.WithPriority(rand.Int63())
		err := mp.Insert(ctx, tx.(mempool.Tx))
		require.NoError(t, err)
	}
	require.Equal(t, 1000, mp.CountTx())
}

func TestTxOrder(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	txs := []testTx{
		{hash: [32]byte{1}, priority: 21, nonce: 4, sender: "a"},
		{hash: [32]byte{2}, priority: 8, nonce: 3, sender: "a"},
		{hash: [32]byte{3}, priority: 6, nonce: 2, sender: "a"},
		{hash: [32]byte{4}, priority: 15, nonce: 1, sender: "b"},
		{hash: [32]byte{5}, priority: 20, nonce: 1, sender: "a"},
	}
	order := []byte{5, 4, 3, 2, 1}
	tests := []struct {
		name  string
		txs   []testTx
		pool  mempool.Mempool
		order []byte
	}{
		{name: "BTreeMempool", txs: txs, order: order, pool: mempool.NewBTreeMempool(1000)},
		{name: "StatefulMempool", txs: txs, order: order, pool: mempool.NewStatefulMempool()},
		{name: "Stateful_3nodes", txs: []testTx{
			{hash: [32]byte{1}, priority: 21, nonce: 4, sender: "a"},
			{hash: [32]byte{4}, priority: 15, nonce: 1, sender: "b"},
			{hash: [32]byte{5}, priority: 20, nonce: 1, sender: "a"},
		},
			order: []byte{5, 1, 4}, pool: mempool.NewStatefulMempool()},
		{name: "GraphMempool", txs: txs, order: order, pool: mempool.NewGraph()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tx := range tt.txs {
				c := ctx.WithPriority(tx.priority)
				err := tt.pool.Insert(c, tx)
				require.NoError(t, err)
			}
			// TODO uncomment
			//require.Equal(t, len(tt.txs), tt.pool.CountTx())

			orderedTxs, err := tt.pool.Select(ctx, nil, 1000)
			require.NoError(t, err)
			require.Equal(t, len(tt.txs), len(orderedTxs))
			for i, h := range tt.order {
				require.Equal(t, h, orderedTxs[i].(testTx).hash[0])
			}
		})
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
		Admin:         "test",
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
