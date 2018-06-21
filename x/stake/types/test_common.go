package types

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
)

var (
	// dummy pubkeys/addresses
	pk1   = crypto.GenPrivKeyEd25519().PubKey()
	pk2   = crypto.GenPrivKeyEd25519().PubKey()
	pk3   = crypto.GenPrivKeyEd25519().PubKey()
	addr1 = pk1.Address()
	addr2 = pk2.Address()
	addr3 = pk3.Address()

	emptyAddr   sdk.Address
	emptyPubkey crypto.PubKey
)

//______________________________________________________________

// any operation that transforms staking state
// takes in RNG instance, pool, validator
// returns updated pool, updated validator, delta tokens, descriptive message
type Operation func(r *rand.Rand, pool Pool, c Validator) (Pool, Validator, int64, string)

// operation: bond or unbond a validator depending on current status
func OpBondOrUnbond(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, int64, string) {
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
	return pool, val, 0, msg
}

// operation: add a random number of tokens to a validator
func OpAddTokens(r *rand.Rand, pool Pool, val Validator) (Pool, Validator, int64, string) {
	tokens := int64(r.Int31n(1000))
	msg := fmt.Sprintf("validator %s (status: %d, poolShares: %v, delShares: %v, DelegatorShareExRate: %v)",
		val.Owner, val.Status(), val.PoolShares.Bonded(), val.DelegatorShares, val.DelegatorShareExRate(pool))
	val, pool, _ = val.AddTokensFromDel(pool, tokens)
	msg = fmt.Sprintf("Added %d tokens to %s", tokens, msg)
	return pool, val, -1 * tokens, msg // tokens are removed so for accounting must be negative
}

// operation: remove a random number of shares from a validator
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

// pick a random staking operation
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

// ensure invariants that should always be true are true
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

	// nonnegative bonded shares
	require.False(t, pMod.BondedShares.LT(sdk.ZeroRat()),
		"Negative bonded shares - msg: %v\npOrig: %v\npMod: %v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// nonnegative unbonded shares
	require.False(t, pMod.UnbondedShares.LT(sdk.ZeroRat()),
		"Negative unbonded shares - msg: %v\npOrig: %v\npMod: %v\ntokens: %v\n",
		msg, pOrig, pMod, tokens)

	// nonnegative bonded ex rate
	require.False(t, pMod.BondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative BondedShareExRate: %d",
		msg, pMod.BondedShareExRate().Evaluate())

	// nonnegative unbonded ex rate
	require.False(t, pMod.UnbondedShareExRate().LT(sdk.ZeroRat()),
		"Applying operation \"%s\" resulted in negative UnbondedShareExRate: %d",
		msg, pMod.UnbondedShareExRate().Evaluate())

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
		Owner:           addr1,
		PubKey:          pk1,
		PoolShares:      pShares,
		DelegatorShares: delShares,
	}
}

// generate a random staking state
func RandomSetup(r *rand.Rand, numValidators int) (Pool, []Validator) {
	pool := InitialPool()
	pool.LooseTokens = 100000

	validators := make([]Validator, numValidators)
	for i := 0; i < numValidators; i++ {
		validator := randomValidator(r, i)
		if validator.Status() == sdk.Bonded {
			pool.BondedShares = pool.BondedShares.Add(validator.PoolShares.Bonded())
			pool.BondedTokens += validator.PoolShares.Bonded().Evaluate()
		} else if validator.Status() == sdk.Unbonded {
			pool.UnbondedShares = pool.UnbondedShares.Add(validator.PoolShares.Unbonded())
			pool.UnbondedTokens += validator.PoolShares.Unbonded().Evaluate()
		}
		validators[i] = validator
	}
	return pool, validators
}
