package appdata

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestAsyncListenerMux(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		listener := AsyncListenerMux(AsyncListenerOptions{}, Listener{}, Listener{})

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
		wg := &sync.WaitGroup{}
		var calls1, calls2 []string
		listener1 := callCollector(1, func(name string, _ int, _ Packet) {
			calls1 = append(calls1, name)
		})
		listener2 := callCollector(2, func(name string, _ int, _ Packet) {
			calls2 = append(calls2, name)
		})
		res := AsyncListenerMux(AsyncListenerOptions{
			BufferSize: 16, Context: ctx, DoneWaitGroup: wg,
		}, listener1, listener2)

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

		// cancel and expect the test to finish - if all goroutines aren't canceled the test will hang
		cancel()
		wg.Wait()
	})

	t.Run("error on commit", func(t *testing.T) {
		var calls1, calls2 []string
		listener1 := callCollector(1, func(name string, _ int, _ Packet) {
			calls1 = append(calls1, name)
		})
		listener1.Commit = func(data CommitData) error {
			return fmt.Errorf("error")
		}
		listener2 := callCollector(2, func(name string, _ int, _ Packet) {
			calls2 = append(calls2, name)
		})
		res := AsyncListenerMux(AsyncListenerOptions{}, listener1, listener2)

		err := res.Commit(CommitData{})
		if err == nil || err.Error() != "error" {
			t.Fatalf("expected error, got %v", err)
		}
	})
}

func TestAsyncListener(t *testing.T) {
	t.Run("call cancel", func(t *testing.T) {
		commitChan := make(chan error)
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		var calls []string
		listener := callCollector(1, func(name string, _ int, _ Packet) {
			calls = append(calls, name)
		})
		res := AsyncListener(AsyncListenerOptions{BufferSize: 16, Context: ctx, DoneWaitGroup: wg},
			commitChan, listener)

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

		// expect wait group to return after cancel is called
		cancel()
		wg.Wait()
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

		res := AsyncListener(AsyncListenerOptions{BufferSize: 16}, commitChan, listener)

		callAllCallbacksOnces(t, res)

		err := <-commitChan
		if err == nil || err.Error() != "error" {
			t.Fatalf("expected error, got %v", err)
		}

		checkExpectedCallOrder(t, calls, []string{"InitializeModuleData", "StartBlock", "OnTx", "OnEvent"})
	})
}
