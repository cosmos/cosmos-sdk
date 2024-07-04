package appdata

// ListenerMux returns a listener that forwards received events to all the provided listeners in order.
// A callback is only registered if a non-nil callback is present in at least one of the listeners.
func ListenerMux(listeners ...Listener) Listener {
	mux := Listener{}

	for _, l := range listeners {
		if l.InitializeModuleData != nil {
			mux.InitializeModuleData = func(data ModuleInitializationData) error {
				for _, l := range listeners {
					if l.InitializeModuleData != nil {
						if err := l.InitializeModuleData(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	for _, l := range listeners {
		if l.StartBlock != nil {
			mux.StartBlock = func(data StartBlockData) error {
				for _, l := range listeners {
					if l.StartBlock != nil {
						if err := l.StartBlock(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	for _, l := range listeners {
		if l.OnTx != nil {
			mux.OnTx = func(data TxData) error {
				for _, l := range listeners {
					if l.OnTx != nil {
						if err := l.OnTx(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnEvent != nil {
			mux.OnEvent = func(data EventData) error {
				for _, l := range listeners {
					if l.OnEvent != nil {
						if err := l.OnEvent(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnKVPair != nil {
			mux.OnKVPair = func(data KVPairData) error {
				for _, l := range listeners {
					if l.OnKVPair != nil {
						if err := l.OnKVPair(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	for _, listener := range listeners {
		if listener.OnObjectUpdate != nil {
			mux.OnObjectUpdate = func(data ObjectUpdateData) error {
				for _, l := range listeners {
					if l.OnObjectUpdate != nil {
						if err := l.OnObjectUpdate(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	for _, listener := range listeners {
		if listener.Commit != nil {
			mux.Commit = func(data CommitData) error {
				for _, l := range listeners {
					if l.Commit != nil {
						if err := l.Commit(data); err != nil {
							return err
						}
					}
				}
				return nil
			}
			break
		}
	}

	return mux
}
