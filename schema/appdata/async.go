package appdata

import "context"

func AsyncListener(listener Listener, bufferSize int, commitChan chan<- error) Listener {
	packetChan := make(chan Packet, bufferSize)
	res := Listener{}

	res.Initialize = func(ctx context.Context, data InitializationData) (lastBlockPersisted int64, err error) {
		if listener.Initialize != nil {
			lastBlockPersisted, err = listener.Initialize(ctx, data)
			if err != nil {
				return
			}
		}

		cancel := ctx.Done()
		go func() {
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

				case <-cancel:
					return
				}
			}
		}()

		return
	}

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

	return res
}

func AsyncListenerMux(listeners []Listener, bufferSize int) Listener {
	asyncListeners := make([]Listener, len(listeners))
	commitChans := make([]chan error, len(listeners))
	for i, l := range listeners {
		commitChan := make(chan error)
		commitChans[i] = commitChan
		asyncListeners[i] = AsyncListener(l, bufferSize, commitChan)
	}
	mux := ListenerMux(asyncListeners...)
	muxCommit := mux.Commit
	mux.Commit = func(data CommitData) error {
		err := muxCommit(data)
		if err != nil {
			return err
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
