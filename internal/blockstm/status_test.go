package blockstm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusEntry_SuspendResume(t *testing.T) {
	var se StatusEntry

	// Transition to Executing first.
	_, ok := se.TrySetExecuting()
	require.True(t, ok)

	cond := NewCondvar()
	se.Suspend(cond)
	require.Equal(t, StatusSuspended, se.status)

	// Resume should transition back to Executing, notify cond, and clear it.
	se.Resume()
	require.Equal(t, StatusExecuting, se.status)
	require.Nil(t, se.cond)

	// Verify cond was notified so Wait() would not block.
	require.True(t, cond.notified)
}

func TestStatusEntry_TryCancelThenResumeIsNoop(t *testing.T) {
	var se StatusEntry

	_, ok := se.TrySetExecuting()
	require.True(t, ok)

	cond := NewCondvar()
	se.Suspend(cond)
	require.Equal(t, StatusSuspended, se.status)

	// Cancel wakes the suspended entry.
	preCancelCalled := false
	se.TryCancel(func() { preCancelCalled = true })
	require.True(t, preCancelCalled)
	require.Equal(t, StatusExecuting, se.status)
	require.Nil(t, se.cond)
	require.True(t, cond.notified)

	// Resume after cancellation is a no-op (already executing, cond cleared).
	se.Resume()
	require.Equal(t, StatusExecuting, se.status)
}

func TestStatusEntry_ResumeWhenNotSuspendedIsNoop(t *testing.T) {
	var se StatusEntry

	// Initial status is ReadyToExecute — Resume should be a no-op.
	se.Resume()
	require.Equal(t, StatusReadyToExecute, se.status)

	// Transition to Executing and Resume again — still a no-op.
	_, ok := se.TrySetExecuting()
	require.True(t, ok)
	se.Resume()
	require.Equal(t, StatusExecuting, se.status)
}

func TestStatusEntry_StateTransitions(t *testing.T) {
	var se StatusEntry

	// ReadyToExecute -> Executing
	inc, ok := se.TrySetExecuting()
	require.True(t, ok)
	require.Equal(t, Incarnation(0), inc)
	require.Equal(t, StatusExecuting, se.status)

	// Executing -> Executed
	se.SetExecuted()
	require.Equal(t, StatusExecuted, se.status)

	// Executed -> Aborting
	ok = se.TryValidationAbort(0)
	require.True(t, ok)
	require.Equal(t, StatusAborting, se.status)

	// Aborting -> ReadyToExecute (incarnation incremented)
	se.SetReadyStatus()
	require.Equal(t, StatusReadyToExecute, se.status)
	require.Equal(t, Incarnation(1), se.incarnation)

	// ReadyToExecute -> Executing (second incarnation)
	inc, ok = se.TrySetExecuting()
	require.True(t, ok)
	require.Equal(t, Incarnation(1), inc)
}
