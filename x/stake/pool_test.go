package stake

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBondedRatio(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.TotalSupply = 3
	pool.BondedPool = 2

	// bonded pool / total supply
	require.Equal(t, pool.bondedRatio(), sdk.NewRat(2).Quo(sdk.NewRat(3)))
	pool.TotalSupply = 0

	// avoids divide-by-zero
	require.Equal(t, pool.bondedRatio(), sdk.ZeroRat())
}

func TestBondedShareExRate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.BondedPool = 3
	pool.BondedShares = sdk.NewRat(10)

	// bonded pool / bonded shares
	require.Equal(t, pool.bondedShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.BondedShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.bondedShareExRate(), sdk.OneRat())
}

func TestUnbondedShareExRate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)
	pool.UnbondedPool = 3
	pool.UnbondedShares = sdk.NewRat(10)

	// unbonded pool / unbonded shares
	require.Equal(t, pool.unbondedShareExRate(), sdk.NewRat(3).Quo(sdk.NewRat(10)))
	pool.UnbondedShares = sdk.ZeroRat()

	// avoids divide-by-zero
	require.Equal(t, pool.unbondedShareExRate(), sdk.OneRat())
}

func TestBondedToUnbondedPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	candA := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    sdk.OneRat(),
		DelegatorShares: sdk.OneRat(),
	}
	poolB, candB := poolA.bondedToUnbondedPool(candA)

	// status unbonded
	assert.Equal(t, candB.Status, sdk.Unbonded)
	// same exchange rate, assets unchanged
	assert.Equal(t, candB.BondedShares, candA.BondedShares)
	// bonded pool decreased
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool-candA.BondedShares.Evaluate())
	// unbonded pool increased
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool+candA.BondedShares.Evaluate())
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool, poolA.BondedPool+poolA.UnbondedPool)
}

func TestUnbonbedtoBondedPool(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	candA := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    sdk.OneRat(),
		DelegatorShares: sdk.OneRat(),
	}
	candA.Status = sdk.Unbonded
	poolB, candB := poolA.unbondedToBondedPool(candA)

	// status bonded
	assert.Equal(t, candB.Status, sdk.Bonded)
	// same exchange rate, assets unchanged
	assert.Equal(t, candB.BondedShares, candA.BondedShares)
	// bonded pool increased
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool+candA.BondedShares.Evaluate())
	// unbonded pool decreased
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool-candA.BondedShares.Evaluate())
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool, poolA.BondedPool+poolA.UnbondedPool)
}

func TestAddTokensBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	poolB, sharesB := poolA.addTokensBonded(10)
	assert.Equal(t, poolB.bondedShareExRate(), sdk.OneRat())

	// correct changes to bonded shares and bonded pool
	assert.Equal(t, poolB.BondedShares, poolA.BondedShares.Add(sharesB))
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool+10)

	// same number of bonded shares / tokens when exchange rate is one
	assert.True(t, poolB.BondedShares.Equal(sdk.NewRat(poolB.BondedPool)))
}

func TestRemoveSharesBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	poolB, tokensB := poolA.removeSharesBonded(sdk.NewRat(10))
	assert.Equal(t, poolB.bondedShareExRate(), sdk.OneRat())

	// correct changes to bonded shares and bonded pool
	assert.Equal(t, poolB.BondedShares, poolA.BondedShares.Sub(sdk.NewRat(10)))
	assert.Equal(t, poolB.BondedPool, poolA.BondedPool-tokensB)

	// same number of bonded shares / tokens when exchange rate is one
	assert.True(t, poolB.BondedShares.Equal(sdk.NewRat(poolB.BondedPool)))
}

func TestAddTokensUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	poolB, sharesB := poolA.addTokensUnbonded(10)
	assert.Equal(t, poolB.unbondedShareExRate(), sdk.OneRat())

	// correct changes to unbonded shares and unbonded pool
	assert.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Add(sharesB))
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool+10)

	// same number of unbonded shares / tokens when exchange rate is one
	assert.True(t, poolB.UnbondedShares.Equal(sdk.NewRat(poolB.UnbondedPool)))
}

func TestRemoveSharesUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	poolB, tokensB := poolA.removeSharesUnbonded(sdk.NewRat(10))
	assert.Equal(t, poolB.unbondedShareExRate(), sdk.OneRat())

	// correct changes to unbonded shares and bonded pool
	assert.Equal(t, poolB.UnbondedShares, poolA.UnbondedShares.Sub(sdk.NewRat(10)))
	assert.Equal(t, poolB.UnbondedPool, poolA.UnbondedPool-tokensB)

	// same number of unbonded shares / tokens when exchange rate is one
	assert.True(t, poolB.UnbondedShares.Equal(sdk.NewRat(poolB.UnbondedPool)))
}

func TestValidatorAddTokens(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	candA := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    sdk.NewRat(9),
		DelegatorShares: sdk.NewRat(9),
	}
	poolA.BondedPool = candA.BondedShares.Evaluate()
	poolA.BondedShares = candA.BondedShares
	assert.Equal(t, candA.delegatorShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	poolB, candB, sharesB := poolA.validatorAddTokens(candA, 10)

	// shares were issued
	assert.Equal(t, sdk.NewRat(10).Mul(candA.delegatorShareExRate()), sharesB)
	// pool shares were added
	assert.Equal(t, candB.BondedShares, candA.BondedShares.Add(sdk.NewRat(10)))
	// conservation of tokens
	assert.Equal(t, poolB.BondedPool, 10+poolA.BondedPool)
}

func TestValidatorRemoveShares(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	candA := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    sdk.NewRat(9),
		DelegatorShares: sdk.NewRat(9),
	}
	poolA.BondedPool = candA.BondedShares.Evaluate()
	poolA.BondedShares = candA.BondedShares
	assert.Equal(t, candA.delegatorShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	poolB, candB, coinsB := poolA.validatorRemoveShares(candA, sdk.NewRat(10))

	// coins were created
	assert.Equal(t, coinsB, int64(10))
	// pool shares were removed
	assert.Equal(t, candB.BondedShares, candA.BondedShares.Sub(sdk.NewRat(10).Mul(candA.delegatorShareExRate())))
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedPool+poolB.BondedPool+coinsB, poolA.UnbondedPool+poolA.BondedPool)

	// specific case from random tests
	assets := sdk.NewRat(5102)
	liabilities := sdk.NewRat(115)
	cand := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    assets,
		DelegatorShares: liabilities,
	}
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      sdk.NewRat(248305),
		UnbondedShares:    sdk.NewRat(232147),
		BondedPool:        248305,
		UnbondedPool:      232147,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	shares := sdk.NewRat(29)
	msg := fmt.Sprintf("validator %s (status: %d, assets: %v, liabilities: %v, delegatorShareExRate: %v)",
		cand.Address, cand.Status, cand.BondedShares, cand.DelegatorShares, cand.delegatorShareExRate())
	msg = fmt.Sprintf("Removed %v shares from %s", shares, msg)
	newPool, _, tokens := pool.validatorRemoveShares(cand, shares)
	require.Equal(t,
		tokens+newPool.UnbondedPool+newPool.BondedPool,
		pool.BondedPool+pool.UnbondedPool,
		"Tokens were not conserved: %s", msg)
}

/////////////////////////////////////
// TODO Make all random tests less obfuscated!

// generate a random validator
func randomValidator(r *rand.Rand) Validator {
	var status sdk.ValidatorStatus
	if r.Float64() < float64(0.5) {
		status = sdk.Bonded
	} else {
		status = sdk.Unbonded
	}
	assets := sdk.NewRat(int64(r.Int31n(10000)))
	liabilities := sdk.NewRat(int64(r.Int31n(10000)))
	return Validator{
		Status:          status,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    assets,
		DelegatorShares: liabilities,
	}
}

// generate a random staking state
func randomSetup(r *rand.Rand, numValidators int) (Pool, Validators) {
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      sdk.ZeroRat(),
		UnbondedShares:    sdk.ZeroRat(),
		BondedPool:        0,
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}

	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := randomValidator(r)
		if validator.Status == sdk.Bonded {
			pool.BondedShares = pool.BondedShares.Add(validator.BondedShares)
			pool.BondedPool += validator.BondedShares.Evaluate()
		} else if validator.Status == sdk.Unbonded {
			pool.UnbondedShares = pool.UnbondedShares.Add(validator.BondedShares)
			pool.UnbondedPool += validator.BondedShares.Evaluate()
		}
		validators[i] = validator
	}
	return pool, validators
}

