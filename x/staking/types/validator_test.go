package types

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidatorTestEquivalent(t *testing.T) {
	val1 := newValidator(t, valAddr1, pk1)
	val2 := newValidator(t, valAddr1, pk1)
	require.Equal(t, val1.String(), val2.String())

	val2 = newValidator(t, valAddr2, pk2)
	require.NotEqual(t, val1.String(), val2.String())
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
	validator := newValidator(t, valAddr1, pk1)
	abciVal := validator.ABCIValidatorUpdate()
	consPk, err := validator.TmConsPubKey()
	require.NoError(t, err)
	pk, err := encoding.PubKeyToProto(consPk)
	require.NoError(t, err)
	require.Equal(t, pk, abciVal.PubKey)
	require.Equal(t, validator.BondedTokens().Int64(), abciVal.Power)
}

func TestABCIValidatorUpdateZero(t *testing.T) {
	validator := newValidator(t, valAddr1, pk1)
	abciVal := validator.ABCIValidatorUpdateZero()
	consPk, err := validator.TmConsPubKey()
	require.NoError(t, err)
	pk, err := encoding.PubKeyToProto(consPk)
	require.NoError(t, err)
	require.Equal(t, pk, abciVal.PubKey)
	require.Equal(t, int64(0), abciVal.Power)
}

func TestShareTokens(t *testing.T) {
	validator := mkValidator(100, sdk.NewDec(100))
	assert.True(sdk.DecEq(t, sdk.NewDec(50), validator.TokensFromShares(sdk.NewDec(50))))

	validator.Tokens = sdk.NewInt(50)
	assert.True(sdk.DecEq(t, sdk.NewDec(25), validator.TokensFromShares(sdk.NewDec(50))))
	assert.True(sdk.DecEq(t, sdk.NewDec(5), validator.TokensFromShares(sdk.NewDec(10))))
}

func TestRemoveTokens(t *testing.T) {
	validator := mkValidator(100, sdk.NewDec(100))

	// remove tokens and test check everything
	validator = validator.RemoveTokens(sdk.NewInt(10))
	require.Equal(t, int64(90), validator.Tokens.Int64())

	// update validator to from bonded -> unbonded
	validator = validator.UpdateStatus(Unbonded)
	require.Equal(t, Unbonded, validator.Status)

	validator = validator.RemoveTokens(sdk.NewInt(10))
	require.Panics(t, func() { validator.RemoveTokens(sdk.NewInt(-1)) })
	require.Panics(t, func() { validator.RemoveTokens(sdk.NewInt(100)) })
}

func TestAddTokensValidatorBonded(t *testing.T) {
	validator := newValidator(t, valAddr1, pk1)
	validator = validator.UpdateStatus(Bonded)
	validator, delShares := validator.AddTokensFromDel(sdk.NewInt(10))

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.BondedTokens()))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
}

func TestAddTokensValidatorUnbonding(t *testing.T) {
	validator := newValidator(t, valAddr1, pk1)
	validator = validator.UpdateStatus(Unbonding)
	validator, delShares := validator.AddTokensFromDel(sdk.NewInt(10))

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.Equal(t, Unbonding, validator.Status)
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
}

func TestAddTokensValidatorUnbonded(t *testing.T) {

	validator := newValidator(t, valAddr1, pk1)
	validator = validator.UpdateStatus(Unbonded)
	validator, delShares := validator.AddTokensFromDel(sdk.NewInt(10))

	assert.True(sdk.DecEq(t, sdk.NewDec(10), delShares))
	assert.Equal(t, Unbonded, validator.Status)
	assert.True(sdk.IntEq(t, sdk.NewInt(10), validator.Tokens))
	assert.True(sdk.DecEq(t, sdk.NewDec(10), validator.DelegatorShares))
}

// TODO refactor to make simpler like the AddToken tests above
func TestRemoveDelShares(t *testing.T) {
	valA := Validator{
		OperatorAddress: valAddr1.String(),
		ConsensusPubkey: pk1Any,
		Status:          Bonded,
		Tokens:          sdk.NewInt(100),
		DelegatorShares: sdk.NewDec(100),
	}

	// Remove delegator shares
	valB, coinsB := valA.RemoveDelShares(sdk.NewDec(10))
	require.Equal(t, int64(10), coinsB.Int64())
	require.Equal(t, int64(90), valB.DelegatorShares.RoundInt64())
	require.Equal(t, int64(90), valB.BondedTokens().Int64())

	// specific case from random tests
	validator := mkValidator(5102, sdk.NewDec(115))
	_, tokens := validator.RemoveDelShares(sdk.NewDec(29))

	require.True(sdk.IntEq(t, sdk.NewInt(1286), tokens))
}

