package stake

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// Number of different random initial states to simulate
	InitialStateCount = 10

	// Number of random operations to simulate in sequence
	FeeOperationCount = 1000
)

// Test paying transaction fee
func TestTransactionFeePayment(t *testing.T) {
}

// Test delegator withdrawing fees
func TestDelegatorWithdrawal(t *testing.T) {
}

// Test validator withdrawing fees
func TestValidatorWithdrawal(t *testing.T) {
}

// Test delegator redelegating
func TestRedelegation(t *testing.T) {
}

// Test delegator unbonding
func TestDelegatorUnbonding(t *testing.T) {
}

// Test validator unbonding
func TestValidatorUnbonding(t *testing.T) {
}

// Test delegator bonding
func TestDelegatorBonding(t *testing.T) {
}

// Test validator bonding
func TestValidatorBonding(t *testing.T) {
}

// Test validator changing commission rate
func TestCommissionChange(t *testing.T) {
}

// Run the full random testsuite
func TestAll(t *testing.T) {
	ops := []FeeOperation{OpTransactionFeePaid, OpDelegatorWithdrawal, OpValidatorWithdrawal, OpRedelegation, OpDelegatorUnbond,
		OpValidatorUnbond, OpDelegatorBond, OpValidatorBond, OpCommissionChange}
	RandomlyTestFeeOperations(t, ops)
}

// Run only withdrawals
func TestOnlyWithdrawals(t *testing.T) {
	ops := []FeeOperation{OpTransactionFeePaid, OpDelegatorWithdrawal, OpValidatorWithdrawal}
	RandomlyTestFeeOperations(t, ops)
}

// Randomly test a list of operations
func RandomlyTestFeeOperations(t *testing.T, ops []FeeOperation) {
	rng := SetupRNG()
	for i := 0; i < InitialStateCount; i++ {
		state := RandomState(rng)
		operation := RandomOp(rng, ops)
		newState, descr := operation(rng, state)
		AssertInvariants(t, state, newState, descr)
		state = newState
	}
}

// State covered by the random tests
type State struct {
	// Track validators & delegators
	// Track two separate fee allocations - one calculated by the stake module, one calculated naively by iteration
}

// An operation takes an RNG & a state, returns an updated state & a description for debugging
type FeeOperation func(r *rand.Rand, s State) (State, string)

// Assert invariants which should always be true
func AssertInvariants(t *testing.T, old State, new State, descr string) {
	// Assert separately calculated fee allocations are equal for all delegators & validators
	// May need to round rationals, see https://github.com/cosmos/cosmos-sdk/issues/753
	assert.Equal(t, true, true)
}

// Pretty-print the state
func PrintState(s State) string {
	return ""
}

// Generate a random state
func RandomState(r *rand.Rand) State {
	return State{}
}

// Pick a random operation
func RandomOp(r *rand.Rand, ops []FeeOperation) FeeOperation {
	r.Shuffle(len(ops), func(i, j int) {
		ops[i], ops[j] = ops[j], ops[i]
	})
	return ops[0]
}

// Setup RNG
func SetupRNG() *rand.Rand {
	r := rand.New(rand.NewSource(42))
	return r
}

// FeeOperation: transaction pays fees
func OpTransactionFeePaid(r *rand.Rand, s State) (State, string) {
	// pay a random transaction fee
	return s, ""
}

// FeeOperation: delegator withdraws fees
func OpDelegatorWithdrawal(r *rand.Rand, s State) (State, string) {
	// pick a random delegator, withdraw fees
	return s, ""
}

// FeeOperation: validator withdraws fees
func OpValidatorWithdrawal(r *rand.Rand, s State) (State, string) {
	// pick a random validator, withdraw fees
	return s, ""
}

// FeeOperation: delegator redelegates
func OpRedelegation(r *rand.Rand, s State) (State, string) {
	// pick a random delegator, redelegate to a random validator
	return s, ""
}

// FeeOperation: delegator unbonds
func OpDelegatorUnbond(r *rand.Rand, s State) (State, string) {
	// pick a random delegator, unbond randomly either all or a random amount
	return s, ""
}

// FeeOperation: validator unbonds
func OpValidatorUnbond(r *rand.Rand, s State) (State, string) {
	// pick a random validator, unbond randomly either all or a random amount
	return s, ""
}

// FeeOperation: delegator bonds
func OpDelegatorBond(r *rand.Rand, s State) (State, string) {
	// pick a random delegator, bond a random amount to a random validator
	return s, ""
}

// FeeOperation: validator bonds
func OpValidatorBond(r *rand.Rand, s State) (State, string) {
	// pick a random validator, bond a random amount
	return s, ""
}

// FeeOperation: validator changes commission rate
func OpCommissionChange(r *rand.Rand, s State) (State, string) {
	// pick a random validator, change to a random commission rate
	return s, ""
}
