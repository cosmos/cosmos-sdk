package stake

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondedToUnbondedPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	candA := Candidate{
		Status:      Bonded,
		Address:     addrs[0],
		PubKey:      pks[0],
		Assets:      sdk.OneRat,
		Liabilities: sdk.OneRat,
	}
	poolB, candB := poolA.bondedToUnbondedPool(candA)

	// status unbonded
	assert.Equal(t, candB.Status, Unbonded)
	// same exchange rate, assets unchanged
	assert.Equal(t, candB.Assets, candA.Assets)
	// bonded pool decreased
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool-candA.Assets.Evaluate())
	// unbonded pool increased
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool+candA.Assets.Evaluate())
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool, poolA.BondedPool+poolA.UnbondedPool)
}

func TestUnbonbedtoBondedPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	candA := Candidate{
		Status:      Bonded,
		Address:     addrs[0],
		PubKey:      pks[0],
		Assets:      sdk.OneRat,
		Liabilities: sdk.OneRat,
	}
	candA.Status = Unbonded
	poolB, candB := poolA.unbondedToBondedPool(candA)

	// status bonded
	assert.Equal(t, candB.Status, Bonded)
	// same exchange rate, assets unchanged
	assert.Equal(t, candB.Assets, candA.Assets)
	// bonded pool increased
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool+candA.Assets.Evaluate())
	// unbonded pool decreased
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool-candA.Assets.Evaluate())
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool, poolA.BondedPool+poolA.UnbondedPool)
}

func TestAddTokensBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	poolB, sharesB := poolA.addTokensBonded(10)
	assert.Equal(t, poolB.bondedShareExRate(), sdk.OneRat)

	// correct changes to bonded shares and bonded pool
	assert.Equal(t, poolB.BondedShares, poolA.BondedShares.Add(sharesB))
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool+10)

	// same number of bonded shares / tokens when exchange rate is one
	assert.Equal(t, poolB.BondedShares, sdk.NewRat(poolB.BondedPool))
}

func TestRemoveSharesBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	poolB, tokensB := poolA.removeSharesBonded(sdk.NewRat(10))
	assert.Equal(t, poolB.bondedShareExRate(), sdk.OneRat)

	// correct changes to bonded shares and bonded pool
	assert.Equal(t, poolB.BondedShares, poolA.BondedShares.Sub(sdk.NewRat(10)))
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool-tokensB)

	// same number of bonded shares / tokens when exchange rate is one
	assert.Equal(t, poolB.BondedShares, sdk.NewRat(poolB.BondedPool))
}

func TestAddTokensUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	poolB, sharesB := poolA.addTokensUnbonded(10)
	assert.Equal(t, poolB.unbondedShareExRate(), sdk.OneRat)

	// correct changes to unbonded shares and unbonded pool
	assert.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Add(sharesB))
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool+10)

	// same number of unbonded shares / tokens when exchange rate is one
	assert.Equal(t, poolB.UnbondedShares, sdk.NewRat(poolB.UnbondedPool))
}

func TestRemoveSharesUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	poolB, tokensB := poolA.removeSharesUnbonded(sdk.NewRat(10))
	assert.Equal(t, poolB.unbondedShareExRate(), sdk.OneRat)

	// correct changes to unbonded shares and bonded pool
	assert.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Sub(sdk.NewRat(10)))
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool-tokensB)

	// same number of unbonded shares / tokens when exchange rate is one
	assert.Equal(t, poolB.UnbondedShares, sdk.NewRat(poolB.UnbondedPool))
}

