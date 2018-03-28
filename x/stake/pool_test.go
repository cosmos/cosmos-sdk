package stake

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

func TestBondedToUnbondedPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	candA := candidate1
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
	candA := candidate1
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
		Address:     addrVal1,
		PubKey:      pk1,
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
		Status:      Bonded,
	}
	poolA.BondedPool = candA.Assets.Evaluate()
	poolA.BondedShares = candA.Assets
	assert.Equal(t, candA.delegatorShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat)
	poolB, candB, sharesB := poolA.candidateAddTokens(candA, 10)
	// shares were issued
	assert.Equal(t, sharesB, sdk.NewRat(10).Mul(candA.delegatorShareExRate()))
	// pool shares were added
	assert.Equal(t, candB.Assets, candA.Assets.Add(sdk.NewRat(10)))
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool, 10+poolA.UnbondedPool+poolA.BondedPool)
}

func TestCandidateRemoveShares(t *testing.T) {
	ctx, _, keeper := createTestInput(t, nil, false, 0)
	poolA := keeper.GetPool(ctx)
	candA := Candidate{
		Address:     addrVal1,
		PubKey:      pk1,
		Assets:      sdk.NewRat(9),
		Liabilities: sdk.NewRat(9),
		Status:      Bonded,
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

// generate a random candidate
func randomCandidate(r *rand.Rand) Candidate {
	var status CandidateStatus
	if r.Float64() < float64(0.5) {
		status = Bonded
	} else {
		status = Unbonded
	}
	address := testAddr("A58856F0FD53BF058B4909A21AEC019107BA6160")
	pubkey := crypto.GenPrivKeyEd25519().PubKey()
	assets := sdk.NewRat(int64(r.Int31n(10000)))
	liabilities := sdk.NewRat(int64(r.Int31n(10000)))
	return Candidate{
		Status:      status,
		Address:     address,
		PubKey:      pubkey,
		Assets:      assets,
		Liabilities: liabilities,
	}
}

// generate a random staking state
func randomSetup(r *rand.Rand) (Pool, Candidates, int64) {
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      sdk.ZeroRat,
		UnbondedShares:    sdk.ZeroRat,
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	var candidates []Candidate
	for i := int32(0); i < r.Int31n(1000); i++ {
		candidate := randomCandidate(r)
		if candidate.Status == Bonded {
			pool.BondedShares = pool.BondedShares.Add(candidate.Assets)
			pool.BondedPool += candidate.Assets.Evaluate()
		} else {
			pool.UnbondedShares = pool.UnbondedShares.Add(candidate.Assets)
			pool.UnbondedPool += candidate.Assets.Evaluate()
		}
		candidates = append(candidates, candidate)
	}
	tokens := int64(r.Int31n(10000))
	return pool, candidates, tokens
}

// operation that transforms staking state
type Operation func(p Pool, c Candidates, t int64) (Pool, Candidates, int64)

// pick a random staking operation
func randomOperation(r *rand.Rand) Operation {
	operations := []Operation{
		// bond/unbond
		func(p Pool, c Candidates, t int64) (Pool, Candidates, int64) {
			index := int(r.Int31n(int32(len(c))))
			cand := c[index]
			if cand.Status == Bonded {
				p, cand = p.bondedToUnbondedPool(cand)
				cand.Status = Unbonded
			} else {
				p, cand = p.unbondedToBondedPool(cand)
				cand.Status = Bonded
			}
			c[index] = cand
			return p, c, t
		},
		// add some tokens to a candidate
		func(p Pool, c Candidates, t int64) (Pool, Candidates, int64) {
			tokens := int64(r.Int31n(1000))
			index := int(r.Int31n(int32(len(c))))
			cand := c[index]
			p, cand, _ = p.candidateAddTokens(cand, tokens)
			c[index] = cand
			t -= tokens
			return p, c, t
		},
		/*
			// remove some shares from a candidate
			func(p Pool, c Candidates, t int64) (Pool, Candidates, int64) {
				shares := sdk.NewRat(int64(r.Int31n(1000)))
				index := int(r.Int31n(int32(len(c))))
				cand := c[index]
				p, cand, tokens := p.candidateRemoveShares(cand, shares)
				c[index] = cand
				t += tokens
				return p, c, t
			},
		*/
	}
	r.Shuffle(len(operations), func(i, j int) {
		operations[i], operations[j] = operations[j], operations[i]
	})
	return operations[0]
}

// ensure invariants that should always be true are true
func assertInvariants(t *testing.T, pA Pool, cA Candidates, tA int64, pB Pool, cB Candidates, tB int64) {
	// total tokens conserved
	assert.Equal(t, pA.UnbondedPool+pA.BondedPool+tA, pB.UnbondedPool+pB.BondedPool+tB)
	// nonnegative shares
	assert.Equal(t, pB.BondedShares.LT(sdk.ZeroRat), false)
	assert.Equal(t, pB.UnbondedShares.LT(sdk.ZeroRat), false)
	bondedSharesHeld := sdk.ZeroRat
	unbondedSharesHeld := sdk.ZeroRat
	for _, candidate := range cA {
		// nonnegative ex rate
		assert.Equal(t, false, candidate.delegatorShareExRate().LT(sdk.ZeroRat))
		// nonnegative assets / liabilities
		assert.Equal(t, false, candidate.Assets.LT(sdk.ZeroRat))
		assert.Equal(t, false, candidate.Liabilities.LT(sdk.ZeroRat))
		if candidate.Status == Bonded {
			bondedSharesHeld = bondedSharesHeld.Add(candidate.Assets)
		} else {
			unbondedSharesHeld = unbondedSharesHeld.Add(candidate.Assets)
		}
	}
	// shares outstanding = total shares held by candidates, both bonded and unbonded
	assert.Equal(t, bondedSharesHeld, pB.BondedShares)
	assert.Equal(t, unbondedSharesHeld, pB.UnbondedShares)
}

// run random operations in a random order on a random state, assert invariants hold
func TestIntegrationInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(int64(42)))
	for i := 0; i < 10; i++ {
		pool, candidates, tokens := randomSetup(r)
		initialPool, initialCandidates, initialTokens := pool, candidates, tokens
		assertInvariants(t, initialPool, initialCandidates, initialTokens, pool, candidates, tokens)
		for j := 0; j < 100; j++ {
			pool, candidates, tokens = randomOperation(r)(pool, candidates, tokens)
			assertInvariants(t, initialPool, initialCandidates, initialTokens, pool, candidates, tokens)
		}
	}
}
