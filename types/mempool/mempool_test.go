package mempool_test

import (
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"math"
	"math/rand"
	"testing"
	"time"

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
	address  sdk.AccAddress
}

func (tx testTx) GetSigners() []sdk.AccAddress {
	// TODO multi sender
	return []sdk.AccAddress{tx.address}
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

func (tx testTx) String() string {
	return fmt.Sprintf("tx %s, %d, %d", tx.address, tx.priority, tx.nonce)
}

func TestNewStatefulMempool(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	// general test
	transactions := simulateManyTx(ctx, 1000)
	require.Equal(t, 1000, len(transactions))
	mp := mempool.NewDefaultMempool()

	for _, tx := range transactions {
		ctx.WithPriority(rand.Int63())
		err := mp.Insert(ctx, tx.(mempool.Tx))
		require.NoError(t, err)
	}
	require.Equal(t, 1000, mp.CountTx())
}

func TestTxOrder(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 2)
	senderA := accounts[0].Address
	senderB := accounts[1].Address
	txs := []testTx{
		{hash: [32]byte{1}, priority: 21, nonce: 4, address: senderA},
		{hash: [32]byte{2}, priority: 8, nonce: 3, address: senderA},
		{hash: [32]byte{3}, priority: 6, nonce: 2, address: senderA},
		{hash: [32]byte{4}, priority: 15, nonce: 1, address: senderB},
		{hash: [32]byte{5}, priority: 20, nonce: 1, address: senderA},
	}

	order := []byte{5, 4, 3, 2, 1}
	tests := []struct {
		name  string
		txs   []testTx
		pool  mempool.Mempool
		order []byte
	}{
		{name: "StatefulMempool", txs: txs, order: order, pool: mempool.NewDefaultMempool()},
		{name: "Stateful_3nodes", txs: []testTx{
			{hash: [32]byte{1}, priority: 21, nonce: 4, address: senderA},
			{hash: [32]byte{4}, priority: 15, nonce: 1, address: senderB},
			{hash: [32]byte{5}, priority: 20, nonce: 1, address: senderA},
		},
			order: []byte{5, 1, 4}, pool: mempool.NewDefaultMempool()},
		{name: "GraphMempool", txs: txs, order: order, pool: mempool.NewGraph()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, tx := range tt.txs {
				c := ctx.WithPriority(tx.priority)
				err := tt.pool.Insert(c, tx)
				require.NoError(t, err)
			}
			require.Equal(t, len(tt.txs), tt.pool.CountTx())

			orderedTxs, err := tt.pool.Select(ctx, nil, 1000)
			require.NoError(t, err)
			require.Equal(t, len(tt.txs), len(orderedTxs))
			for i, h := range tt.order {
				require.Equal(t, h, orderedTxs[i].(testTx).hash[0])
			}
		})
	}
}

func TestRandomTxOrderManyTimes(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Log("iteration", i)
		TestRandomTxOrder(t)
	}
}

func TestRandomTxOrder(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	numTx := 10

	//seed := time.Now().UnixNano()
	// interesting failing seeds:
	// seed := int64(1663971399133628000)
	seed := int64(1663989445512438000)
	//

	ordered, shuffled := genOrderedTxs(seed, numTx, 3)
	mp := mempool.NewDefaultMempool()

	for _, otx := range shuffled {
		tx := testTx{otx.hash, otx.priority, otx.nonce, otx.address}
		c := ctx.WithPriority(tx.priority)
		err := mp.Insert(c, tx)
		require.NoError(t, err)
	}

	require.Equal(t, numTx, mp.CountTx())

	selected, err := mp.Select(ctx, nil, math.MaxInt)
	var orderedStr, selectedStr string

	for i := 0; i < numTx; i++ {
		otx := ordered[i]
		stx := selected[i].(testTx)
		orderedStr = fmt.Sprintf("%s\n%s, %d, %d; %d", orderedStr, otx.address, otx.priority, otx.nonce, otx.hash[0])
		selectedStr = fmt.Sprintf("%s\n%s, %d, %d; %d", selectedStr, stx.address, stx.priority, stx.nonce, stx.hash[0])
	}

	require.NoError(t, err)
	require.Equal(t, numTx, len(selected))

	errMsg := fmt.Sprintf("Expected order: %v\nGot order: %v\nSeed: %v", orderedStr, selectedStr, seed)

	mempool.DebugPrintKeys(mp)

	for i, tx := range selected {
		msg := fmt.Sprintf("Failed tx at index %d\n%s", i, errMsg)
		require.Equal(t, ordered[i], tx.(testTx), msg)
		require.Equal(t, tx.(testTx).priority, ordered[i].priority, msg)
		require.Equal(t, tx.(testTx).nonce, ordered[i].nonce, msg)
		require.Equal(t, tx.(testTx).address, ordered[i].address, msg)
	}

}

type txKey struct {
	sender   string
	nonce    uint64
	priority int64
	hash     [32]byte
}

