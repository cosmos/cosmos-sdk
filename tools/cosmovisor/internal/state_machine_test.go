package internal

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/qmuntal/stateless"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStateMachineGraph(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRunner(ctrl)
	mock.EXPECT().CheckForUpgradeInfoJSON(gomock.Any()).Return(nil).AnyTimes()
	err := os.WriteFile("state_machine.dot", []byte(StateMachine(mock).ToGraph()), 0644)
	require.NoError(t, err, "Failed to write state machine graph to file")
	err = exec.Command("dot", "-Tpng", "state_machine.dot", "-o", "state_machine.png").Run()
	require.NoError(t, err, "Failed to generate PNG from state machine graph")
}

func TestStateMachine(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRunner(ctrl)
	fsm := StateMachine(mock)
	fsm.OnTransitioning(func(ctx context.Context, tx stateless.Transition) {
		t.Logf("Transitioning from %s to %s with trigger %s", tx.Source, tx.Destination, tx.Trigger)
	})
	fsm.OnTransitioned(func(ctx context.Context, tx stateless.Transition) {
		t.Logf("Transitioned from %s to %s with trigger %s", tx.Source, tx.Destination, tx.Trigger)
	})
	require.NoError(t, fsm.ActivateCtx(context.Background()))
	state, err := fsm.State(context.Background())
	require.NoError(t, err, "Failed to get initial state")
	t.Logf("Initial state: %s", state)
}

func TestFSM2(t *testing.T) {
	fsm := stateless.NewStateMachine("A")
	fsm.SetTriggerParameters("go", reflect.TypeOf(""), reflect.TypeOf(""))

	fsm.Configure("A").
		Permit("go", "B")

	fsm.Configure("B").
		OnEntry(func(ctx context.Context, args ...any) error {
			t.Log("Entering state B with args:", args)
			return nil
		})
	require.NoError(t, fsm.Fire("go", "arg1", "arg2"))
}
