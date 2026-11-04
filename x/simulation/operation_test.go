package simulation

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestQueueOperationsTimeOps(t *testing.T) {
	noop := func(*rand.Rand, *baseapp.BaseApp, sdk.Context, []simtypes.Account, string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		return simtypes.NoOpMsg("test", "noop", ""), nil, nil
	}
	t0 := time.Unix(1_000, 0)

	queuedOps := make(OperationQueue)
	var queuedTimeOps []simtypes.FutureOperation
	queueOperations(queuedOps, &queuedTimeOps, []simtypes.FutureOperation{
		{BlockTime: t0.Add(2 * time.Second), Op: noop},
		{BlockHeight: 5, Op: noop},
		{BlockTime: t0.Add(1 * time.Second), Op: noop},
	})

	require.Len(t, queuedOps[5], 1)
	require.Len(t, queuedTimeOps, 2, "time-based future operations must reach the caller's queue")
	require.Equal(t, t0.Add(1*time.Second), queuedTimeOps[0].BlockTime, "time queue must stay sorted")
	require.Equal(t, t0.Add(2*time.Second), queuedTimeOps[1].BlockTime)
}

func TestRunQueuedTimeOperationsConsumesQueue(t *testing.T) {
	ran := 0
	op := func(*rand.Rand, *baseapp.BaseApp, sdk.Context, []simtypes.Account, string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		ran++
		return simtypes.NoOpMsg("test", "op", ""), nil, nil
	}
	t0 := time.Unix(1_000, 0)
	queue := []simtypes.FutureOperation{
		{BlockTime: t0.Add(1 * time.Second), Op: op},
		{BlockTime: t0.Add(10 * time.Second), Op: op},
	}
	r := rand.New(rand.NewSource(1))
	noEvent := func(route, op, evResult string) {}

	n, _ := runQueuedTimeOperations(t, &queue, 1, t0.Add(5*time.Second), r, nil, sdk.Context{}, nil, NewLogWriter(false), noEvent, true, "chain")
	require.Equal(t, 1, n)
	require.Len(t, queue, 1, "executed operations must be removed from the caller's queue")

	// Running again at the same time must not re-execute the consumed operation.
	n, _ = runQueuedTimeOperations(t, &queue, 2, t0.Add(5*time.Second), r, nil, sdk.Context{}, nil, NewLogWriter(false), noEvent, true, "chain")
	require.Equal(t, 0, n)
	require.Equal(t, 1, ran)
}
