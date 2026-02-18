package blockstm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchedulerNextTask_AdvancesValidationCursorWhenLagging(t *testing.T) {
	s := NewScheduler(3)

	// Simulate validation lagging behind execution where txn 0 has not executed yet.
	s.executionIdx.Store(1)

	version, kind := s.NextTask()
	require.Equal(t, TaskKindExecution, kind)
	require.Equal(t, TxnIndex(1), version.Index)
	// Validation cursor should still advance when validation lags execution.
	require.Equal(t, uint64(1), s.validationIdx.Load())
}

func TestSchedulerNextTask_ReturnsValidationTaskForExecutedTxn(t *testing.T) {
	s := NewScheduler(2)

	s.executionIdx.Store(1)
	_, ok := s.txnStatus[0].TrySetExecuting()
	require.True(t, ok)
	s.txnStatus[0].SetExecuted()

	version, kind := s.NextTask()
	require.Equal(t, TaskKindValidation, kind)
	require.Equal(t, TxnVersion{Index: 0, Incarnation: 0}, version)
	require.Equal(t, uint64(1), s.validationIdx.Load())
}
