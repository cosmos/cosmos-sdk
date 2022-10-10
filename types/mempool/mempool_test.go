package mempool_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
	hash     [32]byte
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

func (tx testTx) GetHash() [32]byte { return tx.hash }

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

func TestDefaultMempool(t *testing.T) {
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
	mp := mempool.NewDefaultMempool()
	for _, tx := range txs {
		ctx.WithPriority(tx.priority)
		err := mp.Insert(ctx, tx)
		require.NoError(t, err)
	}
	require.Equal(t, len(accounts), mp.CountTx())

	// distinct sender-nonce should not overwrite a tx
	mp = mempool.NewDefaultMempool()
	for i, tx := range txs {
		tx.nonce = uint64(i)
		err := mp.Insert(ctx, tx)
		require.NoError(t, err)
	}
	require.Equal(t, txCount, mp.CountTx())

	sel, err := mp.Select(nil, 13)
	require.NoError(t, err)
	require.Equal(t, 13, len(sel))

	// a tx which does not implement SigVerifiableTx should not be inserted
	tx := &sigErrTx{getSigs: func() ([]txsigning.SignatureV2, error) {
		return nil, fmt.Errorf("error")
	}}
	require.Error(t, mp.Insert(ctx, tx))
	require.Error(t, mp.Remove(tx))
	tx.getSigs = func() ([]txsigning.SignatureV2, error) {
		return nil, nil
	}
	require.Error(t, mp.Insert(ctx, tx))
	require.Error(t, mp.Remove(tx))

	// removing a tx not in the mempool should error
	mp = mempool.NewDefaultMempool()
	require.NoError(t, mp.Insert(ctx, txs[0]))
	require.ErrorIs(t, mp.Remove(txs[1]), mempool.ErrTxNotFound)
	mempool.DebugPrintKeys(mp)

	// inserting a tx with a different priority should overwrite the old tx
	newPriorityTx := testTx{
		address:  txs[0].address,
		priority: txs[0].priority + 1,
		nonce:    txs[0].nonce,
	}
	require.NoError(t, mp.Insert(ctx, newPriorityTx))
	require.Equal(t, 1, mp.CountTx())
}

type txSpec struct {
	i int
	h int
	p int
	n int
	a sdk.AccAddress
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
		err := validateOrder(mtxs)
		require.Error(t, err)
	}

	seed := time.Now().UnixNano()
	randomTxs := genRandomTxs(seed, 1000, 10)
	var rmtxs []mempool.Tx
	for _, rtx := range randomTxs {
		rmtxs = append(rmtxs, rtx)
	}

	require.Error(t, validateOrder(rmtxs))

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
			order: []int{3, 2, 0, 4, 1, 5, 6, 7, 8},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			pool := s.mempool

			// create test txs and insert into mempool
			for i, ts := range tt.txs {
				tx := testTx{hash: [32]byte{byte(i)}, priority: int64(ts.p), nonce: uint64(ts.n), address: ts.a}
				c := ctx.WithPriority(tx.priority)
				err := pool.Insert(c, tx)
				require.NoError(t, err)
			}

			orderedTxs, err := pool.Select(nil, 1000)
			require.NoError(t, err)
			var txOrder []int
			for _, tx := range orderedTxs {
				txOrder = append(txOrder, int(tx.(testTx).hash[0]))
			}
			require.Equal(t, tt.order, txOrder)
			require.NoError(t, validateOrder(orderedTxs))

			for _, tx := range orderedTxs {
				require.NoError(t, pool.Remove(tx))
			}

			require.Equal(t, 0, pool.CountTx())
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
	s.mempool = mempool.NewDefaultMempool(mempool.WithOnRead(func(tx mempool.Tx) {
		s.iterations++
	}))
}

func (s *MempoolTestSuite) SetupTest() {
	s.numTxs = 1000
	s.numAccounts = 100
	s.resetMempool()
}

func TestMempoolTestSuite(t *testing.T) {
	suite.Run(t, new(MempoolTestSuite))
}

