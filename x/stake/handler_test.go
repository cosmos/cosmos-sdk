package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	coin "github.com/cosmos/cosmos-sdk/x/bank" // XXX fix
)

//______________________________________________________________________

func initAccounts(n int, amount int64) ([]sdk.Address, map[string]int64) {
	accStore := map[string]int64{}
	senders := newAddrs(n)
	for _, sender := range senders {
		accStore[string(sender.Address)] = amount
	}
	return senders, accStore
}

func newTestMsgDeclareCandidacy(amt int64, pubKey crypto.PubKey, address sdk.Address) MsgDeclareCandidacy {
	return MsgDeclareCandidacy{
		MsgAddr: NewMsgAddr(address),
		PubKey:  pubKey,
		Bond:    coin.Coin{"fermion", amt},
		Description{},
	}
}

func newTestMsgDelegate(amt int64, address sdk.Address) MsgDelegate {
	return MsgDelegate{
		MsgAddr: NewMsgAddr(address),
		Bond:    coin.Coin{"fermion", amt},
	}
}

func newMsgUnbond(shares string, pubKey crypto.PubKey) MsgUnbond {
	return MsgUnbond{
		PubKey: pubKey,
		Shares: shares,
	}
}

func paramsNoInflation() Params {
	return Params{
		InflationRateChange: sdk.ZeroRat,
		InflationMax:        sdk.ZeroRat,
		InflationMin:        sdk.ZeroRat,
		GoalBonded:          sdk.New(67, 100),
		MaxVals:             100,
		BondDenom:           "fermion",
		GasDeclareCandidacy: 20,
		GasEditCandidacy:    20,
		GasDelegate:         20,
		GasUnbond:           20,
	}
}

func newTestTransact(t, sender sdk.Address, isCheckTx bool) transact {
	store, mapper, coinKeeper := createTestInput(t, isCheckTx)
	params := paramsNoInflation()
	mapper.saveParams(params)
	newTransact(ctx, sender, mapper, coinKeeper)
}

