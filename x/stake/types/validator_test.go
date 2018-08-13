package types

import (
	"fmt"
	"testing"
	"time"

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
		Website: "https://validator.cosmos",
		Details: "Test validator",
	}

	d2 := Description{
		Moniker:  DoNotModifyDesc,
		Identity: DoNotModifyDesc,
		Website:  DoNotModifyDesc,
		Details:  DoNotModifyDesc,
	}

	d3 := Description{
		Moniker:  "",
		Identity: "",
		Website:  "",
		Details:  "",
	}

	d, err := d1.UpdateDescription(d2)
	require.Nil(t, err)
	require.Equal(t, d, d1)

	d, err = d1.UpdateDescription(d3)
	require.Nil(t, err)
	require.Equal(t, d, d3)
}

func TestABCIValidator(t *testing.T) {
	validator := NewValidator(addr1, pk1, Description{})

	abciVal := validator.ABCIValidator()
	require.Equal(t, tmtypes.TM2PB.PubKey(validator.PubKey), abciVal.PubKey)
	require.Equal(t, validator.BondedTokens().RoundInt64(), abciVal.Power)
}

func TestABCIValidatorZero(t *testing.T) {
	validator := NewValidator(addr1, pk1, Description{})

	abciVal := validator.ABCIValidatorZero()
	require.Equal(t, tmtypes.TM2PB.PubKey(validator.PubKey), abciVal.PubKey)
	require.Equal(t, int64(0), abciVal.Power)
}

func TestRemoveTokens(t *testing.T) {

	validator := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewRat(100),
		DelegatorShares: sdk.NewRat(100),
	}

	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	pool.BondedTokens = validator.BondedTokens()

	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)

	// remove tokens and test check everything
	validator, pool = validator.RemoveTokens(pool, sdk.NewRat(10))
	require.Equal(t, int64(90), validator.Tokens.RoundInt64())
	require.Equal(t, int64(90), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(20), pool.LooseTokens.RoundInt64())

	// update validator to unbonded and remove some more tokens
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(110), pool.LooseTokens.RoundInt64())

	validator, pool = validator.RemoveTokens(pool, sdk.NewRat(10))
	require.Equal(t, int64(80), validator.Tokens.RoundInt64())
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(110), pool.LooseTokens.RoundInt64())
}

func TestAddTokensValidatorBonded(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	validator := NewValidator(addr1, pk1, Description{})
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	validator, pool, delShares := validator.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), validator.DelegatorShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.BondedTokens()))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	validator := NewValidator(addr1, pk1, Description{})
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonding)
	validator, pool, delShares := validator.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), validator.DelegatorShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.Equal(t, sdk.Unbonding, validator.Status)
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.Tokens))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(10)
	validator := NewValidator(addr1, pk1, Description{})
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	validator, pool, delShares := validator.AddTokensFromDel(pool, 10)

	require.Equal(t, sdk.OneRat(), validator.DelegatorShareExRate())

	assert.True(sdk.RatEq(t, sdk.NewRat(10), delShares))
	assert.Equal(t, sdk.Unbonded, validator.Status)
	assert.True(sdk.RatEq(t, sdk.NewRat(10), validator.Tokens))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveDelShares(t *testing.T) {
	valA := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewRat(100),
		DelegatorShares: sdk.NewRat(100),
	}
	poolA := InitialPool()
	poolA.LooseTokens = sdk.NewRat(10)
	poolA.BondedTokens = valA.BondedTokens()
	require.Equal(t, valA.DelegatorShareExRate(), sdk.OneRat())

	// Remove delegator shares
	valB, poolB, coinsB := valA.RemoveDelShares(poolA, sdk.NewRat(10))
	assert.Equal(t, int64(10), coinsB.RoundInt64())
	assert.Equal(t, int64(90), valB.DelegatorShares.RoundInt64())
	assert.Equal(t, int64(90), valB.BondedTokens().RoundInt64())
	assert.Equal(t, int64(90), poolB.BondedTokens.RoundInt64())
	assert.Equal(t, int64(20), poolB.LooseTokens.RoundInt64())

	// conservation of tokens
	require.True(sdk.RatEq(t,
		poolB.LooseTokens.Add(poolB.BondedTokens),
		poolA.LooseTokens.Add(poolA.BondedTokens)))

	// specific case from random tests
	poolTokens := sdk.NewRat(5102)
	delShares := sdk.NewRat(115)
	validator := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}
	pool := Pool{
		BondedTokens:      sdk.NewRat(248305),
		LooseTokens:       sdk.NewRat(232147),
		InflationLastTime: time.Unix(0, 0),
		Inflation:         sdk.NewRat(7, 100),
	}
	shares := sdk.NewRat(29)
	_, newPool, tokens := validator.RemoveDelShares(pool, shares)
	require.True(sdk.RatEq(t, sdk.NewRat(147958, 115), tokens))
	require.True(sdk.RatEq(t,
		newPool.LooseTokens.Add(newPool.BondedTokens),
		pool.LooseTokens.Add(pool.BondedTokens)))
}

func TestUpdateStatus(t *testing.T) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(100)

	validator := NewValidator(addr1, pk1, Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, 100)
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64())
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(100), pool.LooseTokens.RoundInt64())

	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64())
	require.Equal(t, int64(100), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(0), pool.LooseTokens.RoundInt64())

	validator, pool = validator.UpdateStatus(pool, sdk.Unbonding)
	require.Equal(t, sdk.Unbonding, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.RoundInt64())
	require.Equal(t, int64(0), pool.BondedTokens.RoundInt64())
	require.Equal(t, int64(100), pool.LooseTokens.RoundInt64())
}

func TestPossibleOverflow(t *testing.T) {
	poolTokens := sdk.NewRat(2159)
	delShares := sdk.NewRat(391432570689183511).Quo(sdk.NewRat(40113011844664))
	validator := Validator{
		Owner:           addr1,
		PubKey:          pk1,
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}
	pool := Pool{
		LooseTokens:       sdk.NewRat(100),
		BondedTokens:      poolTokens,
		InflationLastTime: time.Unix(0, 0),
		Inflation:         sdk.NewRat(7, 100),
	}
	tokens := int64(71)
	msg := fmt.Sprintf("validator %#v", validator)
	newValidator, _, _ := validator.AddTokensFromDel(pool, tokens)

	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	require.False(t, newValidator.DelegatorShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative DelegatorShareExRate(): %v",
		msg, newValidator.DelegatorShareExRate())
}

func TestHumanReadableString(t *testing.T) {
	validator := NewValidator(addr1, pk1, Description{})

	// NOTE: Being that the validator's keypair is random, we cannot test the
	// actual contents of the string.
	valStr, err := validator.HumanReadableString()
	require.Nil(t, err)
	require.NotEmpty(t, valStr)
}
