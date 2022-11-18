package mempool_test

import (
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mempool2 "github.com/cosmos/cosmos-sdk/types/mempool"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"math/rand"
	"pgregory.net/rapid"
)

var (
	_ sdk.Tx                  = (*testTx)(nil)
	_ signing.SigVerifiableTx = (*testTx)(nil)
	_ cryptotypes.PubKey      = (*testPubKey)(nil)
)

// Property Based Testing
// Split the senders tx in independent slices and then test the following properties in each slice
// same elements input output
// the reverse of the reverse of the list is the same
// for every sequence element pair a, b a < b

func testMempoolProperties(t *rapid.T) {

	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())
	mempool := mempool2.NewSenderNonceMempool()
	genAddress := rapid.Custom(func(t *rapid.T) simtypes.Account {
		accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(rapid.Int64().Draw(t, "seed for account"))), 1)
		return accounts[0]
	})
	genMultipleAddress := rapid.SliceOfDistinct(genAddress, func(acc simtypes.Account) string {
		return acc.Address.String()
	})

	accounts := genMultipleAddress.Draw(t, "address")
	fmt.Println(accounts)
	genTx := rapid.Custom(func(t *rapid.T) testTx {
		return testTx{
			priority: rapid.Int64Range(0, 1000).Draw(t, "priority"),
			nonce:    rapid.Uint64().Draw(t, "nonce"),
			address:  rapid.SampledFrom(accounts).Draw(t, "acc").Address,
		}
		//fmt.Println("genTX:", tx)
		//return tx

	})
	fmt.Println(genTx)
	genMultipleTX := rapid.SliceOf(genTx)

	txs := genMultipleTX.Draw(t, "txs")
	fmt.Println("txs:", txs)

	for _, tx := range txs {
		fmt.Println("tx", tx)
		fmt.Println(ctx)
		fmt.Println(mempool)
		//err := mempool.Insert(ctx, tx)
		//require.NoError(t, err)
	}

	test := rapid.SliceOf(rapid.Int().AsAny())
	for i := 0; i < 5; i++ {
		fmt.Println(i)
		fmt.Println(test.Example(i))
		fmt.Println(genMultipleAddress.Example(i))
		fmt.Println(genMultipleTX.Example(i))

	}

	require.True(t, false)

}

func (s *MempoolTestSuite) TestProperties() {
	t := s.T()
	rapid.Check(t, testMempoolProperties)
}
