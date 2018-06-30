package types

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAddTokensValidatorBonded(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = 10
	val := NewValidator(addr1, pk1, Description{})
	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	val, pool, delShares := val.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	require.Equal(t, sdk.OneRat(), pool.BondedShareExRate())
	require.Equal(t, sdk.OneRat(), pool.UnbondingShareExRate())
	require.Equal(t, sdk.OneRat(), pool.UnbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PoolShares.Bonded()))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = 10
	val := NewValidator(addr1, pk1, Description{})
	val, pool = val.UpdateStatus(pool, sdk.Unbonding)
	val, pool, delShares := val.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	require.Equal(t, sdk.OneRat(), pool.BondedShareExRate())
	require.Equal(t, sdk.OneRat(), pool.UnbondingShareExRate())
	require.Equal(t, sdk.OneRat(), pool.UnbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PoolShares.Unbonding()))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = 10
	val := NewValidator(addr1, pk1, Description{})
	val, pool = val.UpdateStatus(pool, sdk.Unbonded)
	val, pool, delShares := val.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))
	require.Equal(t, sdk.OneRat(), pool.BondedShareExRate())
	require.Equal(t, sdk.OneRat(), pool.UnbondingShareExRate())
	require.Equal(t, sdk.OneRat(), pool.UnbondedShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.PoolShares.Unbonded()))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveDelShares(t *testing.T) {
	poolA := InitialPool()
	poolA.LooseTokens = 10
	valA := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		PoolShares:      NewBondedShares(sdk.NewRat(100)),
		DelegatorShares: sdk.NewRat(100),
	}
	poolA.BondedTokens = valA.PoolShares.Bonded().Evaluate()
	poolA.BondedShares = valA.PoolShares.Bonded()
	require.Equal(t, valA.DelegatorShareExRate(poolA), sdk.OneRat())
	require.Equal(t, poolA.BondedShareExRate(), sdk.OneRat())
	require.Equal(t, poolA.UnbondedShareExRate(), sdk.OneRat())
	valB, poolB, coinsB := valA.RemoveDelShares(poolA, sdk.NewRat(10))

	// coins were created
	require.Equal(t, coinsB, int64(10))
	// pool shares were removed
	require.Equal(t, valB.PoolShares.Bonded(), valA.PoolShares.Bonded().Sub(sdk.NewRat(10).Mul(valA.DelegatorShareExRate(poolA))))
	// conservation of tokens
	require.Equal(t, poolB.UnbondedTokens+poolB.BondedTokens+coinsB, poolA.UnbondedTokens+poolA.BondedTokens)

	// specific case from random tests
	poolShares := sdk.NewRat(5102)
	delShares := sdk.NewRat(115)
	val := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		PoolShares:      NewBondedShares(poolShares),
		DelegatorShares: delShares,
	}
	pool := Pool{
		BondedShares:      sdk.NewRat(248305),
		UnbondedShares:    sdk.NewRat(232147),
		BondedTokens:      248305,
		UnbondedTokens:    232147,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	shares := sdk.NewRat(29)
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	msg = fmt.Sprintf("Removed %v shares from %s", shares, msg)
	_, newPool, tokens := val.RemoveDelShares(pool, shares)
	require.Equal(t,
		tokens+newPool.UnbondedTokens+newPool.BondedTokens,
		pool.BondedTokens+pool.UnbondedTokens,
		"Tokens were not conserved: %s", msg)
}

func TestUpdateStatus(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = 100

	val := NewValidator(addr1, pk1, Description{})
	val, pool, _ = val.AddTokensFromDel(pool, 100)
	require.Equal(t, int64(0), val.PoolShares.Bonded().Evaluate())
	require.Equal(t, int64(0), val.PoolShares.Unbonding().Evaluate())
	require.Equal(t, int64(100), val.PoolShares.Unbonded().Evaluate())
	require.Equal(t, int64(0), pool.BondedTokens)
	require.Equal(t, int64(0), pool.UnbondingTokens)
	require.Equal(t, int64(100), pool.UnbondedTokens)

	val, pool = val.UpdateStatus(pool, sdk.Unbonding)
	require.Equal(t, int64(0), val.PoolShares.Bonded().Evaluate())
	require.Equal(t, int64(100), val.PoolShares.Unbonding().Evaluate())
	require.Equal(t, int64(0), val.PoolShares.Unbonded().Evaluate())
	require.Equal(t, int64(0), pool.BondedTokens)
	require.Equal(t, int64(100), pool.UnbondingTokens)
	require.Equal(t, int64(0), pool.UnbondedTokens)

	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	require.Equal(t, int64(100), val.PoolShares.Bonded().Evaluate())
	require.Equal(t, int64(0), val.PoolShares.Unbonding().Evaluate())
	require.Equal(t, int64(0), val.PoolShares.Unbonded().Evaluate())
	require.Equal(t, int64(100), pool.BondedTokens)
	require.Equal(t, int64(0), pool.UnbondingTokens)
	require.Equal(t, int64(0), pool.UnbondedTokens)
}

func TestPossibleOverflow(t *testing.T) {
	poolShares := sdk.NewRat(2159)
	delShares := sdk.NewRat(391432570689183511).Quo(sdk.NewRat(40113011844664))
	val := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		PoolShares:      NewBondedShares(poolShares),
		DelegatorShares: delShares,
	}
	pool := Pool{
		LooseTokens:       100,
		BondedShares:      poolShares,
		UnbondedShares:    sdk.ZeroRat(),
		BondedTokens:      poolShares.Evaluate(),
		UnbondedTokens:    0,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	tokens := int64(71)
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	newValidator, _, _ := val.AddTokensFromDel(pool, tokens)

	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	require.False(t, newValidator.DelegatorShareExRate(pool).LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative DelegatorShareExRate(): %v",
		msg, newValidator.DelegatorShareExRate(pool))
}

// run random operations in a random order on a random single-validator state, assert invariants hold
func TestSingleValidatorIntegrationInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(41))

	for i := 0; i < 10; i++ {
		poolOrig, validatorsOrig := RandomSetup(r, 1)
		require.Equal(t, 1, len(validatorsOrig))

		// sanity check
		AssertInvariants(t, "no operation",
			poolOrig, validatorsOrig,
			poolOrig, validatorsOrig, 0)

		for j := 0; j < 5; j++ {
			poolMod, validatorMod, tokens, msg := RandomOperation(r)(r, poolOrig, validatorsOrig[0])

			validatorsMod := make([]Validator, len(validatorsOrig))
			copy(validatorsMod[:], validatorsOrig[:])
			require.Equal(t, 1, len(validatorsOrig), "j %v", j)
			require.Equal(t, 1, len(validatorsMod), "j %v", j)
			validatorsMod[0] = validatorMod

			AssertInvariants(t, msg,
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
		poolOrig, validatorsOrig := RandomSetup(r, 100)

		AssertInvariants(t, "no operation",
			poolOrig, validatorsOrig,
			poolOrig, validatorsOrig, 0)

		for j := 0; j < 5; j++ {
			index := int(r.Int31n(int32(len(validatorsOrig))))
			poolMod, validatorMod, tokens, msg := RandomOperation(r)(r, poolOrig, validatorsOrig[index])
			validatorsMod := make([]Validator, len(validatorsOrig))
			copy(validatorsMod[:], validatorsOrig[:])
			validatorsMod[index] = validatorMod

			AssertInvariants(t, msg,
				poolOrig, validatorsOrig,
				poolMod, validatorsMod, tokens)

			poolOrig = poolMod
			validatorsOrig = validatorsMod

		}
	}
}
