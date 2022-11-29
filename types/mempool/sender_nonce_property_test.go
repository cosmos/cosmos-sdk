package mempool_test

import (
	"math/rand"
	"sort"

	"pgregory.net/rapid"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mempool "github.com/cosmos/cosmos-sdk/types/mempool"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	_ sdk.Tx                  = (*testTx)(nil)
	_ signing.SigVerifiableTx = (*testTx)(nil)
	_ cryptotypes.PubKey      = (*testPubKey)(nil)
)

// Property Based Testing
// Split the senders tx in independent slices and then test the following properties in each slice
// same elements input on the mempool should be in the output except for sender nonce duplicates, which are overwritten by the later duplicate entries.
// for every sender transaction tx_n, tx_0.nonce < tx_1.nonce ... < tx_n.nonce

var genAddress = rapid.Custom(func(t *rapid.T) simtypes.Account {
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(rapid.Int64().Draw(t, "seed for account"))), 1)
	return accounts[0]
})

func testMempoolProperties(t *rapid.T) {

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	mp := mempool.NewSenderNonceMempool()

	genMultipleAddress := rapid.SliceOfNDistinct(genAddress, 1, 10, func(acc simtypes.Account) string {
		return acc.Address.String()
	})

	accounts := genMultipleAddress.Draw(t, "address")
	genTx := rapid.Custom(func(t *rapid.T) testTx {
		return testTx{
			priority: rapid.Int64Range(0, 1000).Draw(t, "priority"),
			nonce:    rapid.Uint64().Draw(t, "nonce"),
			address:  rapid.SampledFrom(accounts).Draw(t, "acc").Address,
		}
	})
	genMultipleTX := rapid.SliceOfN(genTx, 1, 500)

	txs := genMultipleTX.Draw(t, "txs")
	senderTxRaw := getSenderTxMap(txs)

	for _, tx := range txs {
		err := mp.Insert(ctx, tx)
		require.NoError(t, err)
	}

	iter := mp.Select(ctx, nil)
	orderTx := fetchAllTxs(iter)
	senderTxOrdered := getSenderTxMap(orderTx)
	for key := range senderTxOrdered {
		ordered, found := senderTxOrdered[key]
		require.True(t, found)
		raw, found := senderTxRaw[key]
		require.True(t, found)
		rawSet := mergeByNonce(raw)
		sort.Slice(rawSet, func(i, j int) bool { return rawSet[i].nonce < rawSet[j].nonce })
		require.Equal(t, rawSet, ordered)
	}
}

func (s *MempoolTestSuite) TestProperties() {
	t := s.T()
	rapid.Check(t, testMempoolProperties)
}

func getSenderTxMap(txs []testTx) map[string][]testTx {
	senderTxs := make(map[string][]testTx)
	for _, tx := range txs {
		stx, found := senderTxs[tx.address.String()]
		if !found {
			stx = make([]testTx, 0)
		}
		stx = append(stx, tx)
		senderTxs[tx.address.String()] = stx
	}
	return senderTxs
}

func fetchAllTxs(iterator mempool.Iterator) []testTx {
	var txs []testTx
	for iterator != nil {
		tx := iterator.Tx()
		txs = append(txs, tx.(testTx))
		i := iterator.Next()
		iterator = i
	}
	return txs
}

func mergeByNonce(raw []testTx) []testTx {
	rawMap := make(map[uint64]testTx)
	for _, v := range raw {
		rawMap[v.nonce] = v
	}
	result := make([]testTx, 0)
	for _, v := range rawMap {
		result = append(result, v)
	}
	return result
}
