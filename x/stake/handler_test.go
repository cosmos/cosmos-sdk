package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/rational"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/state"
)

//______________________________________________________________________

// dummy transfer functions, represents store operations on account balances

type testCoinSender struct {
	store map[string]int64
}

var _ coinSend = testCoinSender{} // enforce interface at compile time

func (c testCoinSender) transferFn(sender, receiver sdk.Actor, coins coin.Coins) error {
	c.store[string(sender.Address)] -= coins[0].Amount
	c.store[string(receiver.Address)] += coins[0].Amount
	return nil
}

//______________________________________________________________________

func initAccounts(n int, amount int64) ([]sdk.Actor, map[string]int64) {
	accStore := map[string]int64{}
	senders := newActors(n)
	for _, sender := range senders {
		accStore[string(sender.Address)] = amount
	}
	return senders, accStore
}

func newTxDeclareCandidacy(amt int64, pubKey crypto.PubKey) TxDeclareCandidacy {
	return TxDeclareCandidacy{
		BondUpdate{
			PubKey: pubKey,
			Bond:   coin.Coin{"fermion", amt},
		},
		Description{},
	}
}

func newTxDelegate(amt int64, pubKey crypto.PubKey) TxDelegate {
	return TxDelegate{BondUpdate{
		PubKey: pubKey,
		Bond:   coin.Coin{"fermion", amt},
	}}
}

func newTxUnbond(shares string, pubKey crypto.PubKey) TxUnbond {
	return TxUnbond{
		PubKey: pubKey,
		Shares: shares,
	}
}

func paramsNoInflation() Params {
	return Params{
		HoldBonded:          sdk.NewActor(stakingModuleName, []byte("77777777777777777777777777777777")),
		HoldUnbonded:        sdk.NewActor(stakingModuleName, []byte("88888888888888888888888888888888")),
		InflationRateChange: rational.Zero,
		InflationMax:        rational.Zero,
		InflationMin:        rational.Zero,
		GoalBonded:          rational.New(67, 100),
		MaxVals:             100,
		AllowedBondDenom:    "fermion",
		GasDeclareCandidacy: 20,
		GasEditCandidacy:    20,
		GasDelegate:         20,
		GasUnbond:           20,
	}
}

func newDeliver(sender sdk.Actor, accStore map[string]int64) deliver {
	store := state.NewMemKVStore()
	params := paramsNoInflation()
	saveParams(store, params)
	return deliver{
		store:    store,
		sender:   sender,
		params:   params,
		gs:       loadGlobalState(store),
		transfer: testCoinSender{accStore}.transferFn,
	}
}

