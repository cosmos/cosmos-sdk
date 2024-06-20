package listener

import (
	"context"

	"cosmossdk.io/schema"
)

// Multiplex returns a listener that multiplexes the given listeners and only
// registers a listener if a non-nil function is present in at least one of the listeners.
func Multiplex(listeners ...Listener) Listener {
	mux := Listener{}

	for _, l := range listeners {
		if l.Initialize != nil {
			mux.Initialize = func(ctx context.Context, data InitializationData) (lastBlockPersisted int64, err error) {
				for _, l := range listeners {
					if l.Initialize != nil {
						return l.Initialize(ctx, data)
					}
				}
				return 0, nil
			}
			break
		}
	}

	for _, l := range listeners {
		if l.InitializeModuleSchema != nil {
			mux.InitializeModuleSchema = func(moduleName string, moduleSchema schema.ModuleSchema) error {
				for _, l := range listeners {
					if l.InitializeModuleSchema != nil {
						if err := l.InitializeModuleSchema(moduleName, moduleSchema); err != nil {
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
		if l.CompleteInitialization != nil {
			mux.CompleteInitialization = func() error {
				for _, l := range listeners {
					if l.CompleteInitialization != nil {
						if err := l.CompleteInitialization(); err != nil {
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
			mux.StartBlock = func(u uint64) error {
				for _, l := range listeners {
					if l.StartBlock != nil {
						if err := l.StartBlock(u); err != nil {
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
		if l.OnBlockHeader != nil {
			mux.OnBlockHeader = func(data BlockHeaderData) error {
				for _, l := range listeners {
					if l.OnBlockHeader != nil {
						if err := l.OnBlockHeader(data); err != nil {
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
			mux.Commit = func() error {
				for _, l := range listeners {
					if l.Commit != nil {
						if err := l.Commit(); err != nil {
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
