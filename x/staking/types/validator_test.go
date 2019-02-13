package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestValidatorTestEquivalent(t *testing.T) {
	val1 := NewValidator(addr1, pk1, Description{})
	val2 := NewValidator(addr1, pk1, Description{})

	ok := val1.TestEquivalent(val2)
	require.True(t, ok)

	val2 = NewValidator(addr2, pk2, Description{})

	ok = val1.TestEquivalent(val2)
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

func TestABCIValidatorUpdate(t *testing.T) {
	validator := NewValidator(addr1, pk1, Description{})

	abciVal := validator.ABCIValidatorUpdate()
	require.Equal(t, tmtypes.TM2PB.PubKey(validator.ConsPubKey), abciVal.PubKey)
	require.Equal(t, validator.BondedTokens().Int64(), abciVal.Power)
}

func TestABCIValidatorUpdateZero(t *testing.T) {
	validator := NewValidator(addr1, pk1, Description{})

	abciVal := validator.ABCIValidatorUpdateZero()
	require.Equal(t, tmtypes.TM2PB.PubKey(validator.ConsPubKey), abciVal.PubKey)
	require.Equal(t, int64(0), abciVal.Power)
}

func TestRemoveTokens(t *testing.T) {

	validator := Validator{
		OperatorAddr:    addr1,
		ConsPubKey:      pk1,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDec(100),
	}

	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(10)
	pool.BondedTokens = validator.BondedTokens()

	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)

	// remove tokens and test check everything
	validator, pool = validator.RemoveTokens(pool, sdk.NewInt(10))
	require.Equal(t, int64(90), validator.Tokens.Int64())
	require.Equal(t, int64(90), pool.BondedTokens.Int64())
	require.Equal(t, int64(20), pool.NotBondedTokens.Int64())

	// update validator to unbonded and remove some more tokens
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(0), pool.BondedTokens.Int64())
	require.Equal(t, int64(110), pool.NotBondedTokens.Int64())

	validator, pool = validator.RemoveTokens(pool, sdk.NewInt(10))
	require.Equal(t, int64(80), validator.Tokens.Int64())
	require.Equal(t, int64(0), pool.BondedTokens.Int64())
	require.Equal(t, int64(110), pool.NotBondedTokens.Int64())
}

func TestAddTokensValidatorBonded(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(10)
	validator := NewValidator(addr1, pk1, Description{})
	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	validator, pool, delShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))

	require.Equal(t, sdk.OneDec(), validator.DelegatorShareExRate())

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.BondedTokens()))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(10)
	validator := NewValidator(addr1, pk1, Description{})
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonding)
	validator, pool, delShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))

	require.Equal(t, sdk.OneDec(), validator.DelegatorShareExRate())

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.Equal(t, sdk.Unbonding, validator.Status)
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.Tokens))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(10)
	validator := NewValidator(addr1, pk1, Description{})
	validator, pool = validator.UpdateStatus(pool, sdk.Unbonded)
	validator, pool, delShares := validator.AddTokensFromDel(pool, sdk.NewInt(10))

	require.Equal(t, sdk.OneDec(), validator.DelegatorShareExRate())

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.Equal(t, sdk.Unbonded, validator.Status)
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.Tokens))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveDelShares(t *testing.T) {
	valA := Validator{
		OperatorAddr:    addr1,
		ConsPubKey:      pk1,
		Status:          sdk.Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDec(100),
	}
	poolA := InitialPool()
	poolA.NotBondedTokens = sdk.NewInt(10)
	poolA.BondedTokens = valA.BondedTokens()
	require.Equal(t, valA.DelegatorShareExRate(), sdk.OneDec())

	// Remove delegator shares
	valB, poolB, coinsB := valA.RemoveDelShares(poolA, sdk.NewDec(10))
	require.Equal(t, int64(10), coinsB.Int64())
	require.Equal(t, int64(90), valB.DelegatorShares.RoundInt64())
	require.Equal(t, int64(90), valB.BondedTokens().Int64())
	require.Equal(t, int64(90), poolB.BondedTokens.Int64())
	require.Equal(t, int64(20), poolB.NotBondedTokens.Int64())

	// conservation of tokens
	require.True(sdk.IntEq(t,
		poolB.NotBondedTokens.Add(poolB.BondedTokens),
		poolA.NotBondedTokens.Add(poolA.BondedTokens)))

	// specific case from random tests
	poolTokens := sdk.NewInt(5102)
	delShares := sdk.NewDec(115)
	validator := Validator{
		OperatorAddr:    addr1,
		ConsPubKey:      pk1,
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}
	pool := Pool{
		BondedTokens:    sdk.NewInt(248305),
		NotBondedTokens: sdk.NewInt(232147),
	}
	shares := sdk.NewDec(29)
	_, newPool, tokens := validator.RemoveDelShares(pool, shares)

	require.True(sdk.IntEq(t, sdk.NewInt(1286), tokens))

	require.True(sdk.IntEq(t,
		newPool.NotBondedTokens.Add(newPool.BondedTokens),
		pool.NotBondedTokens.Add(pool.BondedTokens)))
}

