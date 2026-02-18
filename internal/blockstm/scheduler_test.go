package blockstm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
