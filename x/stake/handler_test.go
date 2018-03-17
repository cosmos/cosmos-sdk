package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//______________________________________________________________________

// XXX delete need to init accounts in the transact!
func initAccounts(amount int64) map[string]int64 {
	accStore := map[string]int64{}
	for _, addr := range addrs {
		accStore[string(addr)] = amount
	}
	return accStore
}

func newTestMsgDeclareCandidacy(address sdk.Address, pubKey crypto.PubKey, amt int64) MsgDeclareCandidacy {
	return MsgDeclareCandidacy{
		MsgAddr:     NewMsgAddr(address),
		PubKey:      pubKey,
		Bond:        sdk.Coin{"fermion", amt},
		Description: Description{},
	}
}

func newTestMsgDelegate(amt int64, address sdk.Address) MsgDelegate {
	return MsgDelegate{
		MsgAddr: NewMsgAddr(address),
		Bond:    sdk.Coin{"fermion", amt},
	}
}

func TestDuplicatesMsgDeclareCandidacy(t *testing.T) {
	accStore := initAccounts(1000) // for accounts
	_, deliverer := createTestInput(t, addrs[0], false)
	_, checker := createTestInput(t, addrs[0], true)

	txDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], 10)
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// one sender can bond to two different addresses
	txDeclareCandidacy.Address = addrs[1]
	err := checker.declareCandidacy(txDeclareCandidacy)
	assert.Nil(t, err, "didn't expected error on checkTx")

	// two addrs cant bond to the same pubkey
	checker.sender = addrs[1]
	txDeclareCandidacy.Address = addrs[0]
	err = checker.declareCandidacy(txDeclareCandidacy)
	assert.NotNil(t, err, "expected error on checkTx")
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initSender := int64(1000)
	accStore := initAccounts(initSender) // for accounts
	mapper, deliverer := createTestInput(t, addrs[0], false)

	// first declare candidacy
	bondAmount := int64(10)
	txDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], bondAmount)
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	assert.NoError(t, got, "expected declare candidacy tx to be ok, got %v", got)
	expectedBond := bondAmount // 1 since we send 1 at the start of loop,

	// just send the same txbond multiple times
	txDelegate := newTestMsgDelegate(bondAmount, addrs[0])
	for i := 0; i < 5; i++ {
		got := deliverer.delegate(txDelegate)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := mapper.loadCandidates()
		expectedBond += bondAmount
		expectedSender := initSender - expectedBond
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotSender := accStore[string(deliverer.sender)]
		assert.Equal(t, expectedBond, gotBonded, "i: %v, %v, %v", i, expectedBond, gotBonded)
		assert.Equal(t, expectedSender, gotSender, "i: %v, %v, %v", i, expectedSender, gotSender)
	}
}