func TestCandidateAddTokens(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	candA := Candidate{
		Status:      Bonded,
		Address:     addrs[0],
		PubKey:      pks[0],
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
	}
	poolA.BondedPool = candA.Assets.Evaluate()
	poolA.BondedShares = candA.Assets
	assert.Equal(t, candA.delegatorShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	poolB, candB, sharesB := poolA.candidateAddTokens(candA, 10)

	// shares were issued
	assert.Equal(t, sdk.NewRat(10).Mul(candA.delegatorShareExRate()), sharesB)
	// pool shares were added
	assert.Equal(t, candB.Assets, candA.Assets.Add(sdk.NewRat(10)))
	// conservation of tokens
	assert.Equal(t, poolB.BondedPool, 10+poolA.BondedPool)
}

func TestCandidateRemoveShares(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)

	poolA := keeper.GetPool(ctx)
	candA := Candidate{
		Status:      Bonded,
		Address:     addrs[0],
		PubKey:      pks[0],
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
	}
	poolA.BondedPool = candA.Assets.Evaluate()
	poolA.BondedShares = candA.Assets
	assert.Equal(t, candA.delegatorShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	poolB, candB, coinsB := poolA.candidateRemoveShares(candA, sdk.NewRat(10))

	// coins were created
	assert.Equal(t, coinsB, int64(10))
	// pool shares were removed
	assert.Equal(t, candB.Assets, candA.Assets.Sub(sdk.NewRat(10).Mul(candA.delegatorShareExRate())))
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool+coinsB, poolA.UnbondedPool+poolA.BondedPool)
}

/////////////////////////////////////
// TODO Make all random tests less obfuscated!

// generate a random candidate
func randomCandidate(r *rand.Rand) Candidate {
	var status CandidateStatus
	if r.Float64() < float64(0.5) {
		status = Bonded
	} else {
		status = Unbonded
	}
	assets := sdk.NewRat(int64(r.Int31n(10000)))
	liabilities := sdk.NewRat(int64(r.Int31n(10000)))
	return Candidate{
		Status:      status,
		Address:     addrs[0],
		PubKey:      pks[0],
		Assets:      assets,
		Liabilities: liabilities,
	}
}

// generate a random staking state
func randomSetup(r *rand.Rand) (Pool, Candidate) {
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      sdk.ZeroRat,
		UnbondedShares:    sdk.ZeroRat,
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}

	candidate := randomCandidate(r)
	if candidate.Status == Bonded {
		pool.BondedShares = pool.BondedShares.Add(candidate.Assets)
		pool.BondedPool += candidate.Assets.Evaluate()
	} else {
		pool.UnbondedShares = pool.UnbondedShares.Add(candidate.Assets)
		pool.UnbondedPool += candidate.Assets.Evaluate()
	}
	return pool, candidate
}

func randomTokens(r *rand.Rand) int64 {
	return int64(r.Int31n(10000))
}

// operation that transforms staking state
type Operation func(p Pool, c Candidate) (Pool, Candidate, int64, string)

// pick a random staking operation
func randomOperation(r *rand.Rand) Operation {
	operations := []Operation{

		// bond/unbond
		func(p Pool, cand Candidate) (Pool, Candidate, int64, string) {

			var msg string
			if cand.Status == Bonded {
				msg = fmt.Sprintf("Unbonded previously bonded candidate %s (assets: %d, liabilities: %d, delegatorShareExRate: %v)",
					cand.Address, cand.Assets.Evaluate(), cand.Liabilities.Evaluate(), cand.delegatorShareExRate())
				p, cand = p.bondedToUnbondedPool(cand)
			} else {
				msg = fmt.Sprintf("Bonded previously unbonded candidate %s (assets: %d, liabilities: %d, delegatorShareExRate: %v)",
					cand.Address, cand.Assets.Evaluate(), cand.Liabilities.Evaluate(), cand.delegatorShareExRate())
				p, cand = p.unbondedToBondedPool(cand)
			}
			return p, cand, 0, msg
		},

		// add some tokens to a candidate
		func(p Pool, cand Candidate) (Pool, Candidate, int64, string) {

			tokens := int64(r.Int31n(1000))

			msg := fmt.Sprintf("candidate %s (assets: %d, liabilities: %d, delegatorShareExRate: %v)",
				cand.Address, cand.Assets.Evaluate(), cand.Liabilities.Evaluate(), cand.delegatorShareExRate())

			p, cand, _ = p.candidateAddTokens(cand, tokens)

			msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
			return p, cand, -1 * tokens, msg // tokens are removed so for accounting must be negative
		},

		// remove some shares from a candidate
		func(p Pool, cand Candidate) (Pool, Candidate, int64, string) {

			shares := sdk.NewRat(int64(r.Int31n(1000)))

			if shares.GT(cand.Liabilities) {
				shares = cand.Liabilities.Quo(sdk.NewRat(2))
			}

			msg := fmt.Sprintf("candidate %s (assets: %d, liabilities: %d, delegatorShareExRate: %v)",
				cand.Address, cand.Assets.Evaluate(), cand.Liabilities.Evaluate(), cand.delegatorShareExRate())
			p, cand, tokens := p.candidateRemoveShares(cand, shares)

			msg = fmt.Sprintf("Removed %d shares from %s", shares.Evaluate(), msg)

			return p, cand, tokens, msg
		},
	}
	r.Shuffle(len(operations), func(i, j int) {
		operations[i], operations[j] = operations[j], operations[i]
	})
	return operations[0]
}

