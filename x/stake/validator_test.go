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
	ctx, _, keeper := createTestInput(t, false, 0)

	pool := keeper.GetPool(ctx)
	val := NewValidator(addrs[0], pks[0], Description{})
	val.Status = sdk.Bonded
	val, pool, delShares := val.addTokensFromDel(pool, 10)

	assert.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	assert.Equal(t, sdk.OneRat(), pool.bondedShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondingShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PShares.Bonded()))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	pool := keeper.GetPool(ctx)
	val := NewValidator(addrs[0], pks[0], Description{})
	val.Status = sdk.Unbonding
	val, pool, delShares := val.addTokensFromDel(pool, 10)

	assert.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	assert.Equal(t, sdk.OneRat(), pool.bondedShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondingShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PShares.Unbonding()))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	pool := keeper.GetPool(ctx)
	val := NewValidator(addrs[0], pks[0], Description{})
	val.Status = sdk.Unbonded
	val, pool, delShares := val.addTokensFromDel(pool, 10)

	assert.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	assert.Equal(t, sdk.OneRat(), pool.bondedShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondingShareExRate())
	assert.Equal(t, sdk.OneRat(), pool.unbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PShares.Unbonded()))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveShares(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)

	poolA := keeper.GetPool(ctx)
	valA := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		PShares:         NewBondedShares(sdk.NewRat(9)),
		DelegatorShares: sdk.NewRat(9),
	}
	poolA.BondedTokens = valA.PShares.Bonded().Evaluate()
	poolA.BondedShares = valA.PShares.Bonded()
	assert.Equal(t, valA.DelegatorShareExRate(poolA), sdk.OneRat())
	assert.Equal(t, poolA.bondedShareExRate(), sdk.OneRat())
	assert.Equal(t, poolA.unbondedShareExRate(), sdk.OneRat())
	valB, poolB, coinsB := valA.removeDelShares(poolA, sdk.NewRat(10))

	// coins were created
	assert.Equal(t, coinsB, int64(10))
	// pool shares were removed
	assert.Equal(t, valB.PShares.Bonded(), valA.PShares.Bonded().Sub(sdk.NewRat(10).Mul(valA.DelegatorShareExRate(poolA))))
	// conservation of tokens
	assert.Equal(t, poolB.UnbondedTokens+poolB.BondedTokens+coinsB, poolA.UnbondedTokens+poolA.BondedTokens)

	// specific case from random tests
	assets := sdk.NewRat(5102)
	liabilities := sdk.NewRat(115)
	val := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		PShares:         NewBondedShares(assets),
		DelegatorShares: liabilities,
	}
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      sdk.NewRat(248305),
		UnbondedShares:    sdk.NewRat(232147),
		BondedTokens:      248305,
		UnbondedTokens:    232147,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	shares := sdk.NewRat(29)
	msg := fmt.Sprintf("validator %s (status: %d, assets: %v, liabilities: %v, DelegatorShareExRate: %v)",
		val.Address, val.Status, val.PShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	msg = fmt.Sprintf("Removed %v shares from %s", shares, msg)
	_, newPool, tokens := val.removeDelShares(pool, shares)
	require.Equal(t,
		tokens+newPool.UnbondedTokens+newPool.BondedTokens,
		pool.BondedTokens+pool.UnbondedTokens,
		"Tokens were not conserved: %s", msg)
}

func TestUpdateSharesLocation(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 0)
	pool := keeper.GetPool(ctx)

	val := NewValidator(addrs[0], pks[0], Description{})
	val.Status = sdk.Unbonded
	val, pool, _ = val.addTokensFromDel(pool, 100)
	assert.Equal(t, int64(0), val.PShares.Bonded().Evaluate())
	assert.Equal(t, int64(0), val.PShares.Unbonding().Evaluate())
	assert.Equal(t, int64(100), val.PShares.Unbonded().Evaluate())
	assert.Equal(t, int64(0), pool.BondedTokens)
	assert.Equal(t, int64(0), pool.UnbondingTokens)
	assert.Equal(t, int64(100), pool.UnbondedTokens)

	val.Status = sdk.Unbonding
	val, pool = val.UpdateSharesLocation(pool)
	//require.Fail(t, "", "%v", val.PShares.Bonded().IsZero())
	assert.Equal(t, int64(0), val.PShares.Bonded().Evaluate())
	assert.Equal(t, int64(100), val.PShares.Unbonding().Evaluate())
	assert.Equal(t, int64(0), val.PShares.Unbonded().Evaluate())
	assert.Equal(t, int64(0), pool.BondedTokens)
	assert.Equal(t, int64(100), pool.UnbondingTokens)
	assert.Equal(t, int64(0), pool.UnbondedTokens)

	val.Status = sdk.Bonded
	val, pool = val.UpdateSharesLocation(pool)
	assert.Equal(t, int64(100), val.PShares.Bonded().Evaluate())
	assert.Equal(t, int64(0), val.PShares.Unbonding().Evaluate())
	assert.Equal(t, int64(0), val.PShares.Unbonded().Evaluate())
	assert.Equal(t, int64(100), pool.BondedTokens)
	assert.Equal(t, int64(0), pool.UnbondingTokens)
	assert.Equal(t, int64(0), pool.UnbondedTokens)
}

//________________________________________________________________________________
// TODO refactor this random setup

// generate a random validator
func randomValidator(r *rand.Rand) Validator {
	var status sdk.BondStatus
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
		PShares:         NewBondedShares(assets),
		DelegatorShares: liabilities,
	}
}

