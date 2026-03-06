package blockstm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Regression test: validation must not advance past a tx that has been
// re-incarnated but is currently executing (not yet EXECUTED).
func TestNextTaskDoesNotSkipValidationForExecutingReincarnation(t *testing.T) {
	s := NewScheduler(1)

	incarnation, ok := s.txnStatus[0].TrySetExecuting()
	require.True(t, ok)
	require.Equal(t, Incarnation(0), incarnation)

	s.txnStatus[0].SetExecuted()
	require.True(t, s.txnStatus[0].TryValidationAbort(incarnation))
	s.txnStatus[0].SetReadyStatus()

	incarnation, ok = s.txnStatus[0].TrySetExecuting()
	require.True(t, ok)
	require.Equal(t, Incarnation(1), incarnation)
	require.True(t, s.txnStatus[0].ExecutedOnce())
	require.Equal(t, uint64(0), s.validationIdx.Load())

	// Force NextTask to consider validation first.
	s.executionIdx.Store(1)

	version, kind := s.NextTask()

	require.Equal(t, TaskKindExecution, kind)
	require.False(t, version.Valid())
	require.Equal(t, uint64(0), s.validationIdx.Load(), "validation index advanced and skipped a tx that is not EXECUTED")
}