func TestDuplicatesTxDeclareCandidacy(t *testing.T) {
	assert := assert.New(t)
	senders, accStore := initAccounts(2, 1000) // for accounts

	deliverer := newDeliver(senders[0], accStore)
	checker := check{
		store:  deliverer.store,
		sender: senders[0],
	}

	txDeclareCandidacy := newTxDeclareCandidacy(10, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(got, "expected no error on runTxDeclareCandidacy")

	// one sender can bond to two different pubKeys
	txDeclareCandidacy.PubKey = pks[1]
	err := checker.declareCandidacy(txDeclareCandidacy)
	assert.Nil(err, "didn't expected error on checkTx")

	// two senders cant bond to the same pubkey
	checker.sender = senders[1]
	txDeclareCandidacy.PubKey = pks[0]
	err = checker.declareCandidacy(txDeclareCandidacy)
	assert.NotNil(err, "expected error on checkTx")
}

func TestIncrementsTxDelegate(t *testing.T) {
	assert := assert.New(t)
	initSender := int64(1000)
	senders, accStore := initAccounts(1, initSender) // for accounts
	deliverer := newDeliver(senders[0], accStore)

	// first declare candidacy
	bondAmount := int64(10)
	txDeclareCandidacy := newTxDeclareCandidacy(bondAmount, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(got, "expected declare candidacy tx to be ok, got %v", got)
	expectedBond := bondAmount // 1 since we send 1 at the start of loop,

	// just send the same txbond multiple times
	holder := deliverer.params.HoldUnbonded // XXX this should be HoldBonded, new SDK updates
	txDelegate := newTxDelegate(bondAmount, pks[0])
	for i := 0; i < 5; i++ {
		got := deliverer.delegate(txDelegate)
		assert.NoError(got, "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := loadCandidates(deliverer.store)
		expectedBond += bondAmount
		expectedSender := initSender - expectedBond
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotHolder := accStore[string(holder.Address)]
		gotSender := accStore[string(deliverer.sender.Address)]
		assert.Equal(expectedBond, gotBonded, "i: %v, %v, %v", i, expectedBond, gotBonded)
		assert.Equal(expectedBond, gotHolder, "i: %v, %v, %v", i, expectedBond, gotHolder)
		assert.Equal(expectedSender, gotSender, "i: %v, %v, %v", i, expectedSender, gotSender)
	}
}

func TestIncrementsTxUnbond(t *testing.T) {
	assert := assert.New(t)
	initSender := int64(0)
	senders, accStore := initAccounts(1, initSender) // for accounts
	deliverer := newDeliver(senders[0], accStore)

	// set initial bond
	initBond := int64(1000)
	accStore[string(deliverer.sender.Address)] = initBond
	got := deliverer.declareCandidacy(newTxDeclareCandidacy(initBond, pks[0]))
	assert.NoError(got, "expected initial bond tx to be ok, got %v", got)

	// just send the same txunbond multiple times
	holder := deliverer.params.HoldUnbonded // XXX new SDK, this should be HoldBonded

	// XXX use decimals here
	unbondShares, unbondSharesStr := int64(10), "10"
	txUndelegate := newTxUnbond(unbondSharesStr, pks[0])
	nUnbonds := 5
	for i := 0; i < nUnbonds; i++ {
		got := deliverer.unbond(txUndelegate)
		assert.NoError(got, "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := loadCandidates(deliverer.store)
		expectedBond := initBond - int64(i+1)*unbondShares // +1 since we send 1 at the start of loop
		expectedSender := initSender + (initBond - expectedBond)
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotHolder := accStore[string(holder.Address)]
		gotSender := accStore[string(deliverer.sender.Address)]

		assert.Equal(expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
		assert.Equal(expectedBond, gotHolder, "%v, %v", expectedBond, gotHolder)
		assert.Equal(expectedSender, gotSender, "%v, %v", expectedSender, gotSender)
	}

	// these are more than we have bonded now
	errorCases := []int64{
		//1<<64 - 1, // more than int64
		//1<<63 + 1, // more than int64
		1<<63 - 1,
		1 << 31,
		initBond,
	}
	for _, c := range errorCases {
		unbondShares := strconv.Itoa(int(c))
		txUndelegate := newTxUnbond(unbondShares, pks[0])
		got = deliverer.unbond(txUndelegate)
		assert.Error(got, "expected unbond tx to fail")
	}

	leftBonded := initBond - unbondShares*int64(nUnbonds)

	// should be unable to unbond one more than we have
	txUndelegate = newTxUnbond(strconv.Itoa(int(leftBonded)+1), pks[0])
	got = deliverer.unbond(txUndelegate)
	assert.Error(got, "expected unbond tx to fail")

	// should be able to unbond just what we have
	txUndelegate = newTxUnbond(strconv.Itoa(int(leftBonded)), pks[0])
	got = deliverer.unbond(txUndelegate)
	assert.NoError(got, "expected unbond tx to pass")
}

func TestMultipleTxDeclareCandidacy(t *testing.T) {
	assert := assert.New(t)
	initSender := int64(1000)
	senders, accStore := initAccounts(3, initSender)
	pubKeys := []crypto.PubKey{pks[0], pks[1], pks[2]}
	deliverer := newDeliver(senders[0], accStore)

	// bond them all
	for i, sender := range senders {
		txDeclareCandidacy := newTxDeclareCandidacy(10, pubKeys[i])
		deliverer.sender = sender
		got := deliverer.declareCandidacy(txDeclareCandidacy)
		assert.NoError(got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := loadCandidates(deliverer.store)
		val := candidates[i]
		balanceGot, balanceExpd := accStore[string(val.Owner.Address)], initSender-10
		assert.Equal(i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		assert.Equal(10, int(val.Liabilities.Evaluate()), "expected %d shares, got %d", 10, val.Liabilities)
		assert.Equal(balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, sender := range senders {
		candidatePre := loadCandidate(deliverer.store, pubKeys[i])
		txUndelegate := newTxUnbond("10", pubKeys[i])
		deliverer.sender = sender
		got := deliverer.unbond(txUndelegate)
		assert.NoError(got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := loadCandidates(deliverer.store)
		assert.Equal(len(senders)-(i+1), len(candidates), "expected %d candidates got %d", len(senders)-(i+1), len(candidates))

		candidatePost := loadCandidate(deliverer.store, pubKeys[i])
		balanceGot, balanceExpd := accStore[string(candidatePre.Owner.Address)], initSender
		assert.Nil(candidatePost, "expected nil candidate retrieve, got %d", 0, candidatePost)
		assert.Equal(balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}
}

func TestMultipleTxDelegate(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	accounts, accStore := initAccounts(3, 1000)
	sender, delegators := accounts[0], accounts[1:]
	deliverer := newDeliver(sender, accStore)

	//first make a candidate
	txDeclareCandidacy := newTxDeclareCandidacy(10, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	require.NoError(got, "expected tx to be ok, got %v", got)

	// delegate multiple parties
	for i, delegator := range delegators {
		txDelegate := newTxDelegate(10, pks[0])
		deliverer.sender = delegator
		got := deliverer.delegate(txDelegate)
		require.NoError(got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond := loadDelegatorBond(deliverer.store, delegator, pks[0])
		assert.NotNil(bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegator := range delegators {
		txUndelegate := newTxUnbond("10", pks[0])
		deliverer.sender = delegator
		got := deliverer.unbond(txUndelegate)
		require.NoError(got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		bond := loadDelegatorBond(deliverer.store, delegator, pks[0])
		assert.Nil(bond, "expected delegatee bond %d to be nil", bond)
	}
}

func TestVoidCandidacy(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	accounts, accStore := initAccounts(2, 1000) // for accounts
	sender, delegator := accounts[0], accounts[1]
	deliverer := newDeliver(sender, accStore)

	// create the candidate
	txDeclareCandidacy := newTxDeclareCandidacy(10, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	require.NoError(got, "expected no error on runTxDeclareCandidacy")

	// bond a delegator
	txDelegate := newTxDelegate(10, pks[0])
	deliverer.sender = delegator
	got = deliverer.delegate(txDelegate)
	require.NoError(got, "expected ok, got %v", got)

	// unbond the candidates bond portion
	txUndelegate := newTxUnbond("10", pks[0])
	deliverer.sender = sender
	got = deliverer.unbond(txUndelegate)
	require.NoError(got, "expected no error on runTxDeclareCandidacy")

	// test that this pubkey cannot yet be bonded too
	deliverer.sender = delegator
	got = deliverer.delegate(txDelegate)
	assert.Error(got, "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	got = deliverer.unbond(txUndelegate)
	require.NoError(got, "expected no error on runTxDeclareCandidacy")

	// verify that the pubkey can now be reused
	got = deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(got, "expected ok, got %v", got)

}
