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
		Bond:          sdk.Coin{"steak", amt},
		PubKey:        pubKey,
	}
}

func newTestMsgDelegate(delegatorAddr, candidateAddr sdk.Address, amt int64) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delegatorAddr,
		CandidateAddr: candidateAddr,
		Bond:          sdk.Coin{"steak", amt},
	}
}

//______________________________________________________________________

func TestDuplicatesMsgDeclareCandidacy(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000)

	candidateAddr := addrs[0]
	pk := pks[0]
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pk, 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "%v", got)
	candidate, found := keeper.GetCandidate(ctx, candidateAddr)
	require.True(t, found)
	assert.Equal(t, Unbonded, candidate.Status)
	assert.Equal(t, candidateAddr, candidate.Address)
	assert.Equal(t, pk, candidate.PubKey)
	assert.Equal(t, sdk.NewRat(10), candidate.Assets)
	assert.Equal(t, sdk.NewRat(10), candidate.Liabilities)
	assert.Equal(t, Description{}, candidate.Description)

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

	candidate, found := keeper.GetCandidate(ctx, candidateAddr)
	require.True(t, found)
	assert.Equal(t, bondAmount, candidate.Liabilities.Evaluate())
	assert.Equal(t, bondAmount, candidate.Assets.Evaluate())

	// just send the same msgbond multiple times
	msgDelegate := newTestMsgDelegate(delegatorAddr, candidateAddr, bondAmount)

	for i := 0; i < 5; i++ {
		ctx = ctx.WithBlockHeight(int64(i))

		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidate, found := keeper.GetCandidate(ctx, candidateAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegatorBond(ctx, delegatorAddr, candidateAddr)
		require.True(t, found)

		expBond := int64(i+1) * bondAmount
		expLiabilities := int64(i+2) * bondAmount // (1 self delegation)
		expDelegatorAcc := initBond - expBond

		require.Equal(t, bond.Height, int64(i), "Incorrect bond height")

		gotBond := bond.Shares.Evaluate()
		gotLiabilities := candidate.Liabilities.Evaluate()
		gotDelegatorAcc := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(params.BondDenom)

		require.Equal(t, expBond, gotBond,
			"i: %v\nexpBond: %v\ngotBond: %v\ncandidate: %v\nbond: %v\n",
			i, expBond, gotBond, candidate, bond)
		require.Equal(t, expLiabilities, gotLiabilities,
			"i: %v\nexpLiabilities: %v\ngotLiabilities: %v\ncandidate: %v\nbond: %v\n",
			i, expLiabilities, gotLiabilities, candidate, bond)
		require.Equal(t, expDelegatorAcc, gotDelegatorAcc,
			"i: %v\nexpDelegatorAcc: %v\ngotDelegatorAcc: %v\ncandidate: %v\nbond: %v\n",
			i, expDelegatorAcc, gotDelegatorAcc, candidate, bond)
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

	msgDelegate := newTestMsgDelegate(delegatorAddr, candidateAddr, initBond)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	candidate, found := keeper.GetCandidate(ctx, candidateAddr)
	require.True(t, found)
	assert.Equal(t, initBond*2, candidate.Liabilities.Evaluate())
	assert.Equal(t, initBond*2, candidate.Assets.Evaluate())

	// just send the same msgUnbond multiple times
	// TODO use decimals here
	unbondShares, unbondSharesStr := int64(10), "10"
	msgUnbond := NewMsgUnbond(delegatorAddr, candidateAddr, unbondSharesStr)
	numUnbonds := 5
	for i := 0; i < numUnbonds; i++ {
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		candidate, found = keeper.GetCandidate(ctx, candidateAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegatorBond(ctx, delegatorAddr, candidateAddr)
		require.True(t, found)

		expBond := initBond - int64(i+1)*unbondShares
		expLiabilities := 2*initBond - int64(i+1)*unbondShares
		expDelegatorAcc := initBond - expBond

		gotBond := bond.Shares.Evaluate()
		gotLiabilities := candidate.Liabilities.Evaluate()
		gotDelegatorAcc := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(params.BondDenom)

		require.Equal(t, expBond, gotBond,
			"i: %v\nexpBond: %v\ngotBond: %v\ncandidate: %v\nbond: %v\n",
			i, expBond, gotBond, candidate, bond)
		require.Equal(t, expLiabilities, gotLiabilities,
			"i: %v\nexpLiabilities: %v\ngotLiabilities: %v\ncandidate: %v\nbond: %v\n",
			i, expLiabilities, gotLiabilities, candidate, bond)
		require.Equal(t, expDelegatorAcc, gotDelegatorAcc,
			"i: %v\nexpDelegatorAcc: %v\ngotDelegatorAcc: %v\ncandidate: %v\nbond: %v\n",
			i, expDelegatorAcc, gotDelegatorAcc, candidate, bond)
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
		require.False(t, got.IsOK(), "expected unbond msg to fail")
	}

	leftBonded := initBond - unbondShares*int64(numUnbonds)

	// should be unable to unbond one more than we have
	unbondSharesStr = strconv.Itoa(int(leftBonded) + 1)
	msgUnbond = NewMsgUnbond(delegatorAddr, candidateAddr, unbondSharesStr)
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	assert.False(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgUnbond, unbondSharesStr, leftBonded)

	// should be able to unbond just what we have
	unbondSharesStr = strconv.Itoa(int(leftBonded))
	msgUnbond = NewMsgUnbond(delegatorAddr, candidateAddr, unbondSharesStr)
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	assert.True(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgUnbond, unbondSharesStr, leftBonded)
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
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		candidates := keeper.GetCandidates(ctx, 100)
		require.Equal(t, (i + 1), len(candidates))
		val := candidates[i]
		balanceExpd := initBond - 10
		balanceGot := accMapper.GetAccount(ctx, val.Address).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, i+1, len(candidates), "expected %d candidates got %d, candidates: %v", i+1, len(candidates), candidates)
		require.Equal(t, 10, int(val.Liabilities.Evaluate()), "expected %d shares, got %d", 10, val.Liabilities)
		require.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, candidateAddr := range candidateAddrs {
		candidatePre, found := keeper.GetCandidate(ctx, candidateAddr)
		require.True(t, found)
		msgUnbond := NewMsgUnbond(candidateAddr, candidateAddr, "10") // self-delegation
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		candidates := keeper.GetCandidates(ctx, 100)
		require.Equal(t, len(candidateAddrs)-(i+1), len(candidates),
			"expected %d candidates got %d", len(candidateAddrs)-(i+1), len(candidates))

		_, found = keeper.GetCandidate(ctx, candidateAddr)
		require.False(t, found)

		expBalance := initBond
		gotBalance := accMapper.GetAccount(ctx, candidatePre.Address).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, expBalance, gotBalance, "expected account to have %d, got %d", expBalance, gotBalance)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000)
	candidateAddr, delegatorAddrs := addrs[0], addrs[1:]

	//first make a candidate
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[0], 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// delegate multiple parties
	for i, delegatorAddr := range delegatorAddrs {
		msgDelegate := newTestMsgDelegate(delegatorAddr, candidateAddr, 10)
		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond, found := keeper.GetDelegatorBond(ctx, delegatorAddr, candidateAddr)
		require.True(t, found)
		require.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegatorAddr := range delegatorAddrs {
		msgUnbond := NewMsgUnbond(delegatorAddr, candidateAddr, "10")
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		_, found := keeper.GetDelegatorBond(ctx, delegatorAddr, candidateAddr)
		require.False(t, found)
	}
}

func TestVoidCandidacy(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000)
	candidateAddr, delegatorAddr := addrs[0], addrs[1]

	// create the candidate
	msgDeclareCandidacy := newTestMsgDeclareCandidacy(candidateAddr, pks[0], 10)
	got := handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDeclareCandidacy")

	// bond a delegator
	msgDelegate := newTestMsgDelegate(delegatorAddr, candidateAddr, 10)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	// unbond the candidates bond portion
	msgUnbondCandidate := NewMsgUnbond(candidateAddr, candidateAddr, "10")
	got = handleMsgUnbond(ctx, msgUnbondCandidate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDeclareCandidacy")
	candidate, found := keeper.GetCandidate(ctx, candidateAddr)
	require.True(t, found)
	require.Equal(t, Revoked, candidate.Status)

	// test that this address cannot yet be bonded too because is revoked
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.False(t, got.IsOK(), "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	msgUnbondDelegator := NewMsgUnbond(delegatorAddr, candidateAddr, "10")
	got = handleMsgUnbond(ctx, msgUnbondDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDeclareCandidacy")

	// verify that the pubkey can now be reused
	got = handleMsgDeclareCandidacy(ctx, msgDeclareCandidacy, keeper)
	assert.True(t, got.IsOK(), "expected ok, got %v", got)
}