func TestAddTokensFromDel(t *testing.T) {
	validator := newValidator(t, valAddr1, pk1)

	validator, shares := validator.AddTokensFromDel(sdk.NewInt(6))
	require.True(sdk.DecEq(t, sdk.NewDec(6), shares))
	require.True(sdk.DecEq(t, sdk.NewDec(6), validator.DelegatorShares))
	require.True(sdk.IntEq(t, sdk.NewInt(6), validator.Tokens))

	validator, shares = validator.AddTokensFromDel(sdk.NewInt(3))
	require.True(sdk.DecEq(t, sdk.NewDec(3), shares))
	require.True(sdk.DecEq(t, sdk.NewDec(9), validator.DelegatorShares))
	require.True(sdk.IntEq(t, sdk.NewInt(9), validator.Tokens))
}

func TestUpdateStatus(t *testing.T) {
	validator := newValidator(t, valAddr1, pk1)
	validator, _ = validator.AddTokensFromDel(sdk.NewInt(100))
	require.Equal(t, Unbonded, validator.Status)
	require.Equal(t, int64(100), validator.Tokens.Int64())

	// Unbonded to Bonded
	validator = validator.UpdateStatus(Bonded)
	require.Equal(t, Bonded, validator.Status)

	// Bonded to Unbonding
	validator = validator.UpdateStatus(Unbonding)
	require.Equal(t, Unbonding, validator.Status)

	// Unbonding to Bonded
	validator = validator.UpdateStatus(Bonded)
	require.Equal(t, Bonded, validator.Status)
}

func TestPossibleOverflow(t *testing.T) {
	delShares := sdk.NewDec(391432570689183511).Quo(sdk.NewDec(40113011844664))
	validator := mkValidator(2159, delShares)
	newValidator, _ := validator.AddTokensFromDel(sdk.NewInt(71))

	require.False(t, newValidator.DelegatorShares.IsNegative())
	require.False(t, newValidator.Tokens.IsNegative())
}

func TestValidatorMarshalUnmarshalJSON(t *testing.T) {
	validator := newValidator(t, valAddr1, pk1)
	js, err := legacy.Cdc.MarshalJSON(validator)
	require.NoError(t, err)
	require.NotEmpty(t, js)
	require.Contains(t, string(js), "\"consensus_pubkey\":{\"type\":\"cosmos/PubKeyEd25519\"")
	got := &Validator{}
	err = legacy.Cdc.UnmarshalJSON(js, got)
	assert.NoError(t, err)
	assert.True(t, validator.Equal(got))
}

func TestValidatorSetInitialCommission(t *testing.T) {
	val := newValidator(t, valAddr1, pk1)
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

// Check that sort will create deterministic ordering of validators
func TestValidatorsSortDeterminism(t *testing.T) {
	vals := make([]Validator, 10)
	sortedVals := make([]Validator, 10)

	// Create random validator slice
	for i := range vals {
		pk := ed25519.GenPrivKey().PubKey()
		vals[i] = newValidator(t, sdk.ValAddress(pk.Address()), pk)
	}

	// Save sorted copy
	sort.Sort(Validators(vals))
	copy(sortedVals, vals)

	// Randomly shuffle validators, sort, and check it is equal to original sort
	for i := 0; i < 10; i++ {
		rand.Shuffle(10, func(i, j int) {
			it := vals[i]
			vals[i] = vals[j]
			vals[j] = it
		})

		Validators(vals).Sort()
		require.Equal(t, sortedVals, vals, "Validator sort returned different slices")
	}
}

func TestValidatorToTm(t *testing.T) {
	vals := make(Validators, 10)
	expected := make([]*tmtypes.Validator, 10)

	for i := range vals {
		pk := ed25519.GenPrivKey().PubKey()
		val := newValidator(t, sdk.ValAddress(pk.Address()), pk)
		val.Status = Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		vals[i] = val
		expected[i] = tmtypes.NewValidator(pk.(cryptotypes.IntoTmPubKey).AsTmPubKey(), val.ConsensusPower())
	}
	vs, err := vals.ToTmValidators()
	require.NoError(t, err)
	require.Equal(t, expected, vs)
}

func TestBondStatus(t *testing.T) {
	require.False(t, Unbonded == Bonded)
	require.False(t, Unbonded == Unbonding)
	require.False(t, Bonded == Unbonding)
	require.Equal(t, BondStatus(4).String(), "4")
	require.Equal(t, BondStatusUnspecified, Unspecified.String())
	require.Equal(t, BondStatusUnbonded, Unbonded.String())
	require.Equal(t, BondStatusBonded, Bonded.String())
	require.Equal(t, BondStatusUnbonding, Unbonding.String())
}

func mkValidator(tokens int64, shares sdk.Dec) Validator {
	return Validator{
		OperatorAddress: valAddr1.String(),
		ConsensusPubkey: pk1Any,
		Status:          Bonded,
		Tokens:          sdk.NewInt(tokens),
		DelegatorShares: shares,
	}
}

// Creates a new validators and asserts the error check.
func newValidator(t *testing.T, operator sdk.ValAddress, pubKey crypto.PubKey) Validator {
	v, err := NewValidator(operator, pubKey, Description{})
	require.NoError(t, err)
	return v
}
