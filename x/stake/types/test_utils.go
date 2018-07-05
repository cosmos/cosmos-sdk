package types

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

var (
	pk1   = crypto.GenPrivKeyEd25519().PubKey()
	pk2   = crypto.GenPrivKeyEd25519().PubKey()
	pk3   = crypto.GenPrivKeyEd25519().PubKey()
	addr1 = sdk.Address(pk1.Address())
	addr2 = sdk.Address(pk2.Address())
	addr3 = sdk.Address(pk3.Address())

	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)

// Operation reflects any operation that transforms staking state. It takes in
// a RNG instance, pool, validator and returns an updated pool, updated
// validator, delta tokens, and descriptive message.
type Operation func(r *rand.Rand, pool Pool, c Validator) (Pool, Validator, int64, string)

// OpBondOrUnbond implements an operation that bonds or unbonds a validator
// depending on current status.
// nolint: unparam
func OpBondOrUnbond(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, int64, string) {
	var (
		msg       string
		newStatus sdk.BondStatus
	)

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
	return pool, val, 0, msg
}

// OpAddTokens implements an operation that adds a random number of tokens to a
// validator.
func OpAddTokens(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, int64, string) {
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))

	tokens := int64(r.Int31n(1000))
	val, pool, _ = val.AddTokensFromDel(pool, tokens)
	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)

	// Tokens are removed so for accounting must be negative
	return pool, val, -1 * tokens, msg
}

// OpRemoveShares implements an operation that removes a random number of
// shares from a validator.
func OpRemoveShares(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, int64, string) {
	var shares sdk.Rat
	for {
		shares = sdk.NewRat(int64(r.Int31n(1000)))
		if shares.LT(val.DelegatorShares) {
			break
		}
	}

	msg := fmt.Sprintf("Removed %v shares from validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		shares, val.Owner, val.Status(), val.PoolShares, val.DelegatorShares, val.DelegatorShareExRate(pool))

	val, pool, tokens := val.RemoveDelShares(pool, shares)
	return pool, val, tokens, msg
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
	pOrig Pool, cOrig []Validator, pMod Pool, vMods []Validator, tokens int64) {

	// total tokens conserved
	require.Equal(t,
		pOrig.UnbondedTokens+pOrig.BondedTokens,
		pMod.UnbondedTokens+pMod.BondedTokens+tokens,
		"Tokens not conserved - msg: %v\n, pOrig.PoolShares.Bonded(): %v, pOrig.PoolShares.Unbonded(): %v, pMod.PoolShares.Bonded(): %v, pMod.PoolShares.Unbonded(): %v, pOrig.UnbondedTokens: %v, pOrig.BondedTokens: %v, pMod.UnbondedTokens: %v, pMod.BondedTokens: %v, tokens: %v\n",
		msg,
		pOrig.BondedShares, pOrig.UnbondedShares,
		pMod.BondedShares, pMod.UnbondedShares,
		pOrig.UnbondedTokens, pOrig.BondedTokens,
		pMod.UnbondedTokens, pMod.BondedTokens, tokens)

	// Nonnegative bonded shares
	require.False(t, pMod.BondedShares.LT(sdk.ZeroRat()),
		"Negative bonded shares - msg: %v\npOrig: %v\npMod: %v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// Nonnegative unbonded shares
	require.False(t, pMod.UnbondedShares.LT(sdk.ZeroRat()),
		"Negative unbonded shares - msg: %v\npOrig: %v\npMod: %v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// Nonnegative bonded ex rate
	require.False(t, pMod.BondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative BondedShareExRate: %d",
		msg, pMod.BondedShareExRate().RoundInt64())

	// Nonnegative unbonded ex rate
	require.False(t, pMod.UnbondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative UnbondedShareExRate: %d",
		msg, pMod.UnbondedShareExRate().RoundInt64())

	for _, vMod := range vMods {
		// Nonnegative ex rate
		require.False(t, vMod.DelegatorShareExRate(pMod).LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.DelegatorShareExRate(): %v (validator.Owner: %s)",
			msg,
			vMod.DelegatorShareExRate(pMod),
			vMod.Owner,
		)

		// Nonnegative poolShares
		require.False(t, vMod.PoolShares.Bonded().LT(sdk.ZeroRat()),
			"Applying operation \"%s\" resulted in negative validator.PoolShares.Bonded(): %v (validator.DelegatorShares: %v, validator.DelegatorShareExRate: %v, validator.Owner: %s)",
			msg,
			vMod.PoolShares.Bonded(),
			vMod.DelegatorShares,
			vMod.DelegatorShareExRate(pMod),
			vMod.Owner,
		)

		// Nonnegative delShares
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

// TODO: refactor this random setup

// randomValidator generates a random validator.
// nolint: unparam
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
		Owner:           addr1,
		PubKey:          pk1,
		PoolShares:      pShares,
		DelegatorShares: delShares,
	}
}

// RandomSetup generates a random staking state.
func RandomSetup(r *rand.Rand, numValidators int) (Pool, []Validator) {
	pool := InitialPool()
	pool.LooseTokens = 100000

	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := randomValidator(r, i)

		if validator.Status() == sdk.Bonded {
			pool.BondedShares = pool.BondedShares.Add(validator.PoolShares.Bonded())
			pool.BondedTokens += validator.PoolShares.Bonded().RoundInt64()
		} else if validator.Status() == sdk.Unbonded {
			pool.UnbondedShares = pool.UnbondedShares.Add(validator.PoolShares.Unbonded())
			pool.UnbondedTokens += validator.PoolShares.Unbonded().RoundInt64()
		}

		validators[i] = validator
	}

	return pool, validators
}