func (s *MempoolTestSuite) TestRandomTxOrderManyTimes() {
	for i := 0; i < 3; i++ {
		s.Run("TestRandomGeneratedTxs", func() {
			s.TestRandomGeneratedTxs()
		})
		s.resetMempool()
		s.Run("TestRandomWalkTxs", func() {
			s.TestRandomWalkTxs()
		})
		s.resetMempool()
	}
}

// validateOrder checks that the txs are ordered by priority and nonce
// in O(n^2) time by checking each tx against all the other txs
func validateOrder(mtxs []mempool.Tx) error {
	iterations := 0
	var itxs []txSpec
	for i, mtx := range mtxs {
		iterations++
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
			iterations++
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
							iterations++
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
	// fmt.Printf("validation in iterations: %d\n", iterations)
	return nil
}

func (s *MempoolTestSuite) TestRandomGeneratedTxs() {
	t := s.T()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	seed := time.Now().UnixNano()

	generated := genRandomTxs(seed, s.numTxs, s.numAccounts)
	mp := s.mempool

	for _, otx := range generated {
		tx := testTx{hash: otx.hash, priority: otx.priority, nonce: otx.nonce, address: otx.address}
		c := ctx.WithPriority(tx.priority)
		err := mp.Insert(c, tx)
		require.NoError(t, err)
	}

	selected, err := mp.Select(nil, 100000)
	require.Equal(t, len(generated), len(selected))
	require.NoError(t, err)

	start := time.Now()
	require.NoError(t, validateOrder(selected))
	duration := time.Since(start)

	fmt.Printf("seed: %d completed in %d iterations; validation in %dms\n",
		seed, s.iterations, duration.Milliseconds())
}

func (s *MempoolTestSuite) TestRandomWalkTxs() {
	t := s.T()
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	seed := time.Now().UnixNano()
	// interesting failing seeds:
	// seed := int64(1663971399133628000)
	// seed := int64(1663989445512438000)
	//

	ordered, shuffled := genOrderedTxs(seed, s.numTxs, s.numAccounts)
	mp := s.mempool

	for _, otx := range shuffled {
		tx := testTx{hash: otx.hash, priority: otx.priority, nonce: otx.nonce, address: otx.address}
		c := ctx.WithPriority(tx.priority)
		err := mp.Insert(c, tx)
		require.NoError(t, err)
	}

	require.Equal(t, s.numTxs, mp.CountTx())

	selected, err := mp.Select(nil, math.MaxInt)
	require.Equal(t, len(ordered), len(selected))
	var orderedStr, selectedStr string

	for i := 0; i < s.numTxs; i++ {
		otx := ordered[i]
		stx := selected[i].(testTx)
		orderedStr = fmt.Sprintf("%s\n%s, %d, %d; %d",
			orderedStr, otx.address, otx.priority, otx.nonce, otx.hash[0])
		selectedStr = fmt.Sprintf("%s\n%s, %d, %d; %d",
			selectedStr, stx.address, stx.priority, stx.nonce, stx.hash[0])
	}

	require.NoError(t, err)
	require.Equal(t, s.numTxs, len(selected))

	errMsg := fmt.Sprintf("Expected order: %v\nGot order: %v\nSeed: %v", orderedStr, selectedStr, seed)

	start := time.Now()
	require.NoError(t, validateOrder(selected), errMsg)
	duration := time.Since(start)

	fmt.Printf("seed: %d completed in %d iterations; validation in %dms\n",
		seed, s.iterations, duration.Milliseconds())
}

func (s *MempoolTestSuite) TestSampleTxs() {
	ctxt := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	t := s.T()
	s.resetMempool()
	mp := s.mempool
	delegatorTx, err := unmarshalTx(msgWithdrawDelegatorReward)

	require.NoError(t, err)
	require.NoError(t, mp.Insert(ctxt, delegatorTx.(mempool.Tx)))
	require.Equal(t, 1, mp.CountTx())

	proposalTx, err := unmarshalTx(msgMultiSigMsgSubmitProposal)
	require.NoError(t, err)
	require.NoError(t, mp.Insert(ctxt, proposalTx.(mempool.Tx)))
	require.Equal(t, 2, mp.CountTx())
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

func unmarshalTx(txBytes []byte) (sdk.Tx, error) {
	cfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{}, gov.AppModuleBasic{})
	cfg.InterfaceRegistry.RegisterImplementations((*govtypes.Content)(nil), &disttypes.CommunityPoolSpendProposal{})
	return cfg.TxConfig.TxJSONDecoder()(txBytes)
}

var msgWithdrawDelegatorReward = []byte("{\"body\":{\"messages\":[{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qxerarrl\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1k2d9ed9vgfuk2m58a2d80q9u6qljkh4vfaqjfq\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1vygmh344ldv9qefss9ek7ggsnxparljlmj56q5\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5wxxp9gd\"}],\"memo\":\"\",\"timeout_height\":\"0\",\"extension_options\":[],\"non_critical_extension_options\":[]},\"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"AmbXAy10a0SerEefTYQzqyGQdX5kiTEWJZ1PZKX1oswX\"},\"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}},\"sequence\":\"119\"}],\"fee\":{\"amount\":[{\"denom\":\"uatom\",\"amount\":\"15968\"}],\"gas_limit\":\"638717\",\"payer\":\"\",\"granter\":\"\"}},\"signatures\":[\"ji+inUo4xGlN9piRQLdLCeJWa7irwnqzrMVPcmzJyG5y6NPc+ZuNaIc3uvk5NLDJytRB8AHX0GqNETR\\/Q8fz4Q==\"]}")
var msgMultiSigMsgSubmitProposal = []byte("{\"body\":{\"messages\":[{\"@type\":\"\\/cosmos.gov.v1beta1.MsgSubmitProposal\",\"content\":{\"@type\":\"\\/cosmos.distribution.v1beta1.CommunityPoolSpendProposal\",\"title\":\"ATOM \\ud83e\\udd1d Osmosis:  Allocate Community Pool to ATOM Liquidity Incentives\",\"description\":\"ATOMs should be the base money of Cosmos, just like ETH is the base money of the entire Ethereum DeFi ecosystem. ATOM is currently well positioned to play this role among Cosmos assets because it has the highest market cap, most liquidity, largest brand, and many integrations with fiat onramps. ATOM is the gateway to Cosmos.\\n\\nIn the Cosmos Hub Port City vision, ATOMs are pitched as equity in the Cosmos Hub.  However, this alone is insufficient to establish ATOM as the base currency of the Cosmos ecosystem as a whole. Instead, the ATOM community must work to actively promote the use of ATOMs throughout the Cosmos ecosystem, rather than passively relying on the Hub's reputation to create ATOM's value.\\n\\nIn order to cement the role of ATOMs in Cosmos DeFi, the Cosmos Hub should leverage its community pool to help align incentives with other protocols within the Cosmos ecosystem. We propose beginning this initiative by using the community pool ATOMs to incentivize deep ATOM base pair liquidity pools on the Osmosis Network.\\n\\nOsmosis is the first IBC-enabled DeFi application. Within its 3 weeks of existence, it has already 100x\\u2019d the number of IBC transactions ever created, demonstrating the power of IBC and the ability of the Cosmos SDK to bootstrap DeFi protocols with $100M+ TVL in a short period of time. Since its announcement Osmosis has helped bring renewed attention and interest to Cosmos from the crypto community at large and kickstarted the era of Cosmos DeFi.\\n\\nOsmosis has already helped in establishing ATOM as the Schelling Point of the Cosmos ecosystem.  The genesis distribution of OSMO was primarily based on an airdrop to ATOM holders specifically, acknowledging the importance of ATOM to all future projects within the Cosmos. Furthermore, the Osmosis LP rewards currently incentivize ATOMs to be one of the main base pairs of the platform.\\n\\nOsmosis has the ability to incentivize AMM liquidity, a feature not available on any other IBC-enabled DEX. Osmosis already uses its own native OSMO liquidity rewards to incentivize ATOMs to be one of the main base pairs, leading to ~2.2 million ATOMs already providing liquidity on the platform.\\n\\nIn addition to these native OSMO LP Rewards, the platform also includes a feature called \\u201cexternal incentives\\u201d that allows anyone to permissionlessly add additional incentives in any token to the LPs of any AMM pools they wish. You can read more about this mechanism here: https:\\/\\/medium.com\\/osmosis\\/osmosis-liquidity-mining-101-2fa58d0e9d4d#f413 . Pools containing Cosmos assets such as AKT and XPRT are already planned to receive incentives from their respective community pools and\\/or foundations.\\n\\nWe propose the Cosmos Hub dedicate 100,000 ATOMs from its Community Pool to be allocated towards liquidity incentives on Osmosis over the next 3 months. This community fund proposal will transfer 100,000 ATOMs to a multisig group who will then allocate the ATOMs to bonded liquidity gauges on Osmosis on a biweekly basis, according to direction given by Cosmos Hub governance.  For simplicity, we propose setting the liquidity incentives to initially point to Osmosis Pool #1, the ATOM\\/OSMO pool, which is the pool with by far the highest TVL and Volume. Cosmos Hub governance can then use Text Proposals to further direct the multisig members to reallocate incentives to new pools.\\n\\nThe multisig will consist of a 2\\/3 key holder set consisting of the following individuals whom have all agreed to participate in this process shall this proposal pass:\\n\\n- Zaki Manian\\n- Federico Kunze\\n- Marko Baricevic\\n\\nThis is one small step for the Hub, but one giant leap for ATOM-aligned.\\n\",\"recipient\":\"cosmos157n0d38vwn5dvh64rc39q3lyqez0a689g45rkc\",\"amount\":[{\"denom\":\"uatom\",\"amount\":\"100000000000\"}]},\"initial_deposit\":[{\"denom\":\"uatom\",\"amount\":\"64000000\"}],\"proposer\":\"cosmos1ey69r37gfxvxg62sh4r0ktpuc46pzjrmz29g45\"}],\"memo\":\"\",\"timeout_height\":\"0\",\"extension_options\":[],\"non_critical_extension_options\":[]},\"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"\\/cosmos.crypto.multisig.LegacyAminoPubKey\",\"threshold\":2,\"public_keys\":[{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"AldOvgv8dU9ZZzuhGydQD5FYreLhfhoBgrDKi8ZSTbCQ\"},{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"AxUMR\\/GKoycWplR+2otzaQZ9zhHRQWJFt3h1bPg1ltha\"},{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"AlI9yVj2Aejow6bYl2nTRylfU+9LjQLEl3keq0sERx9+\"},{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A0UvHPcvCCaIoFY9Ygh0Pxq9SZTAWtduOyinit\\/8uo+Q\"},{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"As7R9fDUnwsUVLDr1cxspp+cY9UfXfUf7i9\\/w+N0EzKA\"}]},\"mode_info\":{\"multi\":{\"bitarray\":{\"extra_bits_stored\":5,\"elems\":\"SA==\"},\"mode_infos\":[{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}},{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}}]}},\"sequence\":\"102\"}],\"fee\":{\"amount\":[],\"gas_limit\":\"10000000\",\"payer\":\"\",\"granter\":\"\"}},\"signatures\":[\"CkB\\/KKWTFntEWbg1A0vu7DCHffJ4x4db\\/EI8dIVzRFFW7iuZBzvq+jYBtrcTlVpEVfmCY3ggIMnWfbMbb1egIlYbCkAmDf6Eaj1NbyXY8JZZtYAX3Qj81ZuKZUBeLW1ZvH1XqAg9sl\\/sqpLMnsJzKfmqEXvhoMwu1YxcSzrY6CJfuYL6\"]}")
