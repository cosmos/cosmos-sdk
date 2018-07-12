package types

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestValidatorEqual(t *testing.T) {
	val1 := NewValidator(addr1, pk1, Description{})
	val2 := NewValidator(addr1, pk1, Description{})

	ok := val1.Equal(val2)
	require.True(t, ok)

	val2 = NewValidator(addr2, pk2, Description{})

	ok = val1.Equal(val2)
	require.False(t, ok)
}

func TestUpdateDescription(t *testing.T) {
	d1 := Description{
		Moniker:  doNotModifyDescVal,
		Identity: doNotModifyDescVal,
		Website:  doNotModifyDescVal,
		Details:  doNotModifyDescVal,
	}
	d2 := Description{
		Website: "https://validator.cosmos",
		Details: "Test validator",
	}

	d, err := d1.UpdateDescription(d2)
	require.Nil(t, err)
	require.Equal(t, d, d1)
}

func TestABCIValidator(t *testing.T) {
	val := NewValidator(addr1, pk1, Description{})

	abciVal := val.ABCIValidator()
	require.Equal(t, tmtypes.TM2PB.PubKey(val.PubKey), abciVal.PubKey)
	require.Equal(t, val.BondedTokens().RoundInt64(), abciVal.Power)
}

func TestABCIValidatorZero(t *testing.T) {
	val := NewValidator(addr1, pk1, Description{})

	abciVal := val.ABCIValidatorZero()
	require.Equal(t, tmtypes.TM2PB.PubKey(val.PubKey), abciVal.PubKey)
	require.Equal(t, int64(0), abciVal.Power)
}

func TestRemoveTokens(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)

	val := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewRat(100),
		DelegatorShares: sdk.NewRat(100),
	}

	pool.BondedTokens = val.BondedTokens()

	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	val, pool = val.RemoveTokens(pool, sdk.NewRat(10))
	require.Equal(t, int64(90), val.Tokens.RoundInt64())
	require.Equal(t, int64(90), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(20), pool.LooseTokens.RoundInt64())

	val, pool = val.UpdateStatus(pool, sdk.Unbonded)
	val, pool = val.RemoveTokens(pool, sdk.NewRat(10))
	require.Equal(t, int64(80), val.Tokens.RoundInt64())
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(30), pool.LooseTokens.RoundInt64())
}

func TestAddTokensValidatorBonded(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	val := NewValidator(addr1, pk1, Description{})
	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	val, pool, delShares := val.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.BondedTokens()))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	val := NewValidator(addr1, pk1, Description{})
	val, pool = val.UpdateStatus(pool, sdk.Unbonding)
	val, pool, delShares := val.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.Equal(t, sdk.Unbonding, val.Status)
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.Tokens))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	val := NewValidator(addr1, pk1, Description{})
	val, pool = val.UpdateStatus(pool, sdk.Unbonded)
	val, pool, delShares := val.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), val.DelegatorShareExRate(pool))

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.Equal(t, sdk.Unbonded, val.Status)
	assert.True(sdk.RatEq(t, sdk.NewRat(10), val.Tokens))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveDelShares(t *testing.T) {
	poolA := InitialPool()
	poolA.LooseTokens = sdk.NewRat(10)
	valA := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewRat(100),
		DelegatorShares: sdk.NewRat(100),
	}
	poolA.BondedTokens = valA.BondedTokens()
	require.Equal(t, valA.DelegatorShareExRate(poolA), sdk.OneRat())
	valB, poolB, coinsB := valA.RemoveDelShares(poolA, sdk.NewRat(10))

	// coins were created
	require.Equal(t, coinsB, int64(10))

	// pool shares were removed
	require.Equal(t, valB.BondedTokens(), valA.BondedTokens().Sub(sdk.NewRat(10).Mul(valA.DelegatorShareExRate(poolA))))

	// conservation of tokens
	require.True(sdk.RatEq(t,
		poolB.LooseTokens.Add(poolB.BondedTokens).Add(coinsB),
		poolA.LooseTokens.Add(poolA.BondedTokens)))

	// specific case from random tests
	poolTokens := sdk.NewRat(5102)
	delShares := sdk.NewRat(115)
	val := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}
	pool := Pool{
		BondedTokens:      sdk.NewRat(248305),
		LooseTokens:       sdk.NewRat(232147),
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	shares := sdk.NewRat(29)
	msg := fmt.Sprintf("validator %#v", val)
	msg = fmt.Sprintf("Removed %v shares from %s", shares, msg)
	_, newPool, tokens := val.RemoveDelShares(pool, shares)
	require.True(sdk.RatEq(t,
		tokens.Add(newPool.LooseTokens).Add(newPool.BondedTokens),
		pool.LooseTokens.Add(pool.BondedTokens)))
}

func TestUpdateStatus(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(100)

	val := NewValidator(addr1, pk1, Description{})
	val, pool, _ = val.AddTokensFromDel(pool, 100)
	require.Equal(t, sdk.Unbonded, val.Status)
	require.Equal(t, int64(100), val.Tokens.RoundInt64())
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(100), pool.LooseTokens.RoundInt64())

	val, pool = val.UpdateStatus(pool, sdk.Unbonding)
	require.Equal(t, sdk.Unbonding, val.Status)
	require.Equal(t, int64(100), val.Tokens.RoundInt64())
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(100), pool.LooseTokens.RoundInt64())

	val, pool = val.UpdateStatus(pool, sdk.Bonded)
	require.Equal(t, sdk.Bonded, val.Status)
	require.Equal(t, int64(100), val.Tokens.RoundInt64())
	require.Equal(t, int64(100), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(0), pool.LooseTokens.RoundInt64())
}

func TestPossibleOverflow(t *testing.T) {
	poolTokens := sdk.NewRat(2159)
	delShares := sdk.NewRat(391432570689183511).Quo(sdk.NewRat(40113011844664))
	val := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}
	pool := Pool{
		LooseTokens:       sdk.NewRat(100),
		BondedTokens:      poolTokens,
		InflationLastTime: 0,
		Inflation:         sdk.NewRat(7, 100),
	}
	tokens := int64(71)
	msg := fmt.Sprintf("validator %#v", val)
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
			poolOrig, validatorsOrig, sdk.ZeroRat())

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
			poolOrig, validatorsOrig, sdk.ZeroRat())

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

func TestHumanReadableString(t *testing.T) {
	val := NewValidator(addr1, pk1, Description{})

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := val.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}
