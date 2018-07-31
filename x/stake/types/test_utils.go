package types

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	pk2   = ed25519.GenPrivKey().PubKey()
	pk3   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.AccAddress(pk1.Address())
	addr2 = sdk.AccAddress(pk2.Address())
	addr3 = sdk.AccAddress(pk3.Address())

	emptyAddr   sdk.AccAddress
	emptyPubkey crypto.PubKey
)

// Operation reflects any operation that transforms staking state. It takes in
// a RNG instance, pool, validator and returns an updated pool, updated
// validator, delta tokens, and descriptive message.
type Operation func(r *rand.Rand, pool Pool, c Validator) (Pool, Validator, sdk.Rat, string)

// OpBondOrUnbond implements an operation that bonds or unbonds a validator
// depending on current status.
// nolint: unparam
// TODO split up into multiple operations
func OpBondOrUnbond(r *rand.Rand, pool Pool, validator Validator) (Pool, Validator, sdk.Rat, string) {
	var (
		msg       string
		newStatus sdk.BondStatus
	)

	if validator.Status == sdk.Bonded {
		msg = fmt.Sprintf("sdk.Unbonded previously bonded validator %#v", validator)
		newStatus = sdk.Unbonded

	} else if validator.Status == sdk.Unbonded {
		msg = fmt.Sprintf("sdk.Bonded previously bonded validator %#v", validator)
		newStatus = sdk.Bonded
	}

	validator, pool = validator.UpdateStatus(pool, newStatus)
	return pool, validator, sdk.ZeroRat(), msg
}

// OpAddTokens implements an operation that adds a random number of tokens to a
// validator.
func OpAddTokens(r *rand.Rand, pool Pool, validator Validator) (Pool, Validator, sdk.Rat, string) {
	msg := fmt.Sprintf("validator %#v", validator)

	tokens := int64(r.Int31n(1000))
	validator, pool, _ = validator.AddTokensFromDel(pool, tokens)
	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)

	// Tokens are removed so for accounting must be negative
	return pool, validator, sdk.NewRat(-1 * tokens), msg
}

// OpRemoveShares implements an operation that removes a random number of
// delegatorshares from a validator.
func OpRemoveShares(r *rand.Rand, pool Pool, validator Validator) (Pool, Validator, sdk.Rat, string) {
	var shares sdk.Rat
	for {
		shares = sdk.NewRat(int64(r.Int31n(1000)))
		if shares.LT(validator.DelegatorShares) {
			break
		}
	}

	msg := fmt.Sprintf("Removed %v shares from validator %#v", shares, validator)

	validator, pool, tokens := validator.RemoveDelShares(pool, shares)
	return pool, validator, tokens, msg
}

// RandomOperation returns a random staking operation.
func RandomOperation(r *rand.Rand) Operation {
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

// AssertInvariants ensures invariants that should always be true are true.
// nolint: unparam
func AssertInvariants(t *testing.T, msg string,
	pOrig Pool, cOrig []Validator, pMod Pool, vMods []Validator) {

	// total tokens conserved
	require.True(t,
		pOrig.LooseTokens.Add(pOrig.BondedTokens).Equal(
			pMod.LooseTokens.Add(pMod.BondedTokens)),
		"Tokens not conserved - msg: %v\n, pOrig.BondedTokens: %v, pOrig.LooseTokens: %v,  pMod.BondedTokens: %v, pMod.LooseTokens: %v",
		msg,
		pOrig.BondedTokens, pOrig.LooseTokens,
		pMod.BondedTokens, pMod.LooseTokens)

	// Nonnegative bonded tokens
	require.False(t, pMod.BondedTokens.LT(sdk.ZeroRat()),
		"Negative bonded shares - msg: %v\npOrig: %v\npMod: %v\n",
		msg, pOrig, pMod)

	// Nonnegative loose tokens
	require.False(t, pMod.LooseTokens.LT(sdk.ZeroRat()),
		"Negative unbonded shares - msg: %v\npOrig: %v\npMod: %v\n",
		msg, pOrig, pMod)

	for _, vMod := range vMods {
		// Nonnegative ex rate
		require.False(t, vMod.DelegatorShareExRate().LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShareExRate(): %v (validator.Owner: %s)",
			msg,
			vMod.DelegatorShareExRate(),
			vMod.Owner,
		)

		// Nonnegative poolShares
		require.False(t, vMod.BondedTokens().LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.BondedTokens(): %#v",
			msg,
			vMod,
		)

		// Nonnegative delShares
		require.False(t, vMod.DelegatorShares.LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShares: %#v",
			msg,
			vMod,
		)
	}
}

// TODO: refactor this random setup

// randomValidator generates a random validator.
// nolint: unparam
func randomValidator(r *rand.Rand, i int) Validator {

	tokens := sdk.NewRat(int64(r.Int31n(10000)))
	delShares := sdk.NewRat(int64(r.Int31n(10000)))

	// TODO add more options here
	status := sdk.Bonded
	if r.Float64() > float64(0.5) {
		status = sdk.Unbonded
	}

	validator := NewValidator(addr1, pk1, Description{})
	validator.Status = status
	validator.Tokens = tokens
	validator.DelegatorShares = delShares

	return validator
}

// RandomSetup generates a random staking state.
func RandomSetup(r *rand.Rand, numValidators int) (Pool, []Validator) {
	pool := InitialPool()
	pool.LooseTokens = sdk.NewRat(100000)

	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := randomValidator(r, i)

		switch validator.Status {
		case sdk.Bonded:
			pool.BondedTokens = pool.BondedTokens.Add(validator.Tokens)
		case sdk.Unbonded, sdk.Unbonding:
			pool.LooseTokens = pool.LooseTokens.Add(validator.Tokens)
		default:
			panic("improper use of RandomSetup")
		}

		validators[i] = validator
	}

	return pool, validators
}