// any operation that transforms staking state
// takes in RNG instance, pool, validator
// returns updated pool, updated validator, delta tokens, descriptive message
type Operation func(r *rand.Rand, p Pool, c Validator) (Pool, Validator, int64, string)

// operation: bond or unbond a validator depending on current status
func OpBondOrUnbond(r *rand.Rand, p Pool, cand Validator) (Pool, Validator, int64, string) {
	var msg string
	if cand.Status == sdk.Bonded {
		msg = fmt.Sprintf("sdk.Unbonded previously bonded validator %s (assets: %v, liabilities: %v, delegatorShareExRate: %v)",
			cand.Address, cand.BondedShares, cand.DelegatorShares, cand.delegatorShareExRate())
		p, cand = p.bondedToUnbondedPool(cand)

	} else if cand.Status == sdk.Unbonded {
		msg = fmt.Sprintf("sdk.Bonded previously unbonded validator %s (assets: %v, liabilities: %v, delegatorShareExRate: %v)",
			cand.Address, cand.BondedShares, cand.DelegatorShares, cand.delegatorShareExRate())
		p, cand = p.unbondedToBondedPool(cand)
	}
	return p, cand, 0, msg
}

// operation: add a random number of tokens to a validator
func OpAddTokens(r *rand.Rand, p Pool, cand Validator) (Pool, Validator, int64, string) {
	tokens := int64(r.Int31n(1000))
	msg := fmt.Sprintf("validator %s (status: %d, assets: %v, liabilities: %v, delegatorShareExRate: %v)",
		cand.Address, cand.Status, cand.BondedShares, cand.DelegatorShares, cand.delegatorShareExRate())
	p, cand, _ = p.validatorAddTokens(cand, tokens)
	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	return p, cand, -1 * tokens, msg // tokens are removed so for accounting must be negative
}

// operation: remove a random number of shares from a validator
func OpRemoveShares(r *rand.Rand, p Pool, cand Validator) (Pool, Validator, int64, string) {
	var shares sdk.Rat
	for {
		shares = sdk.NewRat(int64(r.Int31n(1000)))
		if shares.LT(cand.DelegatorShares) {
			break
		}
	}

	msg := fmt.Sprintf("Removed %v shares from validator %s (status: %d, assets: %v, liabilities: %v, delegatorShareExRate: %v)",
		shares, cand.Address, cand.Status, cand.BondedShares, cand.DelegatorShares, cand.delegatorShareExRate())

	p, cand, tokens := p.validatorRemoveShares(cand, shares)
	return p, cand, tokens, msg
}

// pick a random staking operation
func randomOperation(r *rand.Rand) Operation {
	operations := []Operation{
		OpBondOrUnbond,
		OpAddTokens,
		OpRemoveShares,
	}
	r.Shuffle(len(operations), func(i, j int) {
		operations[i], operations[j] = operations[j], operations[i]
	})
	return operations[0]
}