func TestDuplicatesMsgDeclareCandidacy(t *testing.T) {
	senders, accStore := initAccounts(2, 1000) // for accounts

	deliverer := newDeliver(t, senders[0], accStore)
	checker := check{
		store:  deliverer.store,
		sender: senders[0],
	}

	txDeclareCandidacy := newTestMsgDeclareCandidacy(10, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// one sender can bond to two different pubKeys
	txDeclareCandidacy.PubKey = pks[1]
	err := checker.declareCandidacy(txDeclareCandidacy)
	assert.Nil(t, err, "didn't expected error on checkTx")

	// two senders cant bond to the same pubkey
	checker.sender = senders[1]
	txDeclareCandidacy.PubKey = pks[0]
	err = checker.declareCandidacy(txDeclareCandidacy)
	assert.NotNil(t, err, "expected error on checkTx")
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initSender := int64(1000)
	senders, accStore := initAccounts(1, initSender) // for accounts
	deliverer := newDeliver(t, senders[0], accStore)

	// first declare candidacy
	bondAmount := int64(10)
	txDeclareCandidacy := newTestMsgDeclareCandidacy(bondAmount, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(t, got, "expected declare candidacy tx to be ok, got %v", got)
	expectedBond := bondAmount // 1 since we send 1 at the start of loop,

	// just send the same txbond multiple times
	holder := deliverer.params.HoldUnbonded // XXX this should be HoldBonded, new SDK updates
	txDelegate := newTestMsgDelegate(bondAmount, pks[0])
	for i := 0; i < 5; i++ {
		got := deliverer.delegate(txDelegate)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := loadCandidates(deliverer.store)
		expectedBond += bondAmount
		expectedSender := initSender - expectedBond
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotHolder := accStore[string(holder.Address)]
		gotSender := accStore[string(deliverer.sender.Address)]
		assert.Equal(t, expectedBond, gotBonded, "i: %v, %v, %v", i, expectedBond, gotBonded)
		assert.Equal(t, expectedBond, gotHolder, "i: %v, %v, %v", i, expectedBond, gotHolder)
		assert.Equal(t, expectedSender, gotSender, "i: %v, %v, %v", i, expectedSender, gotSender)
	}
}

func TestIncrementsMsgUnbond(t *testing.T) {
	initSender := int64(0)
	senders, accStore := initAccounts(1, initSender) // for accounts
	deliverer := newDeliver(t, senders[0], accStore)

	// set initial bond
	initBond := int64(1000)
	accStore[string(deliverer.sender.Address)] = initBond
	got := deliverer.declareCandidacy(newTestMsgDeclareCandidacy(initBond, pks[0]))
	assert.NoError(t, got, "expected initial bond tx to be ok, got %v", got)

	// just send the same txunbond multiple times
	holder := deliverer.params.HoldUnbonded // XXX new SDK, this should be HoldBonded

	// XXX use decimals here
	unbondShares, unbondSharesStr := int64(10), "10"
	txUndelegate := newMsgUnbond(unbondSharesStr, pks[0])
	nUnbonds := 5
	for i := 0; i < nUnbonds; i++ {
		got := deliverer.unbond(txUndelegate)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := loadCandidates(deliverer.store)
		expectedBond := initBond - int64(i+1)*unbondShares // +1 since we send 1 at the start of loop
		expectedSender := initSender + (initBond - expectedBond)
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotHolder := accStore[string(holder.Address)]
		gotSender := accStore[string(deliverer.sender.Address)]

		assert.Equal(t, expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
		assert.Equal(t, expectedBond, gotHolder, "%v, %v", expectedBond, gotHolder)
		assert.Equal(t, expectedSender, gotSender, "%v, %v", expectedSender, gotSender)
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
		txUndelegate := newMsgUnbond(unbondShares, pks[0])
		got = deliverer.unbond(txUndelegate)
		assert.Error(t, got, "expected unbond tx to fail")
	}

	leftBonded := initBond - unbondShares*int64(nUnbonds)

	// should be unable to unbond one more than we have
	txUndelegate = newMsgUnbond(strconv.Itoa(int(leftBonded)+1), pks[0])
	got = deliverer.unbond(txUndelegate)
	assert.Error(t, got, "expected unbond tx to fail")

	// should be able to unbond just what we have
	txUndelegate = newMsgUnbond(strconv.Itoa(int(leftBonded)), pks[0])
	got = deliverer.unbond(txUndelegate)
	assert.NoError(t, got, "expected unbond tx to pass")
}

func TestMultipleMsgDeclareCandidacy(t *testing.T) {
	initSender := int64(1000)
	senders, accStore := initAccounts(3, initSender)
	pubKeys := []crypto.PubKey{pks[0], pks[1], pks[2]}
	deliverer := newDeliver(t, senders[0], accStore)

	// bond them all
	for i, sender := range senders {
		txDeclareCandidacy := newTestMsgDeclareCandidacy(10, pubKeys[i])
		deliverer.sender = sender
		got := deliverer.declareCandidacy(txDeclareCandidacy)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := loadCandidates(deliverer.store)
		val := candidates[i]
		balanceGot, balanceExpd := accStore[string(val.Owner.Address)], initSender-10
		assert.Equal(t, i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		assert.Equal(t, 10, int(val.Liabilities.Evaluate()), "expected %d shares, got %d", 10, val.Liabilities)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, sender := range senders {
		candidatePre := loadCandidate(deliverer.store, pubKeys[i])
		txUndelegate := newMsgUnbond("10", pubKeys[i])
		deliverer.sender = sender
		got := deliverer.unbond(txUndelegate)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := loadCandidates(deliverer.store)
		assert.Equal(t, len(senders)-(i+1), len(candidates), "expected %d candidates got %d", len(senders)-(i+1), len(candidates))

		candidatePost := loadCandidate(deliverer.store, pubKeys[i])
		balanceGot, balanceExpd := accStore[string(candidatePre.Owner.Address)], initSender
		assert.Nil(t, candidatePost, "expected nil candidate retrieve, got %d", 0, candidatePost)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	accounts, accStore := initAccounts(3, 1000)
	sender, delegators := accounts[0], accounts[1:]
	deliverer := newDeliver(t, sender, accStore)

	//first make a candidate
	txDeclareCandidacy := newTestMsgDeclareCandidacy(10, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	require.NoError(t, got, "expected tx to be ok, got %v", got)

	// delegate multiple parties
	for i, delegator := range delegators {
		txDelegate := newTestMsgDelegate(10, pks[0])
		deliverer.sender = delegator
		got := deliverer.delegate(txDelegate)
		require.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond := loadDelegatorBond(deliverer.store, delegator, pks[0])
		assert.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegator := range delegators {
		txUndelegate := newMsgUnbond("10", pks[0])
		deliverer.sender = delegator
		got := deliverer.unbond(txUndelegate)
		require.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		bond := loadDelegatorBond(deliverer.store, delegator, pks[0])
		assert.Nil(t, bond, "expected delegatee bond %d to be nil", bond)
	}
}

func TestVoidCandidacy(t *testing.T) {
	accounts, accStore := initAccounts(2, 1000) // for accounts
	sender, delegator := accounts[0], accounts[1]
	deliverer := newDeliver(t, sender, accStore)

	// create the candidate
	txDeclareCandidacy := newTestMsgDeclareCandidacy(10, pks[0])
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// bond a delegator
	txDelegate := newTestMsgDelegate(10, pks[0])
	deliverer.sender = delegator
	got = deliverer.delegate(txDelegate)
	require.NoError(t, got, "expected ok, got %v", got)

	// unbond the candidates bond portion
	txUndelegate := newMsgUnbond("10", pks[0])
	deliverer.sender = sender
	got = deliverer.unbond(txUndelegate)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// test that this pubkey cannot yet be bonded too
	deliverer.sender = delegator
	got = deliverer.delegate(txDelegate)
	assert.Error(t, got, "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	got = deliverer.unbond(txUndelegate)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// verify that the pubkey can now be reused
	got = deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(t, got, "expected ok, got %v", got)
}
