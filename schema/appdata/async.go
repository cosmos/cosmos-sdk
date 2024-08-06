package appdata

import (
	"context"
	"sync"
)

// AsyncListenerOptions are options for async listeners and listener mux's.
type AsyncListenerOptions struct {
	// Context is the context whose Done() channel listeners use will use to listen for completion to close their
	// goroutine. If it is nil, then context.Background() will be used and goroutines may be leaked.
	Context context.Context

	// BufferSize is the buffer size of the channels to use. It defaults to 0.
	BufferSize int

	// DoneWaitGroup is an optional wait-group that listener goroutines will notify via Add(1) when they are started
	// and Done() after they are canceled and completed.
	DoneWaitGroup *sync.WaitGroup
}

// AsyncListenerMux returns a listener that forwards received events to all the provided listeners asynchronously
// with each listener processing in a separate go routine. All callbacks in the returned listener will return nil
// except for Commit which will return an error or nil once all listeners have processed the commit. The context
// is used to signal that the listeners should stop listening and return. bufferSize is the size of the buffer for the
// channels used to send events to the listeners.
func AsyncListenerMux(opts AsyncListenerOptions, listeners ...Listener) Listener {
	asyncListeners := make([]Listener, len(listeners))
	commitChans := make([]chan error, len(listeners))
	for i, l := range listeners {
		commitChan := make(chan error)
		commitChans[i] = commitChan
		asyncListeners[i] = AsyncListener(opts, commitChan, l)
	}
	mux := ListenerMux(asyncListeners...)
	muxCommit := mux.Commit
	mux.Commit = func(data CommitData) error {
		if muxCommit != nil {
			err := muxCommit(data)
			if err != nil {
				return err
			}
		}

		for _, commitChan := range commitChans {
			err := <-commitChan
			if err != nil {
				return err
			}
		}
		return nil
	}

	return mux
}

// AsyncListener returns a listener that forwards received events to the provided listener listening in asynchronously
// in a separate go routine. The listener that is returned will return nil for all methods including Commit and
// an error or nil will only be returned in commitChan once the sender has sent commit and the receiving listener has
// processed it. Thus commitChan can be used as a synchronization and error checking mechanism. The go routine
// that is being used for listening will exit when context.Done() returns and no more events will be received by the listener.
// bufferSize is the size of the buffer for the channel that is used to send events to the listener.
// Instead of using AsyncListener directly, it is recommended to use AsyncListenerMux which does coordination directly
// via its Commit callback.
func AsyncListener(opts AsyncListenerOptions, commitChan chan<- error, listener Listener) Listener {
	packetChan := make(chan Packet, opts.BufferSize)
	res := Listener{}
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	done := ctx.Done()

	go func() {
		if opts.DoneWaitGroup != nil {
			opts.DoneWaitGroup.Add(1)
		}

		var err error
		for {
			select {
			case packet := <-packetChan:
				if err != nil {
					// if we have an error, don't process any more packets
					// and return the error and finish when it's time to commit
					if _, ok := packet.(CommitData); ok {
						commitChan <- err
						return
					}
				} else {
					// process the packet
					err = listener.SendPacket(packet)
					// if it's a commit
					if _, ok := packet.(CommitData); ok {
						commitChan <- err
						if err != nil {
							return
						}
					}
				}

			case <-done:
				close(packetChan)
				if opts.DoneWaitGroup != nil {
					opts.DoneWaitGroup.Done()
				}
				return
			}
		}
	}()

	if listener.InitializeModuleData != nil {
		res.InitializeModuleData = func(data ModuleInitializationData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.StartBlock != nil {
		res.StartBlock = func(data StartBlockData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnTx != nil {
		res.OnTx = func(data TxData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnEvent != nil {
		res.OnEvent = func(data EventData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnKVPair != nil {
		res.OnKVPair = func(data KVPairData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnObjectUpdate != nil {
		res.OnObjectUpdate = func(data ObjectUpdateData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.Commit != nil {
		res.Commit = func(data CommitData) error {
			packetChan <- data
			return nil
		}
	}

	return res
}
