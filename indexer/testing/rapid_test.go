package indexertesting

import (
	"testing"

	"pgregory.net/rapid"
)

func TestStateMachine_ExampleSchema(t *testing.T) {
	rapid.Check(t, AppStateMachineTest(AppSimulatorOptions{
		AppSchema: ExampleAppSchema,
	}, StateMachineTestOptions{}))
}

func TestStateMachine(t *testing.T) {
	rapid.Check(t, AppStateMachineTest(
		AppSimulatorOptions{},
		StateMachineTestOptions{}))
}
