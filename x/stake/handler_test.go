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

func newTestMsgDeclareCandidacy(address sdk.Address, pubKey crypto.PubKey, amt int64) MsgDeclareCandidacy {
	return MsgDeclareCandidacy{
		Description:   Description{},
		CandidateAddr: address,
		Bond:          sdk.Coin{"fermion", amt},
		PubKey:        pubKey,
	}
}

func newTestMsgDelegate(amt int64, delegatorAddr, candidateAddr sdk.Address) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delegatorAddr,
		CandidateAddr: candidateAddr,
		Bond:          sdk.Coin{"fermion", amt},
	}
}

func TestDuplicatesMsgDeclareCandidacy(t *testing.T) {
	ctx, _, keeper := createTestInput(t, addrs[0], false, 1000)

	msgDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "%v", got)

	// one sender cannot bond twice
	msgDeclareCandidacy.PubKey = pks[1]
	got = handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.False(t, got.IsOK(), "%v", got)
}

func TestIncrementsMsgDelegate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, addrs[0], false, 1000)

	// first declare candidacy
	bondAmount := int64(10)
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], bondAmount)
	got := deliverer.declareCandidacy(msgDeclareCandidacy)
	assert.NoError(t, got, "expected declare candidacy msg to be ok, got %v", got)
	expectedBond := bondAmount // 1 since we send 1 at the start of loop,

	// just send the same msgbond multiple times
	msgDelegate := newTestMsgDelegate(bondAmount, addrs[0])
	for i := 0; i < 5; i++ {
		got := deliverer.delegate(msgDelegate)
		assert.NoError(t, got, "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := mapper.GetCandidates()
		expectedBond += bondAmount
		//expectedSender := initSender - expectedBond
		gotBonded := candidates[0].Liabilities.Evaluate()
		//gotSender := accStore[string(deliverer.sender)] //XXX use StoreMapper
		assert.Equal(t, expectedBond, gotBonded, "i: %v, %v, %v", i, expectedBond, gotBonded)
		//assert.Equal(t, expectedSender, gotSender, "i: %v, %v, %v", i, expectedSender, gotSender) // XXX fix
	}
}