func genOrderedTxs(seed int64, maxTx int, numAcc int) (ordered []testTx, shuffled []testTx) {
	r := rand.New(rand.NewSource(seed))
	accountNonces := make(map[string]uint64)
	prange := 10
	randomAccounts := simtypes.RandomAccounts(r, numAcc)
	for _, account := range randomAccounts {
		accountNonces[account.Address.String()] = 0
	}

	getRandAccount := func(notAddress string) simtypes.Account {
		for {
			res := randomAccounts[r.Intn(len(randomAccounts))]
			if res.Address.String() != notAddress {
				return res
			}
		}
	}

	txCursor := int64(10000)
	ptx := testTx{address: getRandAccount("").Address, nonce: 0, priority: txCursor}
	samepChain := make(map[string]bool)
	for i := 0; i < maxTx; {
		var tx testTx
		move := r.Intn(5)
		switch move {
		case 0:
			// same sender, less p
			nonce := ptx.nonce + 1
			tx = testTx{nonce: nonce, address: ptx.address, priority: txCursor - int64(r.Intn(prange)+1)}
			txCursor = tx.priority
		case 1:
			// same sender, same p
			nonce := ptx.nonce + 1
			tx = testTx{nonce: nonce, address: ptx.address, priority: ptx.priority}
		case 2:
			// same sender, greater p
			nonce := ptx.nonce + 1
			tx = testTx{nonce: nonce, address: ptx.address, priority: ptx.priority + int64(r.Intn(prange)+1)}
		case 3:
			// different sender, less p
			sender := getRandAccount(ptx.address.String()).Address
			nonce := accountNonces[sender.String()] + 1
			tx = testTx{nonce: nonce, address: sender, priority: txCursor - int64(r.Intn(prange)+1)}
			txCursor = tx.priority
		case 4:
			// different sender, same p
			sender := getRandAccount(ptx.address.String()).Address
			// disallow generating cycles of same p txs. this is an invalid processing order according to our
			// algorithm decision.
			if _, ok := samepChain[sender.String()]; ok {
				continue
			}
			nonce := accountNonces[sender.String()] + 1
			tx = testTx{nonce: nonce, address: sender, priority: txCursor}
			samepChain[sender.String()] = true
		}
		tx.hash = [32]byte{byte(i)}
		accountNonces[tx.address.String()] = tx.nonce
		ordered = append(ordered, tx)
		ptx = tx
		i++
		if move != 4 {
			samepChain = make(map[string]bool)
		}
	}

	for _, item := range ordered {
		tx := testTx{
			priority: item.priority,
			nonce:    item.nonce,
			address:  item.address,
			hash:     item.hash,
		}
		shuffled = append(shuffled, tx)
	}
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return ordered, shuffled
}

func TestTxOrderN(t *testing.T) {
	numTx := 10

	seed := time.Now().UnixNano()
	ordered, shuffled := genOrderedTxs(seed, numTx, 3)
	require.Equal(t, numTx, len(ordered))
	require.Equal(t, numTx, len(shuffled))

	fmt.Println("ordered")
	for _, tx := range ordered {
		fmt.Printf("%s, %d, %d\n", tx.address, tx.priority, tx.nonce)
	}

	fmt.Println("shuffled")
	for _, tx := range shuffled {
		fmt.Printf("%s, %d, %d\n", tx.address, tx.priority, tx.nonce)
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

type txWithPriority struct {
	priority int64
	tx       sdk.Tx
	address  string
	nonce    uint64 // duplicate from tx.address.sequence
}

func GenTxOrder(ctx sdk.Context, nTx int, nSenders int) (ordered []txWithPriority, shuffled []txWithPriority) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	randomAccounts := simtypes.RandomAccounts(r, nSenders)
	senderNonces := make(map[string]uint64)
	senderLastPriority := make(map[string]int)
	for _, acc := range randomAccounts {
		address := acc.Address.String()
		senderNonces[address] = 1
		senderLastPriority[address] = 999999
	}

	for i := 0; i < nTx; i++ {
		acc := randomAccounts[r.Intn(nSenders)]
		accAddress := acc.Address.String()
		accNonce := senderNonces[accAddress]
		senderNonces[accAddress] += 1
		lastPriority := senderLastPriority[accAddress]
		txPriority := r.Intn(lastPriority)
		if txPriority == 0 {
			txPriority += 1
		}
		senderLastPriority[accAddress] = txPriority
		tx := txWithPriority{
			priority: int64(txPriority),
			tx:       simulateTx2(ctx, acc, accNonce),
			nonce:    accNonce,
			address:  accAddress,
		}
		ordered = append(ordered, tx)
	}
	for _, item := range ordered {
		tx := txWithPriority{
			priority: item.priority,
			tx:       item.tx,
			nonce:    item.nonce,
			address:  item.address,
		}
		shuffled = append(shuffled, tx)
	}
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return ordered, shuffled
}

func simulateTx2(ctx sdk.Context, acc simtypes.Account, nonce uint64) sdk.Tx {
	s := rand.NewSource(1)
	r := rand.New(s)
	txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
	msg := group.MsgUpdateGroupMembers{
		GroupId:       1,
		Admin:         acc.Address.String(),
		MemberUpdates: []group.MemberRequest{},
	}
	fees, _ := simtypes.RandomFees(r, ctx, sdk.NewCoins(sdk.NewCoin("coin", sdk.NewInt(100000000))))

	tx, _ := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{&msg},
		fees,
		simtestutil.DefaultGenTxGas,
		ctx.ChainID(),
		[]uint64{authtypes.NewBaseAccountWithAddress(acc.Address).GetAccountNumber()},
		[]uint64{nonce},
		acc.PrivKey,
	)
	return tx
}
