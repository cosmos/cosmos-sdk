package appdata

import (
	"fmt"
	"testing"
)

func TestAsyncListenerMux(t *testing.T) {

}

func TestAsyncListener(t *testing.T) {
	t.Run("call done", func(t *testing.T) {
		commitChan := make(chan error)
		doneChan := make(chan struct{})
		var calls []string
		listener := callCollector(1, func(name string, i int, _ Packet) {
			calls = append(calls, fmt.Sprintf("%s %d", name, i))
		})
		res := AsyncListener(listener, 16, commitChan, doneChan)

		callAllCallbacksOnces(t, res)

		err := <-commitChan
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		checkExpectedCallOrder(t, calls, []string{
			"InitializeModuleData 1",
			"StartBlock 1",
			"OnTx 1",
			"OnEvent 1",
			"OnKVPair 1",
			"OnObjectUpdate 1",
			"Commit 1",
		})

		calls = nil

		doneChan <- struct{}{}

		callAllCallbacksOnces(t, res)

		checkExpectedCallOrder(t, calls, nil)
	})

	t.Run("error", func(t *testing.T) {
		commitChan := make(chan error)
		doneChan := make(chan struct{})
		var calls []string
		listener := callCollector(1, func(name string, i int, _ Packet) {
			calls = append(calls, fmt.Sprintf("%s %d", name, i))
		})

		listener.OnKVPair = func(updates KVPairData) error {
			return fmt.Errorf("error")
		}

		res := AsyncListener(listener, 16, commitChan, doneChan)

		callAllCallbacksOnces(t, res)

		err := <-commitChan
		if err == nil || err.Error() != "error" {
			t.Fatalf("expected error, got %v", err)
		}

		checkExpectedCallOrder(t, calls, []string{
			"InitializeModuleData 1",
			"StartBlock 1",
			"OnTx 1",
			"OnEvent 1",
		})
	})
}
