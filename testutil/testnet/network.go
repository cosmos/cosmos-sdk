package testnet

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/p2p"
)

// NewNetwork concurrently calls createCometStarter, nVals times;
// then it returns a slice of started comet nodes,
// in order corresponding with the number passed to createCometStarter.
// The returned nodes will all be peered together,
// by dialing each node's [github.com/cometbft/cometbft/p2p/pex.Reactor] to each other.
//
// Every node is attempted to be started,
// and any errors collected are joined together and returned.
//
// In the event of errors, a non-nil Nodes slice may still be returned
// and some elements may be nil.
// Callers should call [Nodes.Stop] and [Nodes.Wait] to perform cleanup,
// regardless of the returned error.
func NewNetwork(nVals int, createCometStarter func(int) *CometStarter) (Nodes, error) {
	// The ordered slice of nodes; correct indexing is important.
	// The creator goroutines will write directly into this slice.
	nodes := make(Nodes, nVals)

	// Every node will be started in its own goroutine.
	// We collect the switches so that each node can dial every other node.
	switchCh := make(chan *p2p.Switch, nVals)
	errCh := make(chan error, nVals)

	var wg sync.WaitGroup
	// Start goroutines to populate nodes slice and notify as each node is available.
	for i := 0; i < nVals; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			n, err := createCometStarter(i).Start()
			if err != nil {
				errCh <- fmt.Errorf("failed to start node %d: %w", i, err)
				return
			}

			// Notify that the new node's switch is available, so this node can be
			// peered with the other nodes.
			switchCh <- n.Switch()

			// And assign the node into its correct index in the ordered slice.
			nodes[i] = n
		}(i)
	}

	// Once all the creation goroutines are complete, close the channels,
	// to signal to the collection goroutines.
	go func() {
		wg.Wait()
		close(errCh)
		close(switchCh)
	}()

	joinPeersDone := make(chan struct{})
	go joinPeers(switchCh, joinPeersDone)

	finalErrCh := make(chan error, 1)
	go collectErrors(errCh, finalErrCh)

	// If there were any errors, return them.
	// And return any set nodes, so that they can be cleaned up properly.
	if err := <-finalErrCh; err != nil {
		return nodes, err
	}

	// No errors, so wait for peer joining to complete
	// before returning the ordered nodes.
	<-joinPeersDone
	return nodes, nil
}

// collectErrors collects all errors that arrive on the in channel,
// joins them, then sends the joined final error on the out channel.
func collectErrors(in <-chan error, out chan<- error) {
	var errs []error
	for err := range in {
		errs = append(errs, err)
	}

	var res error
	if len(errs) > 0 {
		res = errors.Join(errs...)
	}
	out <- res
}

// joinPeers collects each switch arriving on newSwitches;
// each time a new switch arrives, it dials every previously seen switch.
//
// This allows each node to be started independently and concurrently
// without predetermined p2p ports.
func joinPeers(newSwitches <-chan *p2p.Switch, done chan<- struct{}) {
	defer close(done)

	var readySwitches []*p2p.Switch
	for newSwitch := range newSwitches {
		newNetAddr := newSwitch.NetAddress()
		for _, s := range readySwitches {
			// For every new switch, connect with all the previously seen switches.
			// It might not be necessary to dial in both directions, but it shouldn't hurt.
			_ = s.DialPeerWithAddress(newNetAddr)
			_ = newSwitch.DialPeerWithAddress(s.NetAddress())
		}
		readySwitches = append(readySwitches, newSwitch)
	}
}
