package stake

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddTokensValidatorBonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	pool := keeper.GetPool(ctx)
	val := NewValidator(addrs[0], pks[0], Description{})
	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	val, pool, delShares := val.addTokensFromDel(pool, sdk.NewInt(10))

	assert.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	assert.Equal(t, sdk.OneRat(), pool.bondedShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondingShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PoolShares.Bonded()))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	pool := keeper.GetPool(ctx)
	val := NewValidator(addrs[0], pks[0], Description{})
	val, pool = val.UpdateStatus(pool, sdk.Unbonding)
	val, pool, delShares := val.addTokensFromDel(pool, sdk.NewInt(10))

	assert.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	assert.Equal(t, sdk.OneRat(), pool.bondedShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondingShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PoolShares.Unbonding()))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	pool := keeper.GetPool(ctx)
	val := NewValidator(addrs[0], pks[0], Description{})
	val, pool = val.UpdateStatus(pool, sdk.Unbonded)
	val, pool, delShares := val.addTokensFromDel(pool, sdk.NewInt(10))

	assert.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	assert.Equal(t, sdk.OneRat(), pool.bondedShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondingShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PoolShares.Unbonded()))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveShares(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))

	poolA := keeper.GetPool(ctx)
	valA := Validator{
		Owner:           addrs[0],
		PubKey:          pks[0],
		PoolShares:      NewBondedShares(sdk.NewRat(9)),
		DelegatorShares: sdk.NewRat(9),
	}
	poolA.BondedTokens = valA.PoolShares.Bonded().EvaluateInt()
	poolA.BondedShares = valA.PoolShares.Bonded()
	assert.Equal(t, valA.DelegatorShareExRate(poolA), sdk.OneRat())
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	valB, poolB, coinsB := valA.removeDelShares(poolA, sdk.NewRat(10))

	// coins were created
	assert.Equal(t, coinsB.Int64(), int64(10))
	// pool shares were removed
	assert.Equal(t, valB.PoolShares.Bonded(), valA.PoolShares.Bonded().Sub(sdk.NewRat(10).Mul(valA.DelegatorShareExRate(poolA))))
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedTokens.Add(poolB.BondedTokens).Add(coinsB), poolA.UnbondedTokens.Add(poolA.BondedTokens))

	// specific case from random tests
	poolShares := sdk.NewRat(5102)
	delShares := sdk.NewRat(115)
	val := Validator{
		Owner:           addrs[0],
		PubKey:          pks[0],
		PoolShares:      NewBondedShares(poolShares),
		DelegatorShares: delShares,
	}
	pool := Pool{
		BondedShares:      sdk.NewRat(248305),
		UnbondedShares:    sdk.NewRat(232147),
		BondedTokens:      sdk.NewInt(248305),
		UnbondedTokens:    sdk.NewInt(232147),
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	shares := sdk.NewRat(29)
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	msg = fmt.Sprintf("Removed %v shares from %s", shares, msg)
	_, newPool, tokens := val.removeDelShares(pool, shares)
	require.Equal(t,
		tokens.Add(newPool.UnbondedTokens).Add(newPool.BondedTokens),
		pool.BondedTokens.Add(pool.UnbondedTokens),
		"Tokens were not conserved: %s", msg)
}

func TestUpdateStatus(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, sdk.NewInt(0))
	pool := keeper.GetPool(ctx)

	val := NewValidator(addrs[0], pks[0], Description{})
	val, pool, _ = val.addTokensFromDel(pool, sdk.NewInt(100))
	assert.Equal(t, int64(0), val.PoolShares.Bonded().Evaluate())
	assert.Equal(t, int64(0), val.PoolShares.Unbonding().Evaluate())
	assert.Equal(t, int64(100), val.PoolShares.Unbonded().Evaluate())
	assert.Equal(t, int64(0), pool.BondedTokens.Int64())
	assert.Equal(t, int64(0), pool.UnbondingTokens.Int64())
	assert.Equal(t, int64(100), pool.UnbondedTokens.Int64())

	val, pool = val.UpdateStatus(pool, sdk.Unbonding)
	assert.Equal(t, int64(0), val.PoolShares.Bonded().Evaluate())
	assert.Equal(t, int64(100), val.PoolShares.Unbonding().Evaluate())
	assert.Equal(t, int64(0), val.PoolShares.Unbonded().Evaluate())
	assert.Equal(t, int64(0), pool.BondedTokens.Int64())
	assert.Equal(t, int64(100), pool.UnbondingTokens.Int64())
	assert.Equal(t, int64(0), pool.UnbondedTokens.Int64())

	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	assert.Equal(t, int64(100), val.PoolShares.Bonded().Evaluate())
	assert.Equal(t, int64(0), val.PoolShares.Unbonding().Evaluate())
	assert.Equal(t, int64(0), val.PoolShares.Unbonded().Evaluate())
	assert.Equal(t, int64(100), pool.BondedTokens.Int64())
	assert.Equal(t, int64(0), pool.UnbondingTokens.Int64())
	assert.Equal(t, int64(0), pool.UnbondedTokens.Int64())
}

