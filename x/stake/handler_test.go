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
	ctx, _, keeper := createTestInput(t, false, 1000)

	msgDeclareCandidacy := newTestMsgDeclareCandidacy(addrs[0], pks[0], 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "%v", got)

	// one candidate cannot bond twice
	msgDeclareCandidacy.PubKey = pks[1]
	got = handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.False(t, got.IsOK(), "%v", got)
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := createTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)

	bondAmount := int64(10)
	candidateAddr, delegatorAddr := addrs[0], addrs[1]

	// first declare candidacy
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[0], bondAmount)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "expected declare candidacy msg to be ok, got %v", got)
	expectedBond := bondAmount

	// just send the same msgbond multiple times
	msgDelegate := newTestMsgDelegate(bondAmount, delegatorAddr, candidateAddr)
	for i := 0; i < 5; i++ {
		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		assert.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidates := keeper.GetCandidates(ctx, 100)
		expectedBond += bondAmount
		expectedDelegator := initBond - expectedBond
		gotBonded := candidates[0].Liabilities.Evaluate()
		gotDelegator := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(params.BondDenom)
		assert.Equal(t, expectedBond, gotBonded, "i: %v, %v, %v", i, expectedBond, gotBonded)
		assert.Equal(t, expectedDelegator, gotDelegator, "i: %v, %v, %v", i, expectedDelegator, gotDelegator) // XXX fix
	}
}

func TestIncrementsMsgUnbond(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := createTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)

	// declare candidacy, delegate
	candidateAddr, delegatorAddr := addrs[0], addrs[1]
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[0], initBond)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "expected declare-candidacy to be ok, got %v", got)
	msgDelegate := newTestMsgDelegate(initBond, delegatorAddr, candidateAddr)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	// just send the same msgUnbond multiple times
	// TODO use decimals here
	unbondShares, unbondSharesStr := int64(10), "10"
	msgUnbond := NewMsgUnbond(delegatorAddr, candidateAddr, unbondSharesStr)
	numUnbonds := 5
	for i := 0; i < numUnbonds; i++ {
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		assert.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidate, found := keeper.GetCandidate(ctx, candidateAddr)
		require.True(t, found)
		expectedBond := initBond - int64(i+1)*unbondShares
		expectedDelegator := initBond - expectedBond
		gotBonded := candidate.Liabilities.Evaluate()
		gotDelegator := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(params.BondDenom)

		assert.Equal(t, expectedBond, gotBonded, "%v, %v", expectedBond, gotBonded)
		assert.Equal(t, expectedDelegator, gotDelegator, "%v, %v", expectedDelegator, gotDelegator)
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
		msgUnbond := NewMsgUnbond(delegatorAddr, candidateAddr, unbondShares)
		got = handleMsgUnbond(ctx, msgUnbond, keeper)
		assert.False(t, got.IsOK(), "expected unbond msg to fail")
	}

	leftBonded := initBond - unbondShares*int64(numUnbonds)

	// should be unable to unbond one more than we have
	msgUnbond = NewMsgUnbond(delegatorAddr, candidateAddr, strconv.Itoa(int(leftBonded)+1))
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	assert.False(t, got.IsOK(), "expected unbond msg to fail")

	// should be able to unbond just what we have
	msgUnbond = NewMsgUnbond(delegatorAddr, candidateAddr, strconv.Itoa(int(leftBonded)))
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	assert.True(t, got.IsOK(), "expected unbond msg to pass")
}

func TestMultipleMsgDeclareCandidacy(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := createTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)
	candidateAddrs := []sdk.Address{addrs[0], addrs[1], addrs[2]}

	// bond them all
	for i, candidateAddr := range candidateAddrs {
		msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[i], 10)
		got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
		assert.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := keeper.GetCandidates(ctx, 100)
		require.Equal(t, i, len(candidates))
		val := candidates[i]
		balanceExpd := initBond - 10
		balanceGot := accMapper.GetAccount(ctx, val.Address).GetCoins().AmountOf(params.BondDenom)
		assert.Equal(t, i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		assert.Equal(t, 10, int(val.Liabilities.Evaluate()), "expected %d shares, got %d", 10, val.Liabilities)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, candidateAddr := range candidateAddrs {
		candidatePre, found := keeper.GetCandidate(ctx, candidateAddr)
		require.True(t, found)
		msgUnbond := NewMsgUnbond(candidateAddr, candidateAddr, "10") // self-delegation
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		assert.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := keeper.GetCandidates(ctx, 100)
		assert.Equal(t, len(candidateAddrs)-(i+1), len(candidates),
			"expected %d candidates got %d", len(candidateAddrs)-(i+1), len(candidates))

		candidatePost, found := keeper.GetCandidate(ctx, candidateAddr)
		require.True(t, found)
		balanceExpd := initBond
		balanceGot := accMapper.GetAccount(ctx, candidatePre.Address).GetCoins().AmountOf(params.BondDenom)
		assert.Nil(t, candidatePost, "expected nil candidate retrieve, got %d", 0, candidatePost)
		assert.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	candidateAddr, delegatorAddrs := addrs[0], addrs[1:]

	//first make a candidate
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[0], 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// delegate multiple parties
	for i, delegatorAddr := range delegatorAddrs {
		msgDelegate := newTestMsgDelegate(10, delegatorAddr, candidateAddr)
		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond, found := keeper.getDelegatorBond(ctx, delegatorAddr, candidateAddr)
		require.True(t, found)
		assert.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegatorAddr := range delegatorAddrs {
		msgUnbond := NewMsgUnbond(delegatorAddr, candidateAddr, "10")
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		_, found := keeper.getDelegatorBond(ctx, delegatorAddr, candidateAddr)
		require.False(t, found)
	}
}

func TestVoidCandidacy(t *testing.T) {
	candidateAddr, delegatorAddr := addrs[0], addrs[1]
	ctx, _, keeper := createTestInput(t, false, 0)

	// create the candidate
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[0], 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDeclareCandidacy")

	// bond a delegator
	msgDelegate := newTestMsgDelegate(10, delegatorAddr, candidateAddr)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	// unbond the candidates bond portion
	msgUnbond := NewMsgUnbond(delegatorAddr, candidateAddr, "10")
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDeclareCandidacy")

	// test that this pubkey cannot yet be bonded too
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.False(t, got.IsOK(), "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDeclareCandidacy")

	// verify that the pubkey can now be reused
	got = handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "expected ok, got %v", got)
}