func TestIncrementsMsgUnbond(t *testing.T) {
	initSender := int64(0)
	accStore := initAccounts(initSender) // for accounts
	mapper, deliverer := createTestInput(t, addrs[0], false)

	// set initial bond
	initBond := int64(1000)
	accStore[string(deliverer.sender)] = initBond
	got := deliverer.declareCandidacy(newTestMsgDeclareCandidacy(addrs[0], pks[0], initBond))
	assert.NoError(t, got, "expected initial bond tx to be ok, got %v", got)

	// just send the same txunbond multiple times
	// XXX use decimals here
	unbondShares, unbondSharesStr := int64(10), "10"
	txUndelegate := NewMsgUnbond(addrs[0], unbondSharesStr)
	nUnbonds := 5
	for i := 0; i < nUnbonds; i++ {
		got := deliverer.unbond(txUndelegate)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := mapper.loadCandidates()
		expectedBond := initBond - int64(i+1)*unbondShares // +1 since we send 1 at the start of loop
		expectedSender := initSender + (initBond - expectedBond)
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotSender := accStore[string(deliverer.sender)]

		assert.Equal(t, expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
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
		txUndelegate := NewMsgUnbond(addrs[0], unbondShares)
		got = deliverer.unbond(txUndelegate)
		assert.Error(t, got, "expected unbond tx to fail")
	}

	leftBonded := initBond - unbondShares*int64(nUnbonds)

	// should be unable to unbond one more than we have
	txUndelegate = NewMsgUnbond(addrs[0], strconv.Itoa(int(leftBonded)+1))
	got = deliverer.unbond(txUndelegate)
	assert.Error(t, got, "expected unbond tx to fail")

	// should be able to unbond just what we have
	txUndelegate = NewMsgUnbond(addrs[0], strconv.Itoa(int(leftBonded)))
	got = deliverer.unbond(txUndelegate)
	assert.NoError(t, got, "expected unbond tx to pass")
}

func TestMultipleMsgDeclareCandidacy(t *testing.T) {
	initSender := int64(1000)
	accStore := initAccounts(initSender)
	addrs := []sdk.Address{addrs[0], addrs[1], addrs[2]}
	mapper, deliverer := createTestInput(t, addrs[0], false)

	// bond them all
	for i, addr := range addrs {
		txDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[i], pks[i], 10)
		deliverer.sender = addr
		got := deliverer.declareCandidacy(txDeclareCandidacy)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := mapper.loadCandidates()
		val := candidates[i]
		balanceGot, balanceExpd := accStore[string(val.Address)], initSender-10
		assert.Equal(t, i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		assert.Equal(t, 10, int(val.Liabilities.Evaluate()), "expected %d shares, got %d", 10, val.Liabilities)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, addr := range addrs {
		candidatePre := mapper.loadCandidate(addrs[i])
		txUndelegate := NewMsgUnbond(addrs[i], "10")
		deliverer.sender = addr
		got := deliverer.unbond(txUndelegate)
		assert.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := mapper.loadCandidates()
		assert.Equal(t, len(addrs)-(i+1), len(candidates), "expected %d candidates got %d", len(addrs)-(i+1), len(candidates))

		candidatePost := mapper.loadCandidate(addrs[i])
		balanceGot, balanceExpd := accStore[string(candidatePre.Owner.Address)], initSender
		assert.Nil(t, candidatePost, "expected nil candidate retrieve, got %d", 0, candidatePost)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	accStore := initAccounts(1000)
	sender, delegators := addrs[0], addrs[1:]
	mapper, deliverer := createTestInput(t, addrs[0], false)

	//first make a candidate
	txDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], 10)
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	require.NoError(t, got, "expected tx to be ok, got %v", got)

	// delegate multiple parties
	for i, delegator := range delegators {
		txDelegate := newTestMsgDelegate(10, addrs[0])
		deliverer.sender = delegator
		got := deliverer.delegate(txDelegate)
		require.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond := mapper.loadDelegatorBond(delegator, addrs[0])
		assert.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegator := range delegators {
		txUndelegate := NewMsgUnbond(addrs[0], "10")
		deliverer.sender = delegator
		got := deliverer.unbond(txUndelegate)
		require.NoError(t, got, "expected tx %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		bond := mapper.loadDelegatorBond(delegator, addrs[0])
		assert.Nil(t, bond, "expected delegatee bond %d to be nil", bond)
	}
}

func TestVoidCandidacy(t *testing.T) {
	accStore := initAccounts(1000) // for accounts
	sender, delegator := addrs[0], addrs[1]
	_, deliverer := createTestInput(t, addrs[0], false)

	// create the candidate
	txDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], 10)
	got := deliverer.declareCandidacy(txDeclareCandidacy)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// bond a delegator
	txDelegate := newTestMsgDelegate(10, addrs[0])
	deliverer.sender = delegator
	got = deliverer.delegate(txDelegate)
	require.NoError(t, got, "expected ok, got %v", got)

	// unbond the candidates bond portion
	txUndelegate := NewMsgUnbond(addrs[0], "10")
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
