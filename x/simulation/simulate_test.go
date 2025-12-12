package simulation

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// Test that time-based FutureOperation values are enqueued into the shared
// timeOperationQueue and executed/cleared by runQueuedTimeOperations.
func TestTimeOperationQueue_EnqueueAndExecute(t *testing.T) {
	t.Helper()

	var (
		opCalled int
		now      = time.Now()
	)

	// Future operation scheduled in the past relative to currentTime so that it
	// must be executed by runQueuedTimeOperations.
	futureOp := simtypes.FutureOperation{
		BlockTime: now.Add(-1 * time.Second),
		Op: func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, chainID string,
		) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
			opCalled++
			return simtypes.NewOperationMsgBasic("test", "time-op", "", true, nil), nil, nil
		},
	}

	operationQueue := NewOperationQueue()
	timeOperationQueue := []simtypes.FutureOperation{}

	// Enqueue the time-based operation via queueOperations.
	queueOperations(operationQueue, &timeOperationQueue, []simtypes.FutureOperation{futureOp})

	if len(operationQueue) != 0 {
		t.Fatalf("expected height-based operationQueue to be empty, got %d entries", len(operationQueue))
	}
	if got := len(timeOperationQueue); got != 1 {
		t.Fatalf("expected timeOperationQueue length 1 after enqueue, got %d", got)
	}

	// Run queued time operations; the single past-due operation should execute
	// exactly once and be removed from the queue.
	r := rand.New(rand.NewSource(1))
	numRan, futureOps := runQueuedTimeOperations(
		t,
		&timeOperationQueue,
		1,   // height
		now, // currentTime
		r,
		nil,           // app
		sdk.Context{}, // ctx
		nil,           // accounts
		&DummyLogWriter{},
		func(_, _, _ string) {}, // event logger
		true,                    // lean
		"",                      // chainID
	)

	if numRan != 1 {
		t.Fatalf("expected numOpsRan = 1, got %d", numRan)
	}
	if opCalled != 1 {
		t.Fatalf("expected time-based operation to be called once, got %d", opCalled)
	}
	if len(futureOps) != 0 {
		t.Fatalf("expected no new future operations from time-based op, got %d", len(futureOps))
	}
	if got := len(timeOperationQueue); got != 0 {
		t.Fatalf("expected timeOperationQueue to be empty after execution, got %d", got)
	}
}
