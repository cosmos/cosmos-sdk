package simulation

import (
	"math/rand"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Operation runs a state machine transition, and ensures the transition
// happened as expected.  The operation could be running and testing a fuzzed
// transaction, or doing the same for a message.
//
// For ease of debugging, an operation returns a descriptive message "action",
// which details what this fuzzed state machine transition actually did.
//
// Operations can optionally provide a list of "FutureOperations" to run later
// These will be ran at the beginning of the corresponding block.
type Operation func(r *rand.Rand, app *baseapp.BaseApp,
	ctx sdk.Context, accounts []Account, event func(string)) (
	action string, futureOps []FutureOperation, err error)

// queue of operations
type OperationQueue map[int][]Operation

func newOperationQueue() OperationQueue {
	operationQueue := make(OperationQueue)
	return operationQueue
}

// adds all future operations into the operation queue.
func queueOperations(queuedOps OperationQueue,
	queuedTimeOps []FutureOperation, futureOps []FutureOperation) {

	if futureOps == nil {
		return
	}

	for _, futureOp := range futureOps {
		if futureOp.BlockHeight != 0 {
			if val, ok := queuedOps[futureOp.BlockHeight]; ok {
				queuedOps[futureOp.BlockHeight] = append(val, futureOp.Op)
			} else {
				queuedOps[futureOp.BlockHeight] = []Operation{futureOp.Op}
			}
			continue
		}

		// TODO: Replace with proper sorted data structure, so don't have the
		// copy entire slice
		index := sort.Search(
			len(queuedTimeOps),
			func(i int) bool {
				return queuedTimeOps[i].BlockTime.After(futureOp.BlockTime)
			},
		)
		queuedTimeOps = append(queuedTimeOps, FutureOperation{})
		copy(queuedTimeOps[index+1:], queuedTimeOps[index:])
		queuedTimeOps[index] = futureOp
	}
}

//________________________________________________________________________

// FutureOperation is an operation which will be ran at the beginning of the
// provided BlockHeight. If both a BlockHeight and BlockTime are specified, it
// will use the BlockHeight. In the (likely) event that multiple operations
// are queued at the same block height, they will execute in a FIFO pattern.
type FutureOperation struct {
	BlockHeight int
	BlockTime   time.Time
	Op          Operation
}

//________________________________________________________________________

// WeightedOperation is an operation with associated weight.
// This is used to bias the selection operation within the simulator.
type WeightedOperation struct {
	Weight int
	Op     Operation
}

// WeightedOperations is the group of all weighted operations to simulate.
type WeightedOperations []WeightedOperation

func (ops WeightedOperations) totalWeight() int {
	totalOpWeight := 0
	for _, op := range ops {
		totalOpWeight += op.Weight
	}
	return totalOpWeight
}

type selectOpFn func(r *rand.Rand) Operation

func (ops WeightedOperations) getSelectOpFn() selectOpFn {
	totalOpWeight := ops.totalWeight()
	return func(r *rand.Rand) Operation {
		x := r.Intn(totalOpWeight)
		for i := 0; i < len(ops); i++ {
			if x <= ops[i].Weight {
				return ops[i].Op
			}
			x -= ops[i].Weight
		}
		// shouldn't happen
		return ops[0].Op
	}
}