//________________________________________________________________________________
// TODO refactor this random setup

// generate a random validator
func randomValidator(r *rand.Rand, i int) Validator {

	poolSharesAmt := sdk.NewRat(int64(r.Int31n(10000)))
	delShares := sdk.NewRat(int64(r.Int31n(10000)))

	var pShares PoolShares
	if r.Float64() < float64(0.5) {
		pShares = NewBondedShares(poolSharesAmt)
	} else {
		pShares = NewUnbondedShares(poolSharesAmt)
	}
	return Validator{
		Owner:           addrs[i],
		PubKey:          pks[i],
		PoolShares:      pShares,
		DelegatorShares: delShares,
	}
}

// generate a random staking state
func randomSetup(r *rand.Rand, numValidators int) (Pool, Validators) {
	pool := InitialPool()

	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := randomValidator(r, i)
		if validator.Status() == sdk.Bonded {
			pool.BondedShares = pool.BondedShares.Add(validator.PoolShares.Bonded())
			pool.BondedTokens = pool.BondedTokens.Add(validator.PoolShares.Bonded().EvaluateInt())
		} else if validator.Status() == sdk.Unbonded {
			pool.UnbondedShares = pool.UnbondedShares.Add(validator.PoolShares.Unbonded())
			pool.UnbondedTokens = pool.UnbondedTokens.Add(validator.PoolShares.Unbonded().EvaluateInt())
		}
		validators[i] = validator
	}
	return pool, validators
}

// any operation that transforms staking state
// takes in RNG instance, pool, validator
// returns updated pool, updated validator, delta tokens, descriptive message
type Operation func(r *rand.Rand, pool Pool, c Validator) (Pool, Validator, sdk.Int, string)