// ensure invariants that should always be true are true
func assertInvariants(t *testing.T, msg string,
	pOrig Pool, cOrig Candidate, pMod Pool, cMod Candidate, tokens int64) {

	// total tokens conserved
	require.Equal(t,
		pOrig.UnbondedPool+pOrig.BondedPool,
		pMod.UnbondedPool+pMod.BondedPool+tokens,
		"msg: %v\n, pOrig.UnbondedPool: %v, pOrig.BondedPool: %v, pMod.UnbondedPool: %v, pMod.BondedPool: %v, tokens: %v\n",
		msg,
		pOrig.UnbondedPool, pOrig.BondedPool,
		pMod.UnbondedPool, pMod.BondedPool, tokens)

	// nonnegative shares
	require.False(t, pMod.BondedShares.LT(sdk.ZeroRat),
		"msg: %v\n, pOrig: %v\n, pMod: %v\n, cOrig: %v\n, cMod %v, tokens: %v\n",
		msg, pOrig, pMod, cOrig, cMod, tokens)
	require.False(t, pMod.UnbondedShares.LT(sdk.ZeroRat),
		"msg: %v\n, pOrig: %v\n, pMod: %v\n, cOrig: %v\n, cMod %v, tokens: %v\n",
		msg, pOrig, pMod, cOrig, cMod, tokens)

	// nonnegative ex rates
	require.False(t, pMod.bondedShareExRate().LT(sdk.ZeroRat),
		"Applying operation \"%s\" resulted in negative bondedShareExRate: %d",
		msg, pMod.bondedShareExRate().Evaluate())

	require.False(t, pMod.unbondedShareExRate().LT(sdk.ZeroRat),
		"Applying operation \"%s\" resulted in negative unbondedShareExRate: %d",
		msg, pMod.unbondedShareExRate().Evaluate())

	// nonnegative ex rate
	require.False(t, cMod.delegatorShareExRate().LT(sdk.ZeroRat),
		"Applying operation \"%s\" resulted in negative candidate.delegatorShareExRate(): %v (candidate.PubKey: %s)",
		msg,
		cMod.delegatorShareExRate(),
		cMod.PubKey,
	)

	// nonnegative assets / liabilities
	require.False(t, cMod.Assets.LT(sdk.ZeroRat),
		"Applying operation \"%s\" resulted in negative candidate.Assets: %d (candidate.Liabilities: %d, candidate.PubKey: %s)",
		msg,
		cMod.Assets.Evaluate(),
		cMod.Liabilities.Evaluate(),
		cMod.PubKey,
	)

	require.False(t, cMod.Liabilities.LT(sdk.ZeroRat),
		"Applying operation \"%s\" resulted in negative candidate.Liabilities: %d (candidate.Assets: %d, candidate.PubKey: %s)",
		msg,
		cMod.Liabilities.Evaluate(),
		cMod.Assets.Evaluate(),
		cMod.PubKey,
	)
}

// run random operations in a random order on a random state, assert invariants hold
func TestIntegrationInvariants(t *testing.T) {
	for i := 0; i < 10; i++ {

		r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
		pool, candidates := randomSetup(r1)
		initialPool, initialCandidates := pool, candidates

		assertInvariants(t, "no operation",
			initialPool, initialCandidates,
			pool, candidates, 0)

		for j := 0; j < 100; j++ {

			r2 := rand.New(rand.NewSource(time.Now().UnixNano()))
			pool, candidates, tokens, msg := randomOperation(r2)(pool, candidates)

			assertInvariants(t, msg,
				initialPool, initialCandidates,
				pool, candidates, tokens)
		}
	}
}
