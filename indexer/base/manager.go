package indexerbase

type Manager struct {
	logger              Logger
	decoderResolver     DecoderResolver
	decoders            map[string]KVDecoder
	listeners           []Listener
	needLogicalDecoding bool
	listener            Listener
}

// ManagerOptions are the options for creating a new Manager.
type ManagerOptions struct {
	// DecoderResolver is the resolver for module decoders. It is required.
	DecoderResolver DecoderResolver

	// Listeners are the listeners that will be called when the manager receives events.
	Listeners []Listener

	// CatchUpSource is the source that will be used do initial indexing of modules with pre-existing
	// state. It is optional, but if it is not provided, indexing can only be starting when a node
	// is synced from genesis.
	CatchUpSource CatchUpSource

	// Logger is the logger that will be used by the manager. It is optional.
	Logger Logger
}

// NewManager creates a new Manager with the provided options.
func NewManager(opts ManagerOptions) (*Manager, error) {
	if opts.Logger != nil {
		opts.Logger.Info("Initializing indexer manager")
	}

	mgr := &Manager{
		logger:              opts.Logger,
		decoderResolver:     opts.DecoderResolver,
		decoders:            map[string]KVDecoder{},
		listeners:           opts.Listeners,
		needLogicalDecoding: false,
		listener:            Listener{},
	}

	return mgr, nil
}

// Listener returns that listener that should be passed directly to the blockchain for managing
// all indexing.
func (p *Manager) init() error {
	// check each subscribed listener to see if we actually need to register the listener

	for _, listener := range p.listeners {
		if listener.StartBlock != nil {
			p.listener.StartBlock = p.startBlock
			break
		}
	}

	for _, listener := range p.listeners {
		if listener.OnBlockHeader != nil {
			p.listener.OnBlockHeader = p.onBlockHeader
			break
		}
	}

	for _, listener := range p.listeners {
		if listener.OnTx != nil {
			p.listener.OnTx = p.onTx
			break
		}
	}

	for _, listener := range p.listeners {
		if listener.OnEvent != nil {
			p.listener.OnEvent = p.onEvent
			break
		}
	}

	for _, listener := range p.listeners {
		if listener.Commit != nil {
			p.listener.Commit = p.commit
			break
		}
	}

	for _, listener := range p.listeners {
		if listener.OnEntityUpdate != nil {
			p.needLogicalDecoding = true
			p.listener.OnKVPair = p.onKVPair
			break
		}
	}

	if p.listener.OnKVPair == nil {
		for _, listener := range p.listeners {
			if listener.OnKVPair != nil {
				p.listener.OnKVPair = p.onKVPair
				break
			}
		}
	}

	if p.needLogicalDecoding {
		err := p.decoderResolver.Iterate(func(moduleName string, module ModuleDecoder) error {
			p.decoders[moduleName] = module.KVDecoder
			// TODO
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Manager) startBlock(height uint64) error {
	if p.logger != nil {
		p.logger.Debug("start block", "height", height)
	}

	for _, listener := range p.listeners {
		if listener.StartBlock == nil {
			continue
		}
		if err := listener.StartBlock(height); err != nil {
			return err
		}
	}
	return nil
}

func (p *Manager) onBlockHeader(data BlockHeaderData) error {
	if p.logger != nil {
		p.logger.Debug("block header", "height", data.Height)
	}

	for _, listener := range p.listeners {
		if listener.OnBlockHeader == nil {
			continue
		}
		if err := listener.OnBlockHeader(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Manager) onTx(data TxData) error {
	for _, listener := range p.listeners {
		if listener.OnTx == nil {
			continue
		}
		if err := listener.OnTx(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Manager) onEvent(data EventData) error {
	for _, listener := range p.listeners {
		if err := listener.OnEvent(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Manager) commit() error {
	if p.logger != nil {
		p.logger.Debug("commit")
	}

	for _, listener := range p.listeners {
		if err := listener.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (p *Manager) onKVPair(storeKey string, key, value []byte, delete bool) error {
	if p.logger != nil {
		p.logger.Debug("kv pair", "storeKey", storeKey, "delete", delete)
	}

	for _, listener := range p.listeners {
		if listener.OnKVPair == nil {
			continue
		}
		if err := listener.OnKVPair(storeKey, key, value, delete); err != nil {
			return err
		}
	}

	if !p.needLogicalDecoding {
		return nil
	}

	decoder, ok := p.decoders[storeKey]
	if !ok {
		return nil
	}

	update, handled, err := decoder(key, value)
	if err != nil {
		return err
	}
	if !handled {
		p.logger.Info("not decoded", "storeKey", storeKey, "tableName", update.TableName)
		return nil
	}

	p.logger.Info("decoded",
		"storeKey", storeKey,
		"tableName", update.TableName,
		"key", update.Key,
		"values", update.Value,
		"delete", update.Delete,
	)

	for _, indexer := range p.listeners {
		if indexer.OnEntityUpdate == nil {
			continue
		}
		if err := indexer.OnEntityUpdate(storeKey, update); err != nil {
			return err
		}
	}

	return nil
}
