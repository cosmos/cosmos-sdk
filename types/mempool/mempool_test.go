package mempool_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	signing2 "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

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
	hash         [32]byte
	priority     int64
	nonce        uint64
	address      sdk.AccAddress
	multiAddress []sdk.AccAddress
	multiNonces  []uint64
}

func (tx testTx) GetSigners() []sdk.AccAddress {
	if len(tx.multiAddress) == 0 {
		return []sdk.AccAddress{tx.address}
	}
	return tx.multiAddress
}

func (tx testTx) GetPubKeys() ([]cryptotypes.PubKey, error) {
	panic("GetPubkeys not implemented")
}

func (tx testTx) GetSignaturesV2() (res []signing2.SignatureV2, err error) {
	if len(tx.multiNonces) == 0 {
		res = append(res, signing2.SignatureV2{
			PubKey:   nil,
			Data:     nil,
			Sequence: tx.nonce,
		})
	} else {
		for _, nonce := range tx.multiNonces {
			res = append(res, signing2.SignatureV2{
				PubKey:   nil,
				Data:     nil,
				Sequence: nonce,
			})
		}
	}
	return res, nil
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
	return fmt.Sprintf("tx a: %s, p: %d, n: %d", tx.address, tx.priority, tx.nonce)
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

type txSpec struct {
	i     int
	h     int
	p     int
	n     int
	a     sdk.AccAddress
	multi []txSpec
}

func (tx txSpec) String() string {
	return fmt.Sprintf("[tx i: %d, a: %s, p: %d, n: %d]", tx.i, tx.a, tx.p, tx.n)
}

func TestOutOfOrder(t *testing.T) {
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 2)
	sa := accounts[0].Address
	sb := accounts[1].Address

	outOfOrders := [][]testTx{
		{
			{priority: 20, nonce: 1, address: sa},
			{priority: 21, nonce: 4, address: sa},
			{priority: 15, nonce: 1, address: sb},
			{priority: 8, nonce: 3, address: sa},
			{priority: 6, nonce: 2, address: sa},
		},
		{
			{priority: 15, nonce: 1, address: sb},
			{priority: 20, nonce: 1, address: sa},
			{priority: 21, nonce: 4, address: sa},
			{priority: 8, nonce: 3, address: sa},
			{priority: 6, nonce: 2, address: sa},
		}}

	for _, outOfOrder := range outOfOrders {
		var mtxs []mempool.Tx
		for _, mtx := range outOfOrder {
			mtxs = append(mtxs, mtx)
		}
		require.Error(t, validateOrder(mtxs))
	}

	seed := time.Now().UnixNano()
	randomTxs := genRandomTxs(seed, 1000, 10)
	var rmtxs []mempool.Tx
	for _, rtx := range randomTxs {
		rmtxs = append(rmtxs, rtx)
	}

	require.Error(t, validateOrder(rmtxs))
}

func TestTxOrder(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 5)
	sa := accounts[0].Address
	sb := accounts[1].Address
	sc := accounts[2].Address
	//sd := accounts[3].Address
	//se := accounts[4].Address

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
			order: []int{4, 3, 2, 1, 0},
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
			order: []int{3, 4, 5, 0, 1, 2},
		},
		{
			txs: []txSpec{
				{p: 21, n: 4, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 20, n: 1, a: sa},
			},
			order: []int{2, 0, 1}},
		{
			txs: []txSpec{
				{p: 50, n: 3, a: sa},
				{p: 30, n: 2, a: sa},
				{p: 10, n: 1, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 21, n: 2, a: sb},
			},
			order: []int{3, 4, 2, 1, 0},
		},
		{
			txs: []txSpec{
				{p: 50, n: 3, a: sa},
				{p: 10, n: 2, a: sa},
				{p: 99, n: 1, a: sa},
				{p: 15, n: 1, a: sb},
				{p: 8, n: 2, a: sb},
			},
			order: []int{2, 3, 1, 0, 4},
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
			order: []int{3, 2, 0, 4, 1, 5, 6},
		},
		{
			txs: []txSpec{
				{p: 30, multi: []txSpec{
					{n: 2, a: sa},
					{n: 1, a: sc}}},
				{p: 20, a: sb, n: 1},
				{p: 15, a: sa, n: 1},
				{p: 10, a: sa, n: 0},
				{p: 8, a: sb, n: 0},
				{p: 6, a: sa, n: 3},
				{p: 4, a: sb, n: 3},
				{p: 2, a: sc, n: 0},
				{p: 7, a: sc, n: 3},
			},
			order: []int{3, 2, 4, 1, 6, 7, 0, 5, 8},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			// create fresh mempool
			// TODO test both?

			//pool := mempool.NewDefaultMempool()
			pool := mempool.NewGraph()

			// create test txs and insert into mempool
			for i, ts := range tt.txs {
				var tx testTx
				if len(ts.multi) == 0 {
					tx = testTx{hash: [32]byte{byte(i)}, priority: int64(ts.p), nonce: uint64(ts.n), address: ts.a}
				} else {
					var nonces []uint64
					var addresses []sdk.AccAddress
					for _, ms := range ts.multi {
						nonces = append(nonces, uint64(ms.n))
						addresses = append(addresses, ms.a)
					}
					tx = testTx{
						hash:         [32]byte{byte(i)},
						priority:     int64(ts.p),
						multiNonces:  nonces,
						multiAddress: addresses,
					}
				}

				c := ctx.WithPriority(tx.priority)
				err := pool.Insert(c, tx)
				require.NoError(t, err)
			}

			orderedTxs, err := pool.Select(ctx, nil, 1000)
			require.NoError(t, err)
			var txOrder []int
			for _, tx := range orderedTxs {
				txOrder = append(txOrder, int(tx.(testTx).hash[0]))
			}
			require.Equal(t, tt.order, txOrder)
			require.NoError(t, validateOrder(orderedTxs))
		})
	}
}