func TestUpdateStatus(t *testing.T) {
	pool := InitialPool()
	pool.NotBondedTokens = sdk.NewInt(100)

	validator := NewValidator(addr1, pk1, Description{})
	validator, pool, _ = validator.AddTokensFromDel(pool, sdk.NewInt(100))
	require.Equal(t, sdk.Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())
	require.Equal(t, int64(0), pool.BondedTokens.Int64())
	require.Equal(t, int64(100), pool.NotBondedTokens.Int64())

	validator, pool = validator.UpdateStatus(pool, sdk.Bonded)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())
	require.Equal(t, int64(100), pool.BondedTokens.Int64())
	require.Equal(t, int64(0), pool.NotBondedTokens.Int64())

	validator, pool = validator.UpdateStatus(pool, sdk.Unbonding)
	require.Equal(t, sdk.Unbonding, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())
	require.Equal(t, int64(0), pool.BondedTokens.Int64())
	require.Equal(t, int64(100), pool.NotBondedTokens.Int64())
}

func TestPossibleOverflow(t *testing.T) {
	poolTokens := sdk.NewInt(2159)
	delShares := sdk.NewDec(391432570689183511).Quo(sdk.NewDec(40113011844664))
	validator := Validator{
		OperatorAddr:    addr1,
		ConsPubKey:      pk1,
		Status:          sdk.Bonded,
		Tokens:          poolTokens,
		DelegatorShares: delShares,
	}
	pool := Pool{
		NotBondedTokens: sdk.NewInt(100),
		BondedTokens:    poolTokens,
	}
	tokens := int64(71)
	msg := fmt.Sprintf("validator %#v", validator)
	newValidator, _, _ := validator.AddTokensFromDel(pool, sdk.NewInt(tokens))

	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	require.False(t, newValidator.DelegatorShareExRate().LT(sdk.ZeroDec()),
		"Applying operation \"%s\" resulted in negative DelegatorShareExRate(): %v",
		msg, newValidator.DelegatorShareExRate())
}

func TestValidatorMarshalUnmarshalJSON(t *testing.T) {
	validator := NewValidator(addr1, pk1, Description{})
	js, err := codec.Cdc.MarshalJSON(validator)
	require.NoError(t, err)
	require.NotEmpty(t, js)
	require.Contains(t, string(js), "\"consensus_pubkey\":\"cosmosvalconspu")
	got := &Validator{}
	err = codec.Cdc.UnmarshalJSON(js, got)
	assert.NoError(t, err)
	assert.Equal(t, validator, *got)
}

func TestValidatorSetInitialCommission(t *testing.T) {
	val := NewValidator(addr1, pk1, Description{})
	testCases := []struct {
		validator   Validator
		commission  Commission
		expectedErr bool
	}{
		{val, NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()), false},
		{val, NewCommission(sdk.ZeroDec(), sdk.NewDecWithPrec(-1, 1), sdk.ZeroDec()), true},
		{val, NewCommission(sdk.ZeroDec(), sdk.NewDec(15000000000), sdk.ZeroDec()), true},
		{val, NewCommission(sdk.NewDecWithPrec(-1, 1), sdk.ZeroDec(), sdk.ZeroDec()), true},
		{val, NewCommission(sdk.NewDecWithPrec(2, 1), sdk.NewDecWithPrec(1, 1), sdk.ZeroDec()), true},
		{val, NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.NewDecWithPrec(-1, 1)), true},
		{val, NewCommission(sdk.ZeroDec(), sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(2, 1)), true},
	}

	for i, tc := range testCases {
		val, err := tc.validator.SetInitialCommission(tc.commission)

		if tc.expectedErr {
			require.Error(t, err,
				"expected error for test case #%d with commission: %s", i, tc.commission,
			)
		} else {
			require.NoError(t, err,
				"unexpected error for test case #%d with commission: %s", i, tc.commission,
			)
			require.Equal(t, tc.commission, val.Commission,
				"invalid validator commission for test case #%d with commission: %s", i, tc.commission,
			)
		}
	}
}