// generate a random staking state
func randomSetup(r *rand.Rand, numValidators int) (Pool, Validators) {
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      sdk.ZeroRat(),
		UnbondedShares:    sdk.ZeroRat(),
		BondedTokens:      0,
		UnbondedTokens:    0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}

	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := randomValidator(r)
		if validator.Status == sdk.Bonded {
			pool.BondedShares = pool.BondedShares.Add(validator.PShares.Bonded())
			pool.BondedTokens += validator.PShares.Bonded().Evaluate()
		} else if validator.Status == sdk.Unbonded {
			pool.UnbondedShares = pool.UnbondedShares.Add(validator.PShares.Bonded())
			pool.UnbondedTokens += validator.PShares.Bonded().Evaluate()
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
func OpBondOrUnbond(r *rand.Rand, p Pool, val Validator) (Pool, Validator, int64, string) {
	var msg string
	if val.Status == sdk.Bonded {
		msg = fmt.Sprintf("sdk.Unbonded previously bonded validator %s (assets: %v, liabilities: %v, DelegatorShareExRate: %v)",
			val.Address, val.PShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(p))
		val.Status = sdk.Unbonded
	} else if val.Status == sdk.Unbonded {
		msg = fmt.Sprintf("sdk.Bonded previously unbonded validator %s (assets: %v, liabilities: %v, DelegatorShareExRate: %v)",
			val.Address, val.PShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(p))
		val.Status = sdk.Bonded
	}
	val, p = val.UpdateSharesLocation(p)
	return p, val, 0, msg
}

// operation: add a random number of tokens to a validator
func OpAddTokens(r *rand.Rand, p Pool, val Validator) (Pool, Validator, int64, string) {
	tokens := int64(r.Int31n(1000))
	msg := fmt.Sprintf("validator %s (status: %d, assets: %v, liabilities: %v, DelegatorShareExRate: %v)",
		val.Address, val.Status, val.PShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(p))
	val, p, _ = val.addTokensFromDel(p, tokens)
	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	return p, val, -1 * tokens, msg // tokens are removed so for accounting must be negative
}

// operation: remove a random number of shares from a validator
func OpRemoveShares(r *rand.Rand, p Pool, val Validator) (Pool, Validator, int64, string) {
	var shares sdk.Rat
	for {
		shares = sdk.NewRat(int64(r.Int31n(1000)))
		if shares.LT(val.DelegatorShares) {
			break
		}
	}

	msg := fmt.Sprintf("Removed %v shares from validator %s (status: %d, assets: %v, liabilities: %v, DelegatorShareExRate: %v)",
		shares, val.Address, val.Status, val.PShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(p))

	val, p, tokens := val.removeDelShares(p, shares)
	return p, val, tokens, msg
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
	pOrig Pool, cOrig Validators, pMod Pool, vMods Validators, tokens int64) {

	// total tokens conserved
	require.Equal(t,
		pOrig.UnbondedTokens+pOrig.BondedTokens,
		pMod.UnbondedTokens+pMod.BondedTokens+tokens,
		"Tokens not conserved - msg: %v\n, pOrig.PShares.Bonded(): %v, pOrig.PShares.Unbonded(): %v, pMod.PShares.Bonded(): %v, pMod.PShares.Unbonded(): %v, pOrig.UnbondedTokens: %v, pOrig.BondedTokens: %v, pMod.UnbondedTokens: %v, pMod.BondedTokens: %v, tokens: %v\n",
		msg,
		pOrig.BondedShares, pOrig.UnbondedShares,
		pMod.BondedShares, pMod.UnbondedShares,
		pOrig.UnbondedTokens, pOrig.BondedTokens,
		pMod.UnbondedTokens, pMod.BondedTokens, tokens)

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

	for _, vMod := range vMods {

		// nonnegative ex rate
		require.False(t, vMod.DelegatorShareExRate(pMod).LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShareExRate(): %v (validator.Address: %s)",
			msg,
			vMod.DelegatorShareExRate(pMod),
			vMod.Address,
		)

		// nonnegative assets
		require.False(t, vMod.PShares.Bonded().LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.PShares.Bonded(): %v (validator.DelegatorShares: %v, validator.DelegatorShareExRate: %v, validator.Address: %s)",
			msg,
			vMod.PShares.Bonded(),
			vMod.DelegatorShares,
			vMod.DelegatorShareExRate(pMod),
			vMod.Address,
		)

		// nonnegative liabilities
		require.False(t, vMod.DelegatorShares.LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShares: %v (validator.PShares.Bonded(): %v, validator.DelegatorShareExRate: %v, validator.Address: %s)",
			msg,
			vMod.DelegatorShares,
			vMod.PShares.Bonded(),
			vMod.DelegatorShareExRate(pMod),
			vMod.Address,
		)

	}

}

func TestPossibleOverflow(t *testing.T) {
	assets := sdk.NewRat(2159)
	liabilities := sdk.NewRat(391432570689183511).Quo(sdk.NewRat(40113011844664))
	val := Validator{
		Status:          sdk.Bonded,
		Address:         addrs[0],
		PubKey:          pks[0],
		PShares:         NewBondedShares(assets),
		DelegatorShares: liabilities,
	}
	pool := Pool{
		TotalSupply:       0,
		BondedShares:      assets,
		UnbondedShares:    sdk.ZeroRat(),
		BondedTokens:      assets.Evaluate(),
		UnbondedTokens:    0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	tokens := int64(71)
	msg := fmt.Sprintf("validator %s (status: %d, assets: %v, liabilities: %v, DelegatorShareExRate: %v)",
		val.Address, val.Status, val.PShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	newValidator, _, _ := val.addTokensFromDel(pool, tokens)

	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
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