func TestRandomTxOrderManyTimes(t *testing.T) {
	for i := 0; i < 30; i++ {
		TestRandomTxOrder(t)
		TestRandomGeneratedTx(t)
	}
}

// validateOrder checks that the txs are ordered by priority and nonce
// in O(n^2) time by checking each tx against all the other txs
func validateOrder(mtxs []mempool.Tx) error {
	var itxs []txSpec
	for i, mtx := range mtxs {
		tx := mtx.(testTx)
		itxs = append(itxs, txSpec{p: int(tx.priority), n: int(tx.nonce), a: tx.address, i: i})
	}

	// Given 2 transactions t1 and t2, where t2.p > t1.p but t2.i < t1.i
	// Then if t2.sender have the same sender then t2.nonce > t1.nonce
	// or
	// If t1 and t2 have different senders then there must be some t3 with
	// t3.sender == t2.sender and t3.n < t2.n and t3.p <= t1.p

	for _, a := range itxs {
		for _, b := range itxs {
			// when b is before a

			// when a is before b
			if a.i < b.i {
				// same sender
				if a.a.Equals(b.a) {
					// same sender
					if a.n == b.n {
						return fmt.Errorf("same sender tx have the same nonce\n%v\n%v", a, b)
					}
					if a.n > b.n {
						return fmt.Errorf("same sender tx have wrong nonce order\n%v\n%v", a, b)
					}
				} else {
					// different sender
					if a.p < b.p {
						// find a tx with same sender as b and lower nonce
						found := false
						for _, c := range itxs {
							if c.a.Equals(b.a) && c.n < b.n && c.p <= a.p {
								found = true
								break
							}
						}
						if !found {
							return fmt.Errorf("different sender tx have wrong order\n%v\n%v", b, a)
						}
					}
				}
			}
		}
	}
	return nil
}

func TestRandomGeneratedTx(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	numTx := 1000
	numAccounts := 10
	seed := time.Now().UnixNano()

	generated := genRandomTxs(seed, numTx, numAccounts)
	mp := mempool.NewDefaultMempool()

	for _, otx := range generated {
		tx := testTx{hash: otx.hash, priority: otx.priority, nonce: otx.nonce, address: otx.address}
		c := ctx.WithPriority(tx.priority)
		err := mp.Insert(c, tx)
		require.NoError(t, err)
	}

	selected, err := mp.Select(ctx, nil, 100000)
	require.Equal(t, len(generated), len(selected))
	require.NoError(t, err)

	start := time.Now()
	require.NoError(t, validateOrder(selected))
	duration := time.Since(start)

	fmt.Printf("seed: %d completed in %d iterations; validation in %dms\n",
		seed, mempool.Iterations(mp), duration.Milliseconds())
}

func TestRandomTxOrder(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	numTx := 1000
	numAccounts := 10

	seed := time.Now().UnixNano()
	// interesting failing seeds:
	// seed := int64(1663971399133628000)
	// seed := int64(1663989445512438000)
	//

	ordered, shuffled := genOrderedTxs(seed, numTx, numAccounts)
	mp := mempool.NewDefaultMempool()

	for _, otx := range shuffled {
		tx := testTx{hash: otx.hash, priority: otx.priority, nonce: otx.nonce, address: otx.address}
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
		orderedStr = fmt.Sprintf("%s\n%s, %d, %d; %d",
			orderedStr, otx.address, otx.priority, otx.nonce, otx.hash[0])
		selectedStr = fmt.Sprintf("%s\n%s, %d, %d; %d",
			selectedStr, stx.address, stx.priority, stx.nonce, stx.hash[0])
	}

	require.NoError(t, err)
	require.Equal(t, numTx, len(selected))

	errMsg := fmt.Sprintf("Expected order: %v\nGot order: %v\nSeed: %v", orderedStr, selectedStr, seed)

	//mempool.DebugPrintKeys(mp)

	require.NoError(t, validateOrder(selected), errMsg)

	/*for i, tx := range selected {
		msg := fmt.Sprintf("Failed tx at index %d\n%s", i, errMsg)
		require.Equal(t, ordered[i], tx.(testTx), msg)
		require.Equal(t, tx.(testTx).priority, ordered[i].priority, msg)
		require.Equal(t, tx.(testTx).nonce, ordered[i].nonce, msg)
		require.Equal(t, tx.(testTx).address, ordered[i].address, msg)
	}*/

	fmt.Printf("seed: %d completed in %d iterations\n", seed, mempool.Iterations(mp))
}

type txKey struct {
	sender   string
	nonce    uint64
	priority int64
	hash     [32]byte
}

func genRandomTxs(seed int64, countTx int, countAccount int) (res []testTx) {
	maxPriority := 100
	r := rand.New(rand.NewSource(seed))
	accounts := simtypes.RandomAccounts(r, countAccount)
	accountNonces := make(map[string]uint64)
	for _, account := range accounts {
		accountNonces[account.Address.String()] = 0
	}

	for i := 0; i < countTx; i++ {
		addr := accounts[r.Intn(countAccount)].Address
		priority := int64(r.Intn(maxPriority + 1))
		nonce := accountNonces[addr.String()]
		accountNonces[addr.String()] = nonce + 1
		res = append(res, testTx{
			priority: priority,
			nonce:    nonce,
			address:  addr,
			hash:     [32]byte{byte(i)}})
	}

	return res
}

// since there are multiple valid ordered graph traversals for a given set of txs strict
// validation against the ordered txs generated from this function is not possible as written
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