// operation: bond or unbond a validator depending on current status
func OpBondOrUnbond(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, sdk.Int, string) {
	var msg string
	var newStatus sdk.BondStatus
	if val.Status() == sdk.Bonded {
		msg = fmt.Sprintf("sdk.Unbonded previously bonded validator %s (poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
			val.Owner, val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
		newStatus = sdk.Unbonded

	} else if val.Status() == sdk.Unbonded {
		msg = fmt.Sprintf("sdk.Bonded previously unbonded validator %s (poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
			val.Owner, val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
		newStatus = sdk.Bonded
	}
	val, pool = val.UpdateStatus(pool, newStatus)
	return pool, val, sdk.ZeroInt(), msg
}

// operation: add a random number of tokens to a validator
func OpAddTokens(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, sdk.Int, string) {
	tokens := sdk.NewInt(int64(r.Int31n(1000)))
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	val, pool, _ = val.addTokensFromDel(pool, tokens)
	msg = fmt.Sprintf("Added %v tokens to %s", tokens, msg)
	return pool, val, tokens.Neg(), msg // tokens are removed so for accounting must be negative
}

// operation: remove a random number of shares from a validator
func OpRemoveShares(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, sdk.Int, string) {
	var shares sdk.Rat
	for {
		shares = sdk.NewRat(int64(r.Int31n(1000)))
		if shares.LT(val.DelegatorShares) {
			break
		}
	}

	msg := fmt.Sprintf("Removed %v shares from validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		shares, val.Owner, val.Status(), val.PoolShares, val.DelegatorShares, val.DelegatorShareExRate(pool))

	val, pool, tokens := val.removeDelShares(pool, shares)
	return pool, val, tokens, msg
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
	pOrig Pool, cOrig Validators, pMod Pool, vMods Validators, tokens sdk.Int) {

	// total tokens conserved
	require.Equal(t,
		pOrig.UnbondedTokens.Add(pOrig.BondedTokens),
		pMod.UnbondedTokens.Add(pMod.BondedTokens).Add(tokens),
		"Tokens not conserved - msg: %v\n, pOrig.PoolShares.Bonded(): %v, pOrig.PoolShares.Unbonded(): %v, pMod.PoolShares.Bonded(): %v, pMod.PoolShares.Unbonded(): %v, pOrig.UnbondedTokens: %v, pOrig.BondedTokens: %v, pMod.UnbondedTokens: %v, pMod.BondedTokens: %v, tokens: %v\n",
		msg,
		pOrig.BondedShares, pOrig.UnbondedShares,
		pMod.BondedShares, pMod.UnbondedShares,
		pOrig.UnbondedTokens, pOrig.BondedTokens,
		pMod.UnbondedTokens, pMod.BondedTokens, tokens)

	// nonnegative bonded shares
	require.False(t, pMod.BondedShares.LT(sdk.ZeroRat()),
		"Negative bonded shares - msg: %v\npOrig: %v\npMod: %v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// nonnegative unbonded shares
	require.False(t, pMod.UnbondedShares.LT(sdk.ZeroRat()),
		"Negative unbonded shares - msg: %v\npOrig: %v\npMod: %v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// nonnegative bonded ex rate
	require.False(t, pMod.bondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative bondedShareExRate: %d",
		msg, pMod.bondedShareExRate().Evaluate())

	// nonnegative unbonded ex rate
	require.False(t, pMod.unbondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative unbondedShareExRate: %d",
		msg, pMod.unbondedShareExRate().Evaluate())

	for _, vMod := range vMods {

		// nonnegative ex rate
		require.False(t, vMod.DelegatorShareExRate(pMod).LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShareExRate(): %v (validator.Owner: %s)",
			msg,
			vMod.DelegatorShareExRate(pMod),
			vMod.Owner,
		)

		// nonnegative poolShares
		require.False(t, vMod.PoolShares.Bonded().LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.PoolShares.Bonded(): %v (validator.DelegatorShares: %v, validator.DelegatorShareExRate: %v, validator.Owner: %s)",
			msg,
			vMod.PoolShares.Bonded(),
			vMod.DelegatorShares,
			vMod.DelegatorShareExRate(pMod),
			vMod.Owner,
		)

		// nonnegative delShares
		require.False(t, vMod.DelegatorShares.LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShares: %v (validator.PoolShares.Bonded(): %v, validator.DelegatorShareExRate: %v, validator.Owner: %s)",
			msg,
			vMod.DelegatorShares,
			vMod.PoolShares.Bonded(),
			vMod.DelegatorShareExRate(pMod),
			vMod.Owner,
		)

	}

}

func TestPossibleOverflow(t *testing.T) {
	poolShares := sdk.NewRat(2159)
	delShares := sdk.NewRat(391432570689183511).Quo(sdk.NewRat(40113011844664))
	val := Validator{
		Owner:           addrs[0],
		PubKey:          pks[0],
		PoolShares:      NewBondedShares(poolShares),
		DelegatorShares: delShares,
	}
	pool := Pool{
		BondedShares:      poolShares,
		UnbondedShares:    sdk.ZeroRat(),
		BondedTokens:      poolShares.EvaluateInt(),
		UnbondedTokens:    sdk.ZeroInt(),
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	tokens := sdk.NewInt(71)
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	newValidator, _, _ := val.addTokensFromDel(pool, tokens)

	msg = fmt.Sprintf("Added %v tokens to %s", tokens, msg)
	require.False(t, newValidator.DelegatorShareExRate(pool).LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative DelegatorShareExRate(): %v",
		msg, newValidator.DelegatorShareExRate(pool))
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
			poolOrig, validatorsOrig, sdk.ZeroInt())

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
			poolOrig, validatorsOrig, sdk.ZeroInt())

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
