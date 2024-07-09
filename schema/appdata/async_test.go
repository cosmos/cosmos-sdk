package appdata

import (
	"context"
	"fmt"
	"testing"
)

func TestAsyncListenerMux(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		listener := AsyncListenerMux([]Listener{{}, {}}, 16, context.Background())

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

	t.Run("call cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var calls1, calls2 []string
		listener1 := callCollector(1, func(name string, _ int, _ Packet) {
			calls1 = append(calls1, name)
		})
		listener2 := callCollector(2, func(name string, _ int, _ Packet) {
			calls2 = append(calls2, name)
		})
		res := AsyncListenerMux([]Listener{listener1, listener2}, 16, ctx)

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

		cancel()

		// expect a panic if we try to write to the now closed channels
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("expected panic")
			}
		}()
		callAllCallbacksOnces(t, res)
	})
}

func TestAsyncListener(t *testing.T) {
	t.Run("call cancel", func(t *testing.T) {
		commitChan := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		var calls []string
		listener := callCollector(1, func(name string, _ int, _ Packet) {
			calls = append(calls, name)
		})
		res := AsyncListener(listener, 16, commitChan, ctx)

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

		cancel()

		callAllCallbacksOnces(t, res)

		checkExpectedCallOrder(t, calls, nil)
	})

	t.Run("error", func(t *testing.T) {
		commitChan := make(chan error)
		var calls []string
		listener := callCollector(1, func(name string, _ int, _ Packet) {
			calls = append(calls, name)
		})

		listener.OnKVPair = func(updates KVPairData) error {
			return fmt.Errorf("error")
		}

		res := AsyncListener(listener, 16, commitChan, context.Background())

		callAllCallbacksOnces(t, res)

		err := <-commitChan
		if err == nil || err.Error() != "error" {
			t.Fatalf("expected error, got %v", err)
		}

		checkExpectedCallOrder(t, calls, []string{"InitializeModuleData", "StartBlock", "OnTx", "OnEvent"})
	})
}