// ensure invariants that should always be true are true
func assertInvariants(t *testing.T, msg string,
	pOrig Pool, cOrig Validators, pMod Pool, cMods Validators, tokens int64) {

	// total tokens conserved
	require.Equal(t,
		pOrig.UnbondedPool+pOrig.BondedPool,
		pMod.UnbondedPool+pMod.BondedPool+tokens,
		"Tokens not conserved - msg: %v\n, pOrig.BondedShares: %v, pOrig.UnbondedShares: %v, pMod.BondedShares: %v, pMod.UnbondedShares: %v, pOrig.UnbondedPool: %v, pOrig.BondedPool: %v, pMod.UnbondedPool: %v, pMod.BondedPool: %v, tokens: %v\n",
		msg,
		pOrig.BondedShares, pOrig.UnbondedShares,
		pMod.BondedShares, pMod.UnbondedShares,
		pOrig.UnbondedPool, pOrig.BondedPool,
		pMod.UnbondedPool, pMod.BondedPool, tokens)

	// nonnegative bonded shares
	require.False(t, pMod.BondedShares.LT(sdk.ZeroRat()),
		"Negative bonded shares - msg: %v\npOrig: %#v\npMod: %#v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// nonnegative unbonded shares
	require.False(t, pMod.UnbondedShares.LT(sdk.ZeroRat()),
		"Negative unbonded shares - msg: %v\npOrig: %#v\npMod: %#v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// nonnegative bonded ex rate
	require.False(t, pMod.bondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative bondedShareExRate: %d",
		msg, pMod.bondedShareExRate().Evaluate())

	// nonnegative unbonded ex rate
	require.False(t, pMod.unbondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative unbondedShareExRate: %d",
		msg, pMod.unbondedShareExRate().Evaluate())

	for _, cMod := range cMods {

		// nonnegative ex rate
		require.False(t, cMod.delegatorShareExRate().LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.delegatorShareExRate(): %v (validator.Address: %s)",
			msg,
			cMod.delegatorShareExRate(),
			cMod.Address,
		)

		// nonnegative assets
		require.False(t, cMod.BondedShares.LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.BondedShares: %v (validator.DelegatorShares: %v, validator.delegatorShareExRate: %v, validator.Address: %s)",
			msg,
			cMod.BondedShares,
			cMod.DelegatorShares,
			cMod.delegatorShareExRate(),
			cMod.Address,
		)

		// nonnegative liabilities
		require.False(t, cMod.DelegatorShares.LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShares: %v (validator.BondedShares: %v, validator.delegatorShareExRate: %v, validator.Address: %s)",
			msg,
			cMod.DelegatorShares,
			cMod.BondedShares,
			cMod.delegatorShareExRate(),
			cMod.Address,
		)

	}

}

func TestPossibleOverflow(t *testing.T) {
	assets := sdk.NewRat(2159)
	liabilities := sdk.NewRat(391432570689183511).Quo(sdk.NewRat(40113011844664))
	cand := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		BondedShares:    assets,
		DelegatorShares: liabilities,
	}
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      assets,
		UnbondedShares:    sdk.ZeroRat(),
		BondedPool:        assets.Evaluate(),
		UnbondedPool:      0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	tokens := int64(71)
	msg := fmt.Sprintf("validator %s (status: %d, assets: %v, liabilities: %v, delegatorShareExRate: %v)",
		cand.Address, cand.Status, cand.BondedShares, cand.DelegatorShares, cand.delegatorShareExRate())
	_, newValidator, _ := pool.validatorAddTokens(cand, tokens)

	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	require.False(t, newValidator.delegatorShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative delegatorShareExRate(): %v",
		msg, newValidator.delegatorShareExRate())
}

// run random operations in a random order on a random single-validator state, assert invariants hold
func TestSingleValidatorIntegrationInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(41))

	for i := 0; i < 10; i++ {
		poolOrig, validatorsOrig := randomSetup(r, 1)
		require.Equal(t, 1, len(validatorsOrig))

		// sanity check
		assertInvariants(t, "no operation",
			poolOrig, validatorsOrig,
			poolOrig, validatorsOrig, 0)

		for j := 0; j < 5; j++ {
			poolMod, validatorMod, tokens, msg := randomOperation(r)(r, poolOrig, validatorsOrig[0])

			validatorsMod := make([]Validator, len(validatorsOrig))
			copy(validatorsMod[:], validatorsOrig[:])
			require.Equal(t, 1, len(validatorsOrig), "j %v", j)
			require.Equal(t, 1, len(validatorsMod), "j %v", j)
			validatorsMod[0] = validatorMod

			assertInvariants(t, msg,
				poolOrig, validatorsOrig,
				poolMod, validatorsMod, tokens)

			poolOrig = poolMod
			validatorsOrig = validatorsMod
		}
	}
}

// run random operations in a random order on a random multi-validator state, assert invariants hold
func TestMultiValidatorIntegrationInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(42))

	for i := 0; i < 10; i++ {
		poolOrig, validatorsOrig := randomSetup(r, 100)

		assertInvariants(t, "no operation",
			poolOrig, validatorsOrig,
			poolOrig, validatorsOrig, 0)

		for j := 0; j < 5; j++ {
			index := int(r.Int31n(int32(len(validatorsOrig))))
			poolMod, validatorMod, tokens, msg := randomOperation(r)(r, poolOrig, validatorsOrig[index])
			validatorsMod := make([]Validator, len(validatorsOrig))
			copy(validatorsMod[:], validatorsOrig[:])
			validatorsMod[index] = validatorMod

			assertInvariants(t, msg,
				poolOrig, validatorsOrig,
				poolMod, validatorsMod, tokens)

			poolOrig = poolMod
			validatorsOrig = validatorsMod

		}
	}
}
