package appdata

import (
	"fmt"
	"testing"
)

func TestAsyncListenerMux(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		listener := AsyncListenerMux([]Listener{{}, {}}, 16, make(chan struct{}))

		if listener.InitializeModuleData != nil {
			t.Error("expected nil")
		}
		if listener.StartBlock != nil {
			t.Error("expected nil")
		}
		if listener.OnTx != nil {
			t.Error("expected nil")
		}
		if listener.OnEvent != nil {
			t.Error("expected nil")
		}
		if listener.OnKVPair != nil {
			t.Error("expected nil")
		}
		if listener.OnObjectUpdate != nil {
			t.Error("expected nil")
		}

		// commit is not expected to be nil
	})

	t.Run("call done", func(t *testing.T) {
		doneChan := make(chan struct{})
		var calls1, calls2 []string
		listener1 := callCollector(1, func(name string, _ int, _ Packet) {
			calls1 = append(calls1, name)
		})
		listener2 := callCollector(2, func(name string, _ int, _ Packet) {
			calls2 = append(calls2, name)
		})
		res := AsyncListenerMux([]Listener{listener1, listener2}, 16, doneChan)

		callAllCallbacksOnces(t, res)

		expectedCalls := []string{
			"InitializeModuleData",
			"StartBlock",
			"OnTx",
			"OnEvent",
			"OnKVPair",
			"OnObjectUpdate",
			"Commit",
		}

		checkExpectedCallOrder(t, calls1, expectedCalls)
		checkExpectedCallOrder(t, calls2, expectedCalls)

		calls1 = nil
		calls2 = nil

		doneChan <- struct{}{}

		callAllCallbacksOnces(t, res)
		//
		//checkExpectedCallOrder(t, calls1, nil)
		//checkExpectedCallOrder(t, calls2, nil)
	})
}

func TestAsyncListener(t *testing.T) {
	t.Run("call done", func(t *testing.T) {
		commitChan := make(chan error)
		doneChan := make(chan struct{})
		var calls []string
		listener := callCollector(1, func(name string, _ int, _ Packet) {
			calls = append(calls, name)
		})
		res := AsyncListener(listener, 16, commitChan, doneChan)

		callAllCallbacksOnces(t, res)

		err := <-commitChan
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		checkExpectedCallOrder(t, calls, []string{
			"InitializeModuleData",
			"StartBlock",
			"OnTx",
			"OnEvent",
			"OnKVPair",
			"OnObjectUpdate",
			"Commit",
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
		listener := callCollector(1, func(name string, _ int, _ Packet) {
			calls = append(calls, name)
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

		checkExpectedCallOrder(t, calls, []string{"InitializeModuleData", "StartBlock", "OnTx", "OnEvent"})
	})
}