func TestIncrementsMsgUnbond(t *testing.T) {
	ctx, _, keeper := createTestInput(t, addrs[0], false, 0)

	// set initial bond
	initBond := int64(1000)
	//accStore[string(deliverer.sender)] = initBond //XXX use StoreMapper
	got := deliverer.declareCandidacy(newTestMsgDeclareCandidacy(addrs[0], pks[0], initBond))
	assert.NoError(t, got, "expected initial bond msg to be ok, got %v", got)

	// just send the same msgunbond multiple times
	// XXX use decimals here
	unbondShares, unbondSharesStr := int64(10), "10"
	msgUndelegate := NewMsgUnbond(addrs[0], unbondSharesStr)
	nUnbonds := 5
	for i := 0; i < nUnbonds; i++ {
		got := deliverer.unbond(msgUndelegate)
		assert.NoError(t, got, "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := mapper.GetCandidates()
		expectedBond := initBond - int64(i+1)*unbondShares // +1 since we send 1 at the start of loop
		//expectedSender := initSender + (initBond - expectedBond)
		gotBonded := candidates[0].Liabilities.Evaluate()
		//gotSender := accStore[string(deliverer.sender)] // XXX use storemapper

		assert.Equal(t, expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
		//assert.Equal(t, expectedSender, gotSender, "%v, %v", expectedSender, gotSender) //XXX fix
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
		msgUndelegate := NewMsgUnbond(addrs[0], unbondShares)
		got = deliverer.unbond(msgUndelegate)
		assert.Error(t, got, "expected unbond msg to fail")
	}

	leftBonded := initBond - unbondShares*int64(nUnbonds)

	// should be unable to unbond one more than we have
	msgUndelegate = NewMsgUnbond(addrs[0], strconv.Itoa(int(leftBonded)+1))
	got = deliverer.unbond(msgUndelegate)
	assert.Error(t, got, "expected unbond msg to fail")

	// should be able to unbond just what we have
	msgUndelegate = NewMsgUnbond(addrs[0], strconv.Itoa(int(leftBonded)))
	got = deliverer.unbond(msgUndelegate)
	assert.NoError(t, got, "expected unbond msg to pass")
}

func TestMultipleMsgDeclareCandidacy(t *testing.T) {
	initSender := int64(1000)
	//ctx, accStore, mapper, deliverer := createTestInput(t, addrs[0], false, initSender)
	ctx, mapper, keeper := createTestInput(t, addrs[0], false, initSender)
	addrs := []sdk.Address{addrs[0], addrs[1], addrs[2]}

	// bond them all
	for i, addr := range addrs {
		msgDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[i], pks[i], 10)
		deliverer.sender = addr
		got := deliverer.declareCandidacy(msgDeclareCandidacy)
		assert.NoError(t, got, "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := mapper.GetCandidates()
		require.Equal(t, i, len(candidates))
		val := candidates[i]
		balanceExpd := initSender - 10
		balanceGot := accStore.GetAccount(ctx, val.Address).GetCoins()
		assert.Equal(t, i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		assert.Equal(t, 10, int(val.Liabilities.Evaluate()), "expected %d shares, got %d", 10, val.Liabilities)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, addr := range addrs {
		candidatePre := mapper.GetCandidate(addrs[i])
		msgUndelegate := NewMsgUnbond(addrs[i], "10")
		deliverer.sender = addr
		got := deliverer.unbond(msgUndelegate)
		assert.NoError(t, got, "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := mapper.GetCandidates()
		assert.Equal(t, len(addrs)-(i+1), len(candidates), "expected %d candidates got %d", len(addrs)-(i+1), len(candidates))

		candidatePost := mapper.GetCandidate(addrs[i])
		balanceExpd := initSender
		balanceGot := accStore.GetAccount(ctx, candidatePre.Address).GetCoins()
		assert.Nil(t, candidatePost, "expected nil candidate retrieve, got %d", 0, candidatePost)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	sender, delegators := addrs[0], addrs[1:]
	_, _, mapper, deliverer := createTestInput(t, addrs[0], false, 1000)
	ctx, _, keeper := createTestInput(t, addrs[0], false, 0)

	//first make a candidate
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(sender, pks[0], 10)
	got := deliverer.declareCandidacy(msgDeclareCandidacy)
	require.NoError(t, got, "expected msg to be ok, got %v", got)

	// delegate multiple parties
	for i, delegator := range delegators {
		msgDelegate := newTestMsgDelegate(10, sender)
		deliverer.sender = delegator
		got := deliverer.delegate(msgDelegate)
		require.NoError(t, got, "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond := mapper.getDelegatorBond(delegator, sender)
		assert.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegator := range delegators {
		msgUndelegate := NewMsgUnbond(sender, "10")
		deliverer.sender = delegator
		got := deliverer.unbond(msgUndelegate)
		require.NoError(t, got, "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		bond := mapper.getDelegatorBond(delegator, sender)
		assert.Nil(t, bond, "expected delegatee bond %d to be nil", bond)
	}
}

func TestVoidCandidacy(t *testing.T) {
	sender, delegator := addrs[0], addrs[1]
	_, _, _, deliverer := createTestInput(t, addrs[0], false, 1000)

	// create the candidate
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], 10)
	got := deliverer.declareCandidacy(msgDeclareCandidacy)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// bond a delegator
	msgDelegate := newTestMsgDelegate(10, addrs[0])
	deliverer.sender = delegator
	got = deliverer.delegate(msgDelegate)
	require.NoError(t, got, "expected ok, got %v", got)

	// unbond the candidates bond portion
	msgUndelegate := NewMsgUnbond(addrs[0], "10")
	deliverer.sender = sender
	got = deliverer.unbond(msgUndelegate)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// test that this pubkey cannot yet be bonded too
	deliverer.sender = delegator
	got = deliverer.delegate(msgDelegate)
	assert.Error(t, got, "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	got = deliverer.unbond(msgUndelegate)
	require.NoError(t, got, "expected no error on runMsgDeclareCandidacy")

	// verify that the pubkey can now be reused
	got = deliverer.declareCandidacy(msgDeclareCandidacy)
	assert.NoError(t, got, "expected ok, got %v", got)
}
